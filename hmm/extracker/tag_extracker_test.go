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

package extracker

import (
	"testing"

	"github.com/vcaesar/tt"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/relevance"
	"github.com/go-ego/gse/types"
)

const (
	posDictStr = "我 100 r\n喜欢 100 v\n吃 100 v\n苹果 100 n\n香蕉 80 n\n水果 60 n\n和 100 c\n都是 100 d"
	idfDictStr = "喜欢 2\n苹果 8\n香蕉 6\n水果 4"
	tfidfStr   = "喜欢 2 1\n苹果 8 3\n香蕉 6 2\n水果 4 5"
	extText    = "我喜欢吃苹果,我喜欢吃香蕉,苹果和香蕉都是水果"

	testDict    = "../../testdata/test_en_dict3.txt"
	notExistDic = "../../testdata/not_exist_dict.txt"
)

func newPosSeg(t *testing.T) gse.Segmenter {
	var seg gse.Segmenter
	seg.SkipLog = true
	tt.Nil(t, seg.LoadDictStr(posDictStr))
	return seg
}

func newIdfRelevance(t *testing.T, dict string) relevance.Relevance {
	r := relevance.NewIdf()
	r.(*relevance.Idf).Seg.SkipLog = true
	tt.Nil(t, r.LoadDictStr(dict))
	return r
}

func TestTagExtracterLoadDict(t *testing.T) {
	var te TagExtracter
	var seg gse.Segmenter
	seg.SkipLog = true
	te.WithGse(seg)

	tt.Nil(t, te.LoadDict(testDict))
	freq, _, ok := te.seg.Find("hello")
	tt.Bool(t, ok)
	tt.Equal(t, 20.0, freq)

	tt.NotNil(t, te.LoadDict(notExistDic))
}

func TestTagExtracterLoadIdf(t *testing.T) {
	var te TagExtracter
	tt.Nil(t, te.LoadIdf(testDict))
	tt.Equal(t, 3, te.Relevance.NumTokens())

	tt.NotNil(t, te.LoadIdf(notExistDic))

	tt.Nil(t, te.LoadIdfStr(idfDictStr))
	tt.Equal(t, 4, te.Relevance.NumTokens())
	_, ok := te.Relevance.(*relevance.Idf)
	tt.Bool(t, ok)
}

func TestTagExtracterLoadTFIDF(t *testing.T) {
	var te TagExtracter
	tt.NotNil(t, te.LoadTFIDF(notExistDic))
	_, ok := te.Relevance.(*relevance.TFIDF)
	tt.Bool(t, ok)
}

func TestTagExtracterLoadBM25(t *testing.T) {
	var te TagExtracter

	err := te.LoadBM25(&types.BM25Setting{K1: 1.5, B: 0.6}, []*types.LoadDictFile{
		{FilePath: "../../testdata/test_en.txt", FileType: types.LoadDictCorpus},
		{FilePath: notExistDic, FileType: types.LoadDictTypeBM25},
		{FilePath: testDict, FileType: types.LoadDictTypeIDF},
	})
	tt.NotNil(t, err)
	bm25, ok := te.Relevance.(*relevance.BM25)
	tt.Bool(t, ok)
	tt.Equal(t, 1.5, bm25.K1)
	tt.Equal(t, 0.6, bm25.B)
	tt.Bool(t, bm25.AverageDocLength > 0)

	err = te.LoadBM25(nil, []*types.LoadDictFile{
		{FilePath: "../../testdata/not_exist_corpus.txt", FileType: types.LoadDictCorpus},
	})
	tt.NotNil(t, err)
	tt.Equal(t, 0.0, te.Relevance.(*relevance.BM25).AverageDocLength)
}

func TestTagExtracterStopWords(t *testing.T) {
	var te TagExtracter
	te.Relevance = newIdfRelevance(t, idfDictStr)

	tt.Nil(t, te.LoadStopWords("../../testdata/stop.txt"))
	tt.NotNil(t, te.LoadStopWords(notExistDic))
}

// The default Idf has no dictionary loaded, so cutting panics on the nil
// dictionary; the default Relevance must still be set.
func TestExtractTagsDefault(t *testing.T) {
	var te TagExtracter
	defer func() {
		tt.NotNil(t, recover())
		_, ok := te.Relevance.(*relevance.Idf)
		tt.Bool(t, ok)
	}()

	te.ExtractTags("", 5)
}

func TestExtractTags(t *testing.T) {
	var te TagExtracter
	te.Relevance = newIdfRelevance(t, idfDictStr)

	// freq map: 喜欢 2/7, 苹果 2/7, 香蕉 2/7, 水果 1/7
	all := te.ExtractTags(extText, 10)
	tt.Equal(t, 4, len(all))
	tt.Equal(t, "苹果", all[0].Text)
	tt.Equal(t, "香蕉", all[1].Text)
	tt.Equal(t, "喜欢", all[3].Text)

	top := te.ExtractTags(extText, 2)
	tt.Equal(t, 2, len(top))
	tt.Equal(t, "苹果", top[0].Text)
	tt.Equal(t, "香蕉", top[1].Text)
}

func TestExtractTagsTFIDF(t *testing.T) {
	var te TagExtracter
	te.Relevance = relevance.NewTFIDF()
	te.Relevance.(*relevance.TFIDF).Seg.SkipLog = true
	tt.Nil(t, te.Relevance.LoadDictStr(tfidfStr))

	// weight = tf * idf: 苹果 24, 水果 20, 香蕉 12, 喜欢 2
	tags := te.ExtractTags(extText, 3)
	tt.Equal(t, 3, len(tags))
	tt.Equal(t, "苹果", tags[0].Text)
	tt.Equal(t, 24.0, tags[0].Weight)
	tt.Equal(t, "水果", tags[1].Text)
	tt.Equal(t, "香蕉", tags[2].Text)
}
