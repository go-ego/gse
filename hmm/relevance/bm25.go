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
	"strings"
	"unicode/utf8"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/consts"
	"github.com/go-ego/gse/hmm/segment"
	"github.com/go-ego/gse/hmm/stopwords"
	"github.com/go-ego/gse/types"
)

// BM25 Best Match
// ref: https://en.wikipedia.org/wiki/Okapi_BM25
type BM25 struct {
	// K1 Saturation Parameter
	// Controls the saturation of the TF,
	// i.e. a word frequency that exceeds the value of this parameter is not given more weight.
	// A lower k1 will make the word frequency less influential
	// and a higher k1 will make the word frequency more influential.
	// if not defined K1 , we will define it in 1.25
	K1 float64

	// B Length Normalization Parameter
	// Controls the degree of normalization of document length.
	// A lower b will make shorter documents more important
	// and a higher b will make longer documents more important.
	// so and K1 , if not defined by client, we will define it in 0.75
	B float64

	// AverageDocLength Average Document Length
	// Indicates the average vocabulary per document in the entire document set.
	// This value is used to normalize the document length in order to compare documents of different lengths.
	AverageDocLength float64

	// Base default setting
	Base
}

// AddToken add a new word with TFIDF into the dictionary.
func (bm25 *BM25) AddToken(text string, freq float64, pos ...string) error {
	err := bm25.Seg.AddToken(text, freq, pos...)
	return err
}

// LoadStopWord load stop word for TFIDF
func (bm25 *BM25) LoadStopWord(fileName ...string) error {
	return bm25.StopWord.LoadDict(fileName...)
}

// LoadDict load dict for TFIDF seg
func (bm25 *BM25) LoadDict(files ...string) error {
	if len(files) <= 0 {
		// bm25 needs tf and idf value , so we just get the tfidf path and loading it.
		files = bm25.Seg.GetTfIdfPath(files...)
	}

	dictFiles := make([]*types.LoadDictFile, len(files))
	for i, v := range files {
		dictFiles[i] = &types.LoadDictFile{
			FilePath: v,
			FileType: consts.LoadDictTypeBM25,
		}
	}

	return bm25.Seg.LoadTFIDFDict(dictFiles)
}

// LoadDictStr load dict for BM25 seg
func (bm25 *BM25) LoadDictStr(dictStr string) error {
	dictFile := &types.LoadDictFile{
		FilePath: dictStr,
		FileType: consts.LoadDictTypeBM25,
	}
	return bm25.Seg.LoadTFIDFDictStr(dictFile)
}

// Freq return the BM25 of the word
// BM25 need TF and IDF value, so we just use FindTFIDF func
func (bm25 *BM25) Freq(key string) (float64, interface{}, bool) {
	return bm25.Seg.FindTFIDF(key)
}

// NumTokens return the BM25 tokens' num
func (bm25 *BM25) NumTokens() int {
	return bm25.Seg.Dict.NumTokens()
}

// TotalFreq return the BM25 total frequency
func (bm25 *BM25) TotalFreq() float64 {
	return bm25.Seg.Dict.TotalFreq()
}

// FreqMap return the BM25 freq map
func (bm25 *BM25) FreqMap(text string) map[string]float64 {
	freqMap := make(map[string]float64)

	for _, w := range bm25.Seg.Cut(text, true) {
		w = strings.TrimSpace(w)
		if utf8.RuneCountInString(w) < 2 {
			continue
		}
		if bm25.StopWord.IsStopWord(w) {
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

// calculateK Calculate the K value for bm25
func (bm25 *BM25) calculateK(docNum float64) float64 {
	t := docNum / bm25.AverageDocLength
	return bm25.K1 * ((1 - bm25.B) + bm25.B*(t))
}

// calculateWeight calculate the word's weight by BM25
func (bm25 *BM25) calculateWeight(term string) float64 {
	tf, idf, _ := bm25.Freq(term)
	k := bm25.calculateK(float64(utf8.RuneCountInString(term)))

	return idf.(float64) * ((tf * (bm25.K1 + 1)) / (tf + k))
}

// ConstructSeg construct segment with weight
func (bm25 *BM25) ConstructSeg(text string) segment.Segments {
	// make segment list by total freq num
	ws := make([]segment.Segment, 0)
	for k := range bm25.FreqMap(text) {
		ws = append(ws, segment.Segment{Text: k, Weight: bm25.calculateWeight(k)})
	}

	return ws
}

// GetSeg get TFIDF Segmenter
func (bm25 *BM25) GetSeg() gse.Segmenter {
	return bm25.Seg
}

// LoadCorpus for calculate the average length of corpus
func (bm25 *BM25) LoadCorpus(path ...string) (err error) {
	averLength, err := bm25.Seg.LoadCorpusAverLen(path...)
	if err != nil {
		return
	}

	bm25.AverageDocLength = averLength
	return
}

// NewBM25 create a new BM25
func NewBM25(bm25Setting *types.BM25Setting) Relevance {
	// init value
	if bm25Setting == nil {
		bm25Setting = &types.BM25Setting{
			K1: consts.BM25DefaultK1,
			B:  consts.BM25DefaultB,
		}
	}
	if bm25Setting.K1 == 0 {
		bm25Setting.K1 = consts.BM25DefaultK1
	}
	if bm25Setting.B == 0 {
		bm25Setting.K1 = consts.BM25DefaultB
	}

	bm25 := &BM25{
		K1: bm25Setting.K1,
		B:  bm25Setting.B,
	}
	bm25.StopWord = stopwords.NewStopWord()

	return Relevance(bm25)
}
