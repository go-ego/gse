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

// AnalyzeToken analyze the segment info structure
type AnalyzeToken struct {
	// the start of the segment in the text
	Start int
	End   int

	Position int
	Len      int

	Type string

	Text string
	Freq float64
	Pos  string
}

// Segment a segment in the text
type Segment struct {
	// the start of the segment in the text
	start int

	// the bytes end of the segment in the text (not including this)
	end int

	Position int

	// segment information
	token *Token
}

// Start returns the start byte position of the segment
func (s *Segment) Start() int {
	return s.start
}

// End return the end byte position of the segment (not including this)
func (s *Segment) End() int {
	return s.end
}

// Token return the segment token information
func (s *Segment) Token() *Token {
	return s.token
}

// Text a string type，used to parse text
// 1. a word, such as "world" or "boundary", in English a word is a word
// 2. a participle, such as "world" a.k.a. "population"
// 3. a text, such as "the world has seven billion people"
type Text []byte

// Token define a segment token structure
type Token struct {
	// a segment string，it's []Text
	text []Text

	// a frequency of the token
	freq float64

	// part of speech label
	pos string

	// log2(total frequency/this segment frequency)，equal to log2(1/p(segment)))，
	// used by the short path as the path length of the clause in dynamic programming.
	// Solving for the maximum of prod(p(segment)) is equivalent to solving for the minimum of
	// the minimum of sum(distance(segment)),
	// which is where "shortest path" comes from.
	distance float32

	// the inverse document frequency of the token
	inverseFreq float64

	// For further segmentation of this segmented text,
	// see the Segments function comment.
	segments []*Segment
}

// Text return the text of the segment
func (token *Token) Text() string {
	return textSliceToString(token.text)
}

// Freq returns the frequency in the dictionary token
func (token *Token) Freq() float64 {
	return token.freq
}

// Pos returns the part of speech in the dictionary token
func (token *Token) Pos() string {
	return token.pos
}

// Segments will segment further subdivisions of the text of this participle,
// the participle has two subclauses.
//
// Subclauses can also have further subclauses forming a tree structure,
// which can be traversed to get all the detailed subdivisions of the participle,
// which is mainly Used by search engines to perform full-text searches on a piece of text.
func (token *Token) Segments() []*Segment {
	return token.segments
}

// Equals compare str split tokens
func (token *Token) Equals(str string) bool {
	tokenLen := 0
	for _, t := range token.text {
		tokenLen += len(t)
	}
	if tokenLen != len(str) {
		return false
	}

	bytStr := []byte(str)
	index := 0
	for i := 0; i < len(token.text); i++ {
		textArray := []byte(token.text[i])
		for j := 0; j < len(textArray); j++ {
			if textArray[j] != bytStr[index] {
				return false
			}

			index++
		}
	}

	return true
}
