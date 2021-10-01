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

package idf

import (
	"sort"

	"github.com/go-ego/gse"
)

// Idf type a dictionary for all words with the
// IDFs(Inverse Document Frequency).
type Idf struct {
	median float64
	freqs  []float64

	seg gse.Segmenter
}

// AddToken add a new word with IDF into the dictionary.
func (i *Idf) AddToken(text string, freq float64, pos ...string) error {
	err := i.seg.AddToken(text, freq, pos...)

	i.freqs = append(i.freqs, freq)
	sort.Float64s(i.freqs)
	i.median = i.freqs[len(i.freqs)/2]
	return err
}

// LoadDict load the idf dictionary
func (i *Idf) LoadDict(files ...string) error {
	if len(files) <= 0 {
		files = gse.GetIdfPath(files...)
	}

	return i.seg.LoadDict(files...)
}

// Freq return the IDF of the word
func (i *Idf) Freq(key string) (float64, string, bool) {
	return i.seg.Find(key)
}

// NumTokens return the IDF tokens' num
func (i *Idf) NumTokens() int {
	return i.seg.Dict.NumTokens()
}

// TotalFreq reruen the IDF total frequency
func (i *Idf) TotalFreq() float64 {
	return i.seg.Dict.TotalFreq()
}

// NewIdf create a new Idf
func NewIdf() *Idf {
	return &Idf{freqs: make([]float64, 0)}
}
