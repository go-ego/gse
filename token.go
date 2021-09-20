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
	// 分词在文本中的起始位置
	Start int
	End   int

	Position int
	Len      int

	Type string

	Text string
	Freq float64
	Pos  string
}

// Segment 文本中的一个分词
type Segment struct {
	// 分词在文本中的起始字节位置
	start int

	// 分词在文本中的结束字节位置（不包括该位置）
	end int

	Position int

	// 分词信息
	token *Token
}

// Start 返回分词在文本中的起始字节位置
func (s *Segment) Start() int {
	return s.start
}

// End 返回分词在文本中的结束字节位置（不包括该位置）
func (s *Segment) End() int {
	return s.end
}

// Token 返回分词信息
func (s *Segment) Token() *Token {
	return s.token
}

// Text 字串类型，可以用来表达
//	1. 一个字元，比如 "世" 又如 "界", 英文的一个字元是一个词
//	2. 一个分词，比如 "世界" 又如 "人口"
//	3. 一段文字，比如 "世界有七十亿人口"
type Text []byte

// Token 一个分词
type Token struct {
	// 分词的字串，这实际上是个字元数组
	text []Text

	// 分词在语料库中的词频
	frequency float64

	// log2(总词频/该分词词频)，这相当于 log2(1/p(分词))，用作动态规划中
	// 该分词的路径长度。求解 prod(p(分词)) 的最大值相当于求解
	// sum(distance(分词)) 的最小值，这就是“最短路径”的来历。
	distance float32

	// 词性标注
	pos string

	// 该分词文本的进一步分词划分，见 Segments 函数注释。
	segments []*Segment
}

// Text 返回分词文本
func (token *Token) Text() string {
	return textSliceToString(token.text)
}

// Frequency 返回分词在语料库中的词频
func (token *Token) Frequency() float64 {
	return token.frequency
}

// Pos 返回分词词性标注
func (token *Token) Pos() string {
	return token.pos
}

// Segments 该分词文本的进一步分词划分，比如 "山达尔星联邦共和国联邦政府" 这个分词
// 有两个子分词 "山达尔星联邦共和国 " 和 "联邦政府"。子分词也可以进一步有子分词
// 形成一个树结构，遍历这个树就可以得到该分词的所有细致分词划分，这主要
// 用于搜索引擎对一段文本进行全文搜索。
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
