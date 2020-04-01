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

// Segmenter 分词器结构体
type Segmenter struct {
	dict *Dictionary
}

// jumper 该结构体用于记录 Viterbi 算法中某字元处的向前分词跳转信息
type jumper struct {
	minDistance float32
	token       *Token
}

// Segment 对文本分词
//
// 输入参数：
//	bytes	UTF8 文本的字节数组
//
// 输出：
//	[]Segment	划分的分词
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
	// 处理特殊情况
	if len(bytes) == 0 {
		// return []Segment{}
		return nil
	}

	// 划分字元
	text := SplitTextToWords(bytes)

	return seg.segmentWords(text, searchMode)
}

func (seg *Segmenter) segmentWords(text []Text, searchMode bool) []Segment {
	// 搜索模式下该分词已无继续划分可能的情况
	if searchMode && len(text) == 1 {
		return nil
	}

	// jumpers 定义了每个字元处的向前跳转信息，
	// 包括这个跳转对应的分词，
	// 以及从文本段开始到该字元的最短路径值
	jumpers := make([]jumper, len(text))

	if seg.dict == nil {
		return nil
	}

	tokens := make([]*Token, seg.dict.maxTokenLen)
	for current := 0; current < len(text); current++ {
		// 找到前一个字元处的最短路径，以便计算后续路径值
		var baseDistance float32
		if current == 0 {
			// 当本字元在文本首部时，基础距离应该是零
			baseDistance = 0
		} else {
			baseDistance = jumpers[current-1].minDistance
		}

		// 寻找所有以当前字元开头的分词
		tx := text[current:minInt(current+seg.dict.maxTokenLen, len(text))]
		numTokens := seg.dict.LookupTokens(tx, tokens)

		// 对所有可能的分词，更新分词结束字元处的跳转信息
		for iToken := 0; iToken < numTokens; iToken++ {
			location := current + len(tokens[iToken].text) - 1
			if !searchMode || current != 0 || location != len(text)-1 {
				updateJumper(&jumpers[location], baseDistance, tokens[iToken])
			}
		}

		// 当前字元没有对应分词时补加一个伪分词
		if numTokens == 0 || len(tokens[0].text) > 1 {
			updateJumper(&jumpers[current], baseDistance,
				&Token{text: []Text{text[current]}, frequency: 1, distance: 32, pos: "x"})
		}
	}

	// 从后向前扫描第一遍得到需要添加的分词数目
	numSeg := 0
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg++
		index = location - 1
	}

	// 从后向前扫描第二遍添加分词到最终结果
	outputSegments := make([]Segment, numSeg)
	for index := len(text) - 1; index >= 0; {
		location := index - len(jumpers[index].token.text) + 1
		numSeg--
		outputSegments[numSeg].token = jumpers[index].token
		index = location - 1
	}

	// 计算各个分词的字节位置
	bytePosition := 0
	for iSeg := 0; iSeg < len(outputSegments); iSeg++ {
		outputSegments[iSeg].start = bytePosition
		bytePosition += textSliceByteLen(outputSegments[iSeg].token.text)
		outputSegments[iSeg].end = bytePosition
	}

	return outputSegments
}

// updateJumper 更新跳转信息:
// 	1. 当该位置从未被访问过时 (jumper.minDistance 为零的情况)，或者
//	2. 当该位置的当前最短路径大于新的最短路径时
// 将当前位置的最短路径值更新为 baseDistance 加上新分词的概率
func updateJumper(jumper *jumper, baseDistance float32, token *Token) {
	newDistance := baseDistance + token.distance
	if jumper.minDistance == 0 || jumper.minDistance > newDistance {
		jumper.minDistance = newDistance
		jumper.token = token
	}
}

// SplitTextToWords 将文本划分成字元
func SplitTextToWords(text Text) []Text {
	output := make([]Text, 0, len(text)/3)
	current := 0
	inAlphanumeric := true
	alphanumericStart := 0

	for current < len(text) {
		r, size := utf8.DecodeRune(text[current:])
		if size <= 2 && (unicode.IsLetter(r) || unicode.IsNumber(r)) {
			// 当前是拉丁字母或数字（非中日韩文字）
			if !inAlphanumeric {
				alphanumericStart = current
				inAlphanumeric = true
			}

			if AlphaNum {
				output = append(output, toLow(text[current:current+size]))
			}
		} else {
			if inAlphanumeric {
				inAlphanumeric = false
				if current != 0 && !AlphaNum {
					output = append(output, toLow(text[alphanumericStart:current]))
				}
			}

			output = append(output, text[current:current+size])
		}
		current += size
	}

	// 处理最后一个字元是英文的情况
	if inAlphanumeric && !AlphaNum {
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

// toLower 将英文词转化为小写
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

// minInt 取两整数较小值
func minInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// maxInt 取两整数较大值
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
