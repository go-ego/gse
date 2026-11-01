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
	"fmt"
	"testing"

	"github.com/vcaesar/tt"

	"github.com/go-ego/gse/types"
)

const (
	idfDictStr   = "hello 10\nworld 5\ngolang 3"
	tfidfDictStr = "hello 10 2\nworld 5 3\ngolang 3 4"
	freqText     = "hello world hello the golang gopher"
)

func newTestIdf(t *testing.T) *Idf {
	idf := NewIdf().(*Idf)
	idf.Seg.SkipLog = true
	tt.Nil(t, idf.LoadDictStr(idfDictStr))
	return idf
}

func newTestTFIDF(t *testing.T) *TFIDF {
	tfidf := NewTFIDF().(*TFIDF)
	tfidf.Seg.SkipLog = true
	tt.Nil(t, tfidf.LoadDictStr(tfidfDictStr))
	return tfidf
}

func newTestBM25(t *testing.T) *BM25 {
	bm25 := NewBM25(nil).(*BM25)
	bm25.Seg.SkipLog = true
	tt.Nil(t, bm25.LoadDictStr(tfidfDictStr))
	return bm25
}

func TestIdfDictStr(t *testing.T) {
	idf := newTestIdf(t)
	tt.Equal(t, 3, idf.NumTokens())
	tt.Equal(t, 18.0, idf.TotalFreq())

	freq, pos, ok := idf.Freq("hello")
	tt.Bool(t, ok)
	tt.Equal(t, 10.0, freq)
	tt.Equal(t, "", pos)

	_, _, ok = idf.Freq("gopher")
	tt.Bool(t, !ok)

	seg := idf.GetSeg()
	tt.Equal(t, idf.Seg.Dict, seg.Dict)
	tt.Nil(t, idf.LoadCorpus("not_used.txt"))
}

func TestIdfAddToken(t *testing.T) {
	idf := newTestIdf(t)
	tt.Nil(t, idf.AddToken("gse", 1))
	tt.Nil(t, idf.AddToken("search", 7, "n"))
	tt.Nil(t, idf.AddToken("engine", 3))

	tt.Equal(t, []float64{1, 3, 7}, idf.freqs)
	tt.Equal(t, 3.0, idf.median)
	tt.Equal(t, 6, idf.NumTokens())
	tt.Equal(t, 29.0, idf.TotalFreq())

	freq, pos, ok := idf.Freq("search")
	tt.Bool(t, ok)
	tt.Equal(t, 7.0, freq)
	tt.Equal(t, "n", pos)
}

func TestIdfFreqMap(t *testing.T) {
	idf := newTestIdf(t)
	m := idf.FreqMap(freqText)
	tt.Equal(t, map[string]float64{
		"hello": 0.4, "world": 0.2, "golang": 0.2, "gopher": 0.2,
	}, m)

	tt.Equal(t, 0, len(idf.FreqMap("")))
}

func TestIdfCalculateWeight(t *testing.T) {
	idf := newTestIdf(t)
	tt.Nil(t, idf.AddToken("gse", 4))

	// found in the dictionary: freq * v
	tt.Equal(t, 5.0, idf.calculateWeight("hello", 0.5))
	// not found: median * v
	tt.Equal(t, 2.0, idf.calculateWeight("gopher", 0.5))

	ws := idf.ConstructSeg(freqText)
	tt.Equal(t, 4, len(ws))
	weights := make(map[string]float64)
	for _, s := range ws {
		weights[s.Text] = s.Weight
	}
	tt.Equal(t, 4.0, weights["hello"])
	tt.Equal(t, 0.8, weights["gopher"])
}

func TestTFIDFDictStr(t *testing.T) {
	tfidf := newTestTFIDF(t)
	tt.Equal(t, 3, tfidf.NumTokens())
	tt.Equal(t, 18.0, tfidf.TotalFreq())

	tf, idf, ok := tfidf.Freq("world")
	tt.Bool(t, ok)
	tt.Equal(t, 5.0, tf)
	tt.Equal(t, 3.0, idf)

	seg := tfidf.GetSeg()
	tt.Equal(t, tfidf.Seg.Dict, seg.Dict)
	tt.Nil(t, tfidf.LoadCorpus("not_used.txt"))

	tt.Nil(t, tfidf.AddToken("gse", 7))
	tt.Nil(t, tfidf.AddToken("search", 1))
	tt.Equal(t, []float64{1, 7}, tfidf.freqs)
	tt.Equal(t, 5, tfidf.NumTokens())
	tt.Equal(t, 26.0, tfidf.TotalFreq())
}

