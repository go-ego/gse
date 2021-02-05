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
