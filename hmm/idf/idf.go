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

// Idf represents a dictionary for all words with their
// IDFs(Inverse Document Frequency).
type Idf struct {
	median float64
	freqs  []float64

	seg gse.Segmenter
}

// AddToken adds a new word with IDF into it's dictionary.
func (i *Idf) AddToken(text string, frequency float64, pos ...string) {
	i.seg.AddToken(text, frequency, pos...)

	i.freqs = append(i.freqs, frequency)
	sort.Float64s(i.freqs)
	i.median = i.freqs[len(i.freqs)/2]
}

// LoadDict load idf dictionary
func (i *Idf) LoadDict(files ...string) error {
	if len(files) <= 0 {
		files = gse.GetIdfPath(files...)
	}

	return i.seg.LoadDict(files...)
}

// Frequency returns the IDF of given word.
func (i *Idf) Frequency(key string) (float64, bool) {
	return i.seg.Find(key)
}

// NewIdf creates a new Idf instance.
func NewIdf() *Idf {
	return &Idf{freqs: make([]float64, 0)}
}
