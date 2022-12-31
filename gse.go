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

/*
Package gse Go efficient multilingual NLP and text segmentation,
*/
package gse

import (
	"regexp"

	"github.com/go-ego/gse/hmm"
)

const (
	// Version get the gse version
	Version = "v0.69.9.593, Green Lake!"

	// minTokenFrequency = 2 // only read tokens with frequency >= 2 from the dictionary
)

func init() {
	hmm.LoadModel()
}

// GetVersion get the gse version
func GetVersion() string {
	return Version
}

// Prob type hmm model struct
type Prob struct {
	B, E, M, S map[rune]float64
}

// New return new gse segmenter
func New(files ...string) (seg Segmenter, err error) {
	if len(files) > 1 && files[1] == "alpha" {
		seg.AlphaNum = true
	}

	err = seg.LoadDict(files...)
	return
}

// Cut cuts a str into words using accurate mode.
// Parameter hmm controls whether to use the HMM(Hidden Markov Model)
// or use the user's model.
//
// seg.Cut(text):
//
//	use the shortest path
//
// seg.Cut(text, false):
//
//	use cut dag not hmm
//
// seg.Cut(text, true):
//
//	use cut dag and hmm mode
func (seg *Segmenter) Cut(str string, hmm ...bool) []string {
	if len(hmm) <= 0 {
		return seg.Slice(str)
		// return seg.cutDAGNoHMM(str)
	}

	if len(hmm) > 0 && !hmm[0] {
		return seg.cutDAGNoHMM(str)
	}

	return seg.cutDAG(str)
}

// CutSearch cuts str into words using search engine mode.
func (seg *Segmenter) CutSearch(str string, hmm ...bool) []string {
	if len(hmm) <= 0 {
		return seg.Slice(str, true)
	}

	return seg.cutForSearch(str, hmm...)
}

// CutAll cuts a str into words using full mode.
func (seg *Segmenter) CutAll(str string) []string {
	return seg.cutAll(str)
}

// CutDAG cut string with DAG use hmm and regexp
func (seg *Segmenter) CutDAG(str string, reg ...*regexp.Regexp) []string {
	return seg.cutDAG(str, reg...)
}

// CutDAGNoHMM cut string with DAG not use hmm
func (seg *Segmenter) CutDAGNoHMM(str string) []string {
	return seg.cutDAGNoHMM(str)
}

// CutStr cut []string with Cut return string
func (seg *Segmenter) CutStr(str []string, separator ...string) (r string) {
	sep := " "
	if len(separator) > 0 {
		sep = separator[0]
	}

	for i := 0; i < len(str); i++ {
		if i == len(str)-1 {
			r += str[i]
		} else {
			r += str[i] + sep
		}
	}

	return
}

// LoadModel load the hmm model
//
// Use the user's model:
//
//	seg.LoadModel(B, E, M, S map[rune]float64)
func (seg *Segmenter) LoadModel(prob ...map[rune]float64) {
	hmm.LoadModel(prob...)
}

// HMMCut cut sentence string use HMM with Viterbi
func (seg *Segmenter) HMMCut(str string, reg ...*regexp.Regexp) []string {
	// hmm.LoadModel(prob...)
	return hmm.Cut(str, reg...)
}

// HMMCutMod cut sentence string use HMM with Viterbi
func (seg *Segmenter) HMMCutMod(str string, prob ...map[rune]float64) []string {
	hmm.LoadModel(prob...)
	return hmm.Cut(str)
}

// Slice use modeSegment segment retrun []string
// using search mode if searchMode is true
func (seg *Segmenter) Slice(s string, searchMode ...bool) []string {
	segs := seg.ModeSegment([]byte(s), searchMode...)
	return ToSlice(segs, searchMode...)
}

// Slice use modeSegment segment retrun string
// using search mode if searchMode is true
func (seg *Segmenter) String(s string, searchMode ...bool) string {
	segs := seg.ModeSegment([]byte(s), searchMode...)
	return ToString(segs, searchMode...)
}

// SegPos type a POS struct
type SegPos struct {
	Text, Pos string
}

// Pos return text and pos array
func (seg *Segmenter) Pos(s string, searchMode ...bool) []SegPos {
	sa := seg.ModeSegment([]byte(s), searchMode...)
	return ToPos(sa, searchMode...)
}

// PosStr cut []SegPos with Pos return string
func (seg *Segmenter) PosStr(str []SegPos, separator ...string) (r string) {
	sep := " "
	if len(separator) > 0 {
		sep = separator[0]
	}

	for i := 0; i < len(str); i++ {
		add := str[i].Text
		if !seg.SkipPos {
			add += "/" + str[i].Pos
		}

		if i == len(str)-1 {
			r += add
		} else {
			r += add + sep
		}
	}

	return
}
