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

package relevance

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/consts"
	"github.com/go-ego/gse/hmm/segment"
	"github.com/go-ego/gse/hmm/stopwords"
	"github.com/go-ego/gse/types"
)

// TFIDF a measure of importance of a word to a document in a collection.
// Term Frequency-Inverse Document Frequency
// ref:https://en.wikipedia.org/wiki/Tf–idf
type TFIDF struct {
	// the list of word frequencies
	freqs []float64

	Base
}

// AddToken add a new word with TFIDF into the dictionary.
func (t *TFIDF) AddToken(text string, freq float64, pos ...string) error {
	err := t.Seg.AddToken(text, freq, pos...)

	t.freqs = append(t.freqs, freq)
	sort.Float64s(t.freqs)
	return err
}

// LoadStopWord load stop word for TFIDF
func (t *TFIDF) LoadStopWord(fileName ...string) error {
	return t.StopWord.LoadDict(fileName...)
}

// LoadDict load dict for TFIDF seg
func (t *TFIDF) LoadDict(files ...string) error {
	if len(files) <= 0 {
		files = t.Seg.GetTfIdfPath(files...)
	}
	dictFiles := make([]*types.LoadDictFile, len(files))
	for i, v := range files {
		dictFiles[i] = &types.LoadDictFile{
			File:     v,
			FileType: consts.LoadDictTypeTFIDF,
		}
	}

	return t.Seg.LoadTFIDFDict(dictFiles)
}

// LoadDictStr load dict for TFIDF seg
func (t *TFIDF) LoadDictStr(dictStr string) error {
	dictFile := &types.LoadDictFile{
		File:     dictStr,
		FileType: consts.LoadDictTypeTFIDF,
	}
	return t.Seg.LoadTFIDFDictStr(dictFile)
}

// Freq return the TFIDF of the word
func (t *TFIDF) Freq(key string) (float64, interface{}, bool) {
	return t.Seg.FindTFIDF(key)
}

// NumTokens return the TFIDF tokens' num
func (t *TFIDF) NumTokens() int {
	return t.Seg.Dict.NumTokens()
}

// TotalFreq return the TFIDF total frequency
func (t *TFIDF) TotalFreq() float64 {
	return t.Seg.Dict.TotalFreq()
}

// FreqMap return the TFIDF freq map
func (t *TFIDF) FreqMap(text string) map[string]float64 {
	freqMap := make(map[string]float64)

	for _, w := range t.Seg.Cut(text, true) {
		w = strings.TrimSpace(w)
		if utf8.RuneCountInString(w) < 2 {
			continue
		}
		if t.StopWord.IsStopWord(w) {
			continue
		}

		if f, ok := freqMap[w]; ok {
			freqMap[w] = f + 1.0
		} else {
			freqMap[w] = 1.0
		}
	}

	total := 0.0
	for _, freq := range freqMap {
		total += freq
	}

	for k, v := range freqMap {
		freqMap[k] = v / total
	}

	return freqMap
}

// calculateIdf calculate the word's weight by TFIDF
func (t *TFIDF) calculateWeight(term string) float64 {
	tf, idf, _ := t.Freq(term)
	return tf * idf.(float64)
}

// ConstructSeg construct segment with weight
func (t *TFIDF) ConstructSeg(text string) segment.Segments {
	// make segment list by total freq num
	ws := make([]segment.Segment, 0)
	for k := range t.FreqMap(text) {
		ws = append(ws, segment.Segment{Text: k, Weight: t.calculateWeight(k)})
	}

	return ws
}

// GetSeg get TFIDF Segmenter
func (t *TFIDF) GetSeg() gse.Segmenter {
	return t.Seg
}

// NewTFIDF create a new TFIDF
func NewTFIDF() Relevance {
	tfidf := &TFIDF{
		freqs: make([]float64, 0),
	}

	tfidf.StopWord = stopwords.NewStopWord()

	return Relevance(tfidf)
}
