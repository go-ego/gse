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
	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/segment"
	"github.com/go-ego/gse/hmm/stopwords"
)

// Relevance easily scalable Relevance calculations (for idf, tf-idf, bm25 and so on)
type Relevance interface {
	// AddToken add text, frequency, position on obj
	AddToken(text string, freq float64, pos ...string) error

	// LoadDict load file from incoming parameters,
	// if incoming params no exist, will load file from default file path
	LoadDict(files ...string) error

	// LoadDictStr loading dict file by file path
	LoadDictStr(pathStr string) error

	// LoadCorpus loading corpus
	LoadCorpus(path ...string) error

	// LoadStopWord loading word file by filename
	LoadStopWord(fileName ...string) error

	// Freq find the frequency, position, existence information of the key
	Freq(key string) (float64, interface{}, bool)

	// NumTokens  the number of tokens in the dictionary
	NumTokens() int

	// TotalFreq the total number of tokens in the dictionary
	TotalFreq() float64

	// FreqMap get frequency map
	// key: word, value: frequency
	FreqMap(text string) map[string]float64

	// GetSeg Get the segmenter of Relevance algorithms
	GetSeg() gse.Segmenter

	// ConstructSeg return the segment with weight
	ConstructSeg(text string) segment.Segments
}

type Base struct {
	// loading some stop words
	StopWord *stopwords.StopWord

	// loading segmenter for cut word
	Seg gse.Segmenter
}
