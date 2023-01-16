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
	"bytes"
	"fmt"
)

// ToString converts a segments slice to string retrun the string
//
//	 two output modes:
//
//		normal mode (searchMode=false）
//		search mode（searchMode=true）
//
// default searchMode=false
// search mode is used search engine, and will output more results
func ToString(segs []Segment, searchMode ...bool) (output string) {
	var mode bool
	if len(searchMode) > 0 {
		mode = searchMode[0]
	}

	if mode {
		for _, seg := range segs {
			output += tokenToString(seg.token)
		}
		return
	}

	for _, seg := range segs {
		output += fmt.Sprintf("%s/%s ",
			textSliceToString(seg.token.text), seg.token.pos)
	}
	return
}

func tokenToString(token *Token) (output string) {
	hasOnlyTerminalToken := true
	for _, s := range token.segments {
		if len(s.token.segments) > 1 || IsJp(string(s.token.text[0])) {
			hasOnlyTerminalToken = false
		}

		if !hasOnlyTerminalToken && s != nil {
			output += tokenToString(s.token)
		}
	}

	output += fmt.Sprintf("%s/%s ", textSliceToString(token.text), token.pos)
	return
}

func tokenToBytes(token *Token) (output []byte) {
	for _, s := range token.segments {
		output = append(output, tokenToBytes(s.token)...)
	}
	output = append(output,
		[]byte(fmt.Sprintf("%s/%s ", textSliceToString(token.text), token.pos))...)

	return
}

// ToSlice converts a segments to slice retrun string slice
func ToSlice(segs []Segment, searchMode ...bool) (output []string) {
	var mode bool
	if len(searchMode) > 0 {
		mode = searchMode[0]
	}

	if mode {
		for _, seg := range segs {
			output = append(output, tokenToSlice(seg.token)...)
		}
		return
	}

	for _, seg := range segs {
		output = append(output, seg.token.Text())
	}
	return
}

func tokenToSlice(token *Token) (output []string) {
	hasOnlyTerminalToken := true
	for _, s := range token.segments {
		if len(s.token.segments) > 1 || IsJp(string(s.token.text[0])) {
			hasOnlyTerminalToken = false
		}

		if !hasOnlyTerminalToken {
			output = append(output, tokenToSlice(s.token)...)
		}
	}

	output = append(output, textSliceToString(token.text))
	return
}

// ToPos converts a segments slice to []SegPos
func ToPos(segs []Segment, searchMode ...bool) (output []SegPos) {
	var mode bool
	if len(searchMode) > 0 {
		mode = searchMode[0]
	}

	if mode {
		for _, seg := range segs {
			output = append(output, tokenToPos(seg.token)...)
		}
		return
	}

	for _, seg := range segs {
		pos1 := SegPos{
			Text: textSliceToString(seg.token.text),
			Pos:  seg.token.pos,
		}

		output = append(output, pos1)
	}

	return
}

func tokenToPos(token *Token) (output []SegPos) {
	hasOnlyTerminalToken := true
	for _, s := range token.segments {
		if len(s.token.segments) > 1 || IsJp(string(s.token.text[0])) {
			hasOnlyTerminalToken = false
		}

		if !hasOnlyTerminalToken {
			output = append(output, tokenToPos(s.token)...)
		}
	}

	pos1 := SegPos{
		Text: textSliceToString(token.text),
		Pos:  token.pos,
	}
	output = append(output, pos1)

	return
}

// let make multiple []Text into one string ooutput
func textToString(text []Text) (output string) {
	for _, word := range text {
		output += string(word)
	}
	return
}

// let make []Text toString returns a string output
func textSliceToString(text []Text) string {
	return Join(text)
}

// retrun total length of text slice
func textSliceByteLen(text []Text) (length int) {
	for _, word := range text {
		length += len(word)
	}
	return
}

func textSliceToBytes(text []Text) []byte {
	var buf bytes.Buffer
	for _, word := range text {
		buf.Write(word)
	}

	return buf.Bytes()
}

// Join is better string splicing
func Join(text []Text) string {
	switch len(text) {
	case 0:
		return ""
	case 1:
		return string(text[0])
	case 2:
		// Special case for common small values.
		// Remove if github.com/golang/go/issues/6714 is fixed
		return string(text[0]) + string(text[1])
	case 3:
		// Special case for common small values.
		// Remove if #6714 is fixed
		return string(text[0]) + string(text[1]) + string(text[2])
	}

	n := 0
	for i := 0; i < len(text); i++ {
		n += len(text[i])
	}

	b := make([]byte, n)
	bp := copy(b, text[0])
	for _, str := range text[1:] {
		bp += copy(b[bp:], str)
	}
	return string(b)
}

func printTokens(tokens []*Token, numTokens int) (output string) {
	for iToken := 0; iToken < numTokens; iToken++ {
		for _, word := range tokens[iToken].text {
			output += fmt.Sprint(string(word))
		}
		output += " "
	}
	return
}

func toWords(strings ...string) []Text {
	words := []Text{}
	for _, s := range strings {
		words = append(words, []byte(s))
	}
	return words
}

func bytesToString(bytes []Text) (output string) {
	for _, b := range bytes {
		output += (string(b) + "/")
	}
	return
}