func TestTFIDFConstructSeg(t *testing.T) {
	tfidf := newTestTFIDF(t)
	m := tfidf.FreqMap(freqText)
	tt.Equal(t, map[string]float64{
		"hello": 0.4, "world": 0.2, "golang": 0.2, "gopher": 0.2,
	}, m)

	tt.Equal(t, 20.0, tfidf.calculateWeight("hello"))
	tt.Equal(t, 0.0, tfidf.calculateWeight("gopher"))

	ws := tfidf.ConstructSeg(freqText)
	tt.Equal(t, 4, len(ws))
	weights := make(map[string]float64)
	for _, s := range ws {
		weights[s.Text] = s.Weight
	}
	tt.Equal(t, 15.0, weights["world"])
	tt.Equal(t, 12.0, weights["golang"])
}

func TestNewBM25(t *testing.T) {
	bm25 := NewBM25(nil).(*BM25)
	tt.Equal(t, types.BM25DefaultK1, bm25.K1)
	tt.Equal(t, types.BM25DefaultB, bm25.B)
	tt.NotNil(t, bm25.StopWord)

	bm25 = NewBM25(&types.BM25Setting{B: 0.5}).(*BM25)
	tt.Equal(t, types.BM25DefaultK1, bm25.K1)
	tt.Equal(t, 0.5, bm25.B)

	bm25 = NewBM25(&types.BM25Setting{K1: 2}).(*BM25)
	tt.Equal(t, 0.0, bm25.B)
}

func TestBM25DictStr(t *testing.T) {
	bm25 := newTestBM25(t)
	tt.Equal(t, 3, bm25.NumTokens())
	tt.Equal(t, 18.0, bm25.TotalFreq())

	tf, idf, ok := bm25.Freq("golang")
	tt.Bool(t, ok)
	tt.Equal(t, 3.0, tf)
	tt.Equal(t, 4.0, idf)

	seg := bm25.GetSeg()
	tt.Equal(t, bm25.Seg.Dict, seg.Dict)

	tt.Nil(t, bm25.AddToken("gse", 7))
	tt.Equal(t, 4, bm25.NumTokens())
	tt.Equal(t, 25.0, bm25.TotalFreq())
}

func TestBM25ConstructSeg(t *testing.T) {
	bm25 := newTestBM25(t)
	bm25.AverageDocLength = 5

	m := bm25.FreqMap(freqText)
	tt.Equal(t, map[string]float64{
		"hello": 0.4, "world": 0.2, "golang": 0.2, "gopher": 0.2,
	}, m)

	// k = 1.25 * (0.25 + 0.75 * 5 / 5) = 1.25
	tt.Equal(t, 1.25, bm25.calculateK(5))
	// idf * (tf * (k1 + 1)) / (tf + k) = 2 * (10 * 2.25) / (10 + 1.25)
	tt.Equal(t, 4.0, bm25.calculateWeight("hello"))
	tt.Equal(t, 0.0, bm25.calculateWeight("gopher"))

	ws := bm25.ConstructSeg(freqText)
	tt.Equal(t, 4, len(ws))
	for _, s := range ws {
		tt.Equal(t, bm25.calculateWeight(s.Text), s.Weight)
	}
}

func TestBM25LoadCorpus(t *testing.T) {
	bm25 := NewBM25(nil).(*BM25)
	bm25.Seg.SkipLog = true

	err := bm25.LoadCorpus("../../testdata/test_en.txt")
	tt.Nil(t, err)
	tt.Bool(t, bm25.AverageDocLength > 0)

	bm25.AverageDocLength = 0
	err = bm25.LoadCorpus("../../testdata/not_exist_corpus.txt")
	tt.NotNil(t, err)
	tt.Equal(t, 0.0, bm25.AverageDocLength)
}

func ExampleIdf_FreqMap() {
	idf := NewIdf().(*Idf)
	idf.Seg.SkipLog = true
	if err := idf.LoadDictStr(idfDictStr); err != nil {
		fmt.Println(err)
		return
	}

	m := idf.FreqMap("hello hello world")
	fmt.Printf("%.2f %.2f\n", m["hello"], m["world"])
	// Output: 0.67 0.33
}
