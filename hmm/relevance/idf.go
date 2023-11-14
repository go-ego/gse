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
	"github.com/go-ego/gse/hmm/segment"
	"github.com/go-ego/gse/hmm/stopwords"
)

// Idf type a dictionary for all words with the
// IDFs(Inverse Document Frequency).
type Idf struct {
	// median of word frequencies for calculate the weight of backup
	median float64

	// the list of word frequencies
	freqs []float64

	Base
}

// AddToken add a new word with IDF into the dictionary.
func (i *Idf) AddToken(text string, freq float64, pos ...string) error {
	err := i.Seg.AddToken(text, freq, pos...)

	i.freqs = append(i.freqs, freq)
	sort.Float64s(i.freqs)
	i.median = i.freqs[len(i.freqs)/2]
	return err
}

// LoadDict load the idf dictionary
func (i *Idf) LoadDict(files ...string) error {
	if len(files) <= 0 {
		files = i.Seg.GetIdfPath(files...)
	}

	return i.Seg.LoadDict(files...)
}

// LoadStopWord load stop word for IDF
func (i *Idf) LoadStopWord(fileName ...string) error {
	return i.StopWord.LoadDict(fileName...)
}

// LoadDictStr load dict for IDF seg
func (i *Idf) LoadDictStr(dictStr string) error {
	return i.Seg.LoadDictStr(dictStr)
}

// Freq return the IDF of the word
func (i *Idf) Freq(key string) (float64, interface{}, bool) {
	return i.Seg.Find(key)
}

// NumTokens return the IDF tokens' num
func (i *Idf) NumTokens() int {
	return i.Seg.Dict.NumTokens()
}

// TotalFreq return the IDF total frequency
func (i *Idf) TotalFreq() float64 {
	return i.Seg.Dict.TotalFreq()
}

// FreqMap return the IDF freq map
func (i *Idf) FreqMap(text string) map[string]float64 {
	freqMap := make(map[string]float64)

	for _, w := range i.Seg.Cut(text, true) {
		w = strings.TrimSpace(w)
		if utf8.RuneCountInString(w) < 2 {
			continue
		}
		if i.StopWord.IsStopWord(w) {
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

// calculateWeight calculate the word's weight by IDF
func (i *Idf) calculateWeight(k string, v float64) float64 {
	if freq, _, ok := i.Freq(k); ok {
		return freq * v
	}

	return i.median * v
}

// ConstructSeg construct segment with weight
func (i *Idf) ConstructSeg(text string) segment.Segments {
	// make segment list by total freq num
	ws := make([]segment.Segment, 0)

	for k, v := range i.FreqMap(text) {
		ws = append(ws, segment.Segment{Text: k, Weight: i.calculateWeight(k, v)})
	}

	return ws
}

// GetSeg get IDF Segmenter
func (i *Idf) GetSeg() gse.Segmenter {
	return i.Seg
}

// LoadCorpus idf no need to load corpus
func (i *Idf) LoadCorpus() error {
	return nil
}

// NewIdf create a new Idf
func NewIdf() Relevance {
	idf := &Idf{
		freqs: make([]float64, 0),
	}

	idf.StopWord = stopwords.NewStopWord()

	return Relevance(idf)
}
