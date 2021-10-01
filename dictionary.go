// Copyright 2013 Hui Chen
// Copyright 2016 ego authors
//
// Copyright 2016 The go-ego Project Developers. See the COPYRIGHT
// file at the top-level directory of this distribution and at
// https://github.com/go-ego/gse/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

package gse

import (
	"github.com/vcaesar/cedar"
)

// Dictionary 结构体实现了一个字串双数组树，
// 一个分词可能出现在叶子节点也有可能出现在非叶节点
type Dictionary struct {
	trie *cedar.Cedar // Cedar 双数组树

	maxTokenLen int     // 词典中最长的分词
	Tokens      []Token // 词典中所有的分词，方便遍历
	totalFreq   float64 // 词典中所有分词的频率之和
}

// NewDict new dictionary
func NewDict() *Dictionary {
	return &Dictionary{trie: cedar.New()}
}

// MaxTokenLen 词典中最长的分词
func (dict *Dictionary) MaxTokenLen() int {
	return dict.maxTokenLen
}

// NumTokens 词典中分词数目
func (dict *Dictionary) NumTokens() int {
	return len(dict.Tokens)
}

// TotalFreq 词典中所有分词的频率之和
func (dict *Dictionary) TotalFreq() float64 {
	return dict.totalFreq
}

// AddToken 向词典中加入一个分词
func (dict *Dictionary) AddToken(token Token) error {
	bytes := textSliceToBytes(token.text)
	val, err := dict.trie.Get(bytes)
	if err == nil || val > 0 {
		return nil
	}

	err = dict.trie.Insert(bytes, dict.NumTokens())
	if err != nil {
		return err
	}

	dict.Tokens = append(dict.Tokens, token)
	dict.totalFreq += token.freq

	if len(token.text) > dict.maxTokenLen {
		dict.maxTokenLen = len(token.text)
	}

	return nil
}

// RemoveToken remove token in dictionary
func (dict *Dictionary) RemoveToken(token Token) error {
	bytes := textSliceToBytes(token.text)

	return dict.trie.Delete(bytes)
}

// LookupTokens 在词典中查找和字元组 words 可以前缀匹配的所有分词
// 返回值为找到的分词数
func (dict *Dictionary) LookupTokens(
	words []Text, tokens []*Token) (numOfTokens int) {
	var (
		id, value int
		err       error
	)

	for _, word := range words {
		id, err = dict.trie.Jump(word, id)
		if err != nil {
			break
		}

		value, err = dict.trie.Value(id)
		if err == nil {
			tokens[numOfTokens] = &dict.Tokens[value]
			numOfTokens++
		}
	}

	return
}

// Find find the word in the dictionary is non-existent
// and the word's frequency, pos
func (dict *Dictionary) Find(word []byte) (float64, string, bool) {
	var (
		id, value int
		freq      float64
		err       error
	)

	id, err = dict.trie.Jump(word, id)
	if err != nil {
		return 0, "", false
	}

	value, err = dict.trie.Value(id)
	if err != nil && id != 0 {
		return 0, "", true
	}

	if err != nil {
		return 0, "", false
	}

	freq = dict.Tokens[value].freq
	pos := dict.Tokens[value].pos
	return freq, pos, true
}

// Value find word in the dictionary
// retrun the word's value, id
func (dict *Dictionary) Value(word []byte) (val, id int, err error) {
	id, err = dict.trie.Jump(word, id)
	if err != nil {
		return 0, id, err
	}

	val, err = dict.trie.Value(id)
	return
}
