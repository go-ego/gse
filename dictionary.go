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

// Dictionary struct implements a string double array trie.
// one segment maybe in leaf node or not
type Dictionary struct {
	trie *cedar.Cedar // Cedar double array trie

	maxTokenLen int     // the maximum length of the dictionary
	Tokens      []Token // the all tokens in the dictionary, to traverse
	totalFreq   float64 // the total number of tokens in the dictionary
}

// NewDict a new dictionary trie
func NewDict() *Dictionary {
	return &Dictionary{trie: cedar.New()}
}

// MaxTokenLen the maximum length of the dictionary
func (dict *Dictionary) MaxTokenLen() int {
	return dict.maxTokenLen
}

// NumTokens the number of tokens in the dictionary
func (dict *Dictionary) NumTokens() int {
	return len(dict.Tokens)
}

// TotalFreq the total frequency of the dictionary
func (dict *Dictionary) TotalFreq() float64 {
	return dict.totalFreq
}

// AddToken add a token to the dictionary
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

// LookupTokens finds tokens and words in the dictionary, matching the given pattern
// and returns the number of tokens
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
// and the word's frequency and pos
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

func (dict *Dictionary) FindTFIDF(word []byte) (float64, float64, bool) {
	var (
		id, value int
		freq      float64
		err       error
	)

	id, err = dict.trie.Jump(word, id)
	if err != nil {
		return 0, 0, false
	}

	value, err = dict.trie.Value(id)
	if err != nil && id != 0 {
		return 0, 0, true
	}

	if err != nil {
		return 0, 0, false
	}

	freq = dict.Tokens[value].freq
	inverseFreq := dict.Tokens[value].inverseFreq
	return freq, inverseFreq, true
}

// Value find word in the dictionary
// return the word's value and id
func (dict *Dictionary) Value(word []byte) (val, id int, err error) {
	id, err = dict.trie.Jump(word, id)
	if err != nil {
		return 0, id, err
	}

	val, err = dict.trie.Value(id)
	return
}
