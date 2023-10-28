// Copyright 2013 Hui Chen
// Copyright 2016 ego authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gse

import (
	"unicode"
	"unicode/utf8"
)

// Segmenter define the segmenter structure
type Segmenter struct {
	Dict     *Dictionary
	Load     bool
	DictSep  string
	DictPath string

	// NotLoadHMM option load the default hmm model config (Chinese char)
	NotLoadHMM bool

	// AlphaNum set splitTextToWords can add token
	// when words in alphanum
	// set up alphanum dictionary word segmentation
	AlphaNum bool
	Alpha    bool
	Num      bool
	// ToLower set alpha tolower
	// ToLower bool

	// LoadNoFreq load not have freq dict word
	LoadNoFreq bool
	// MinTokenFreq load min freq token
	MinTokenFreq float64
	// TextFreq add token frequency when not specified freq
	TextFreq string

	// SkipLog set skip log print
	SkipLog bool
	MoreLog bool

	// SkipPos skip PosStr pos
	SkipPos bool

	NotStop bool
	// StopWordMap the stop word map
	StopWordMap map[string]bool
}

// jumper this structure is used to record information
// about the forward leap at a word in the Viterbi algorithm
type jumper struct {
	minDistance float32
	token       *Token
}

// Segment use the shortest path to segment the text
//
// input parameter：
//
// bytes	UTF8 text []byte
//
// output：
//
// []Segment return segments result
func (seg *Segmenter) Segment(bytes []byte) []Segment {
	return seg.internalSegment(bytes, false)
}

// ModeSegment segment using search mode if searchMode is true
func (seg *Segmenter) ModeSegment(bytes []byte, searchMode ...bool) []Segment {
	var mode bool
	if len(searchMode) > 0 {
		mode = searchMode[0]
	}

	return seg.internalSegment(bytes, mode)
}

func (seg *Segmenter) internalSegment(bytes []byte, searchMode bool) []Segment {
	// special cases
	if len(bytes) == 0 {
		// return []Segment{}
		return nil
	}

	// split text to words
	text := seg.SplitTextToWords(bytes)

	return seg.segmentWords(text, searchMode)
}

func (seg *Segmenter) segmentWords(text []Text, searchMode bool) []Segment {
	// The case where the division is no longer possible in the search mode
	if searchMode && len(text) == 1 {
		return nil
	}

	// jumpers defines the forward jump information at each literal,
	// including the subword corresponding to this jump,
	// the and the value of the shortest path from the start
	// of the text segment to that literal
	//
	jumpers := make([]jumper, len(text))

	if seg.Dict == nil {
		return nil
	}

	tokens := make([]*Token, seg.Dict.maxTokenLen)
	for current := 0; current < len(text); current++ {
		// find the shortest path of the previous token,
		// to calculate the subsequent path values
		var baseDistance float32
		if current == 0 {
			// When this character is at the beginning of the text,
			// the base distance should be zero
			baseDistance = 0
		} else {
			baseDistance = jumpers[current-1].minDistance
		}

		// find all the segments starting with this token
		tx := text[current:minInt(current+seg.Dict.maxTokenLen, len(text))]
		numTokens := seg.Dict.LookupTokens(tx, tokens)

		// Update the jump information at the end of the split word
		// for all possible splits
		for iToken := 0; iToken < numTokens; iToken++ {
			location := current + len(tokens[iToken].text) - 1
			if !searchMode || current != 0 || location != len(text)-1 {
				updateJumper(&jumpers[location], baseDistance, tokens[iToken])
			}
		}

		// Add a pseudo-syllable if there is no corresponding syllable
		// for the current character
		if numTokens == 0 || len(tokens[0].text) > 1 {
			updateJumper(&jumpers[current], baseDistance,
				&Token{text: []Text{text[current]}, freq: 1, distance: 32, pos: "x"})
		}
	}

	// Scan the first pass from back to front
	// to get the number of subwords to be added
	numSeg := 0
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg++
		index = location - 1
	}

	// Scan from back to front for a second time
	// to add the split to the final result
	outputSegments := make([]Segment, numSeg)
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg--
		outputSegments[numSeg].token = jumpers[index].token
		index = location - 1
	}

	// Calculate the byte position of each participle
	bytePosition := 0
	for iSeg := 0; iSeg < len(outputSegments); iSeg++ {
		outputSegments[iSeg].start = bytePosition
		bytePosition += textSliceByteLen(outputSegments[iSeg].token.text)
		outputSegments[iSeg].end = bytePosition
	}

	return outputSegments
}

// updateJumper Update the jump information:
//  1. When the location has never been visited
//     (the case where jumper.minDistance is zero), or
//  2. When the current shortest path at the location
//     is greater than the new shortest path
//
// Update the shortest path value of the current location to baseDistance
// add the probability of the new split
func updateJumper(jumper *jumper, baseDistance float32, token *Token) {
	newDistance := baseDistance + token.distance
	if jumper.minDistance == 0 || jumper.minDistance > newDistance {
		jumper.minDistance = newDistance
		jumper.token = token
	}
}

// SplitWords splits a string to token words
func SplitWords(text Text) []Text {
	var seg Segmenter
	return seg.SplitTextToWords(text)
}

// SplitTextToWords splits a string to token words
func (seg *Segmenter) SplitTextToWords(text Text) []Text {
	output := make([]Text, 0, len(text)/3)
	current, alphanumericStart := 0, 0
	inAlphanumeric := true

	for current < len(text) {
		r, size := utf8.DecodeRune(text[current:])
		isNum := unicode.IsNumber(r) && !seg.Num
		isAlpha := unicode.IsLetter(r) && !seg.Alpha
		if size <= 2 && (isAlpha || isNum) {
			// Currently is Latin alphabet or numbers (not in CJK)
			if !inAlphanumeric {
				alphanumericStart = current
				inAlphanumeric = true
			}

			if seg.AlphaNum {
				output = append(output, toLow(text[current:current+size]))
			}
		} else {
			if inAlphanumeric {
				inAlphanumeric = false
				if current != 0 && !seg.AlphaNum {
					output = append(output, toLow(text[alphanumericStart:current]))
				}
			}

			output = append(output, text[current:current+size])
		}
		current += size
	}

	// process last byte is alpha and num
	if inAlphanumeric && !seg.AlphaNum {
		if current != 0 {
			output = append(output, toLow(text[alphanumericStart:current]))
		}
	}

	return output
}

func toLow(text []byte) []byte {
	if ToLower {
		return toLower(text)
	}

	return text
}

// toLower converts a string to lower
func toLower(text []byte) []byte {
	output := make([]byte, len(text))
	for i, t := range text {
		if t >= 'A' && t <= 'Z' {
			output[i] = t - 'A' + 'a'
		} else {
			output[i] = t
		}
	}

	return output
}

// minInt get min value of int
func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// maxInt get max value of int
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
