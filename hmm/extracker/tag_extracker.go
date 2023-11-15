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
	"sort"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/consts"
	"github.com/go-ego/gse/hmm/relevance"
	"github.com/go-ego/gse/hmm/segment"
	"github.com/go-ego/gse/types"
)

// TagExtracter is extract tags struct.
type TagExtracter struct {
	seg gse.Segmenter

	// calculate weight by Relevance(including IDF,TF-IDF,BM25 and so on)
	Relevance relevance.Relevance
	// stopWord *stopwords.StopWord
}

// WithGse register the gse segmenter
func (t *TagExtracter) WithGse(segs gse.Segmenter) {
	t.seg = segs
}

// LoadDict load and create a new dictionary from the file
func (t *TagExtracter) LoadDict(fileName ...string) error {
	return t.seg.LoadDict(fileName...)
}

// LoadIdf load and create a new Idf dictionary from the file.
func (t *TagExtracter) LoadIdf(fileName ...string) error {
	t.Relevance = relevance.NewIdf()
	return t.Relevance.LoadDict(fileName...)
}

// LoadIdfStr load and create a new Idf dictionary from the string.
func (t *TagExtracter) LoadIdfStr(str string) error {
	t.Relevance = relevance.NewIdf()
	return t.Relevance.LoadDictStr(str)
}

// LoadTFIDF load and create a new TFIDF dictionary from the file.
func (t *TagExtracter) LoadTFIDF(fileName ...string) error {
	t.Relevance = relevance.NewTFIDF()
	return t.Relevance.LoadDict(fileName...)
}

// LoadNewTFIDFStr load and create a new TFIDF dictionary from the string.
func (t *TagExtracter) LoadNewTFIDFStr(str string) error {
	t.Relevance = relevance.NewTFIDF()
	return t.Relevance.LoadDictStr(str)
}

// LoadBM25 load and create a new BM25 dictionary from the file.
func (t *TagExtracter) LoadBM25(setting *types.BM25Setting, fileList []*types.LoadBM25DictFile) (err error) {
	t.Relevance = relevance.NewBM25(setting)
	// load dict file and corpus file
	dictBM25 := []string{}
	corpusBM25 := []string{}

	for _, v := range fileList {
		switch v.FileType {

		case consts.LoadDictCorpus:
			corpusBM25 = append(corpusBM25, v.FilePath)

		case consts.LoadDictTypeBM25:
			dictBM25 = append(dictBM25, v.FilePath)
		}
	}

	err = t.Relevance.LoadCorpus(corpusBM25...)
	if err != nil {
		return
	}

	return t.Relevance.LoadDict(dictBM25...)
}

// LoadNewBM25Str load and create a new BM25 dictionary from the string.
func (t *TagExtracter) LoadNewBM25Str(setting *types.BM25Setting, str string) error {
	t.Relevance = relevance.NewBM25(setting)
	return t.Relevance.LoadDictStr(str)
}

// LoadStopWords load and create a new StopWord dictionary from the file.
func (t *TagExtracter) LoadStopWords(fileName ...string) error {
	return t.Relevance.LoadStopWord(fileName...)
}

// ExtractTags extract the topK keywords from text.
func (t *TagExtracter) ExtractTags(text string, topK int) (tags segment.Segments) {
	if t.Relevance == nil {
		// If no correlation algorithm, we will set the idf for default.
		t.Relevance = relevance.NewIdf()
	}

	// handler text to construct segment with weight
	ws := t.Relevance.ConstructSeg(text)

	// sort by weight desc
	sort.Sort(sort.Reverse(ws))

	// choose the top keywords if length of weightSeg bigger than topK
	if len(ws) > topK {
		tags = ws[:topK]
		return
	}

	tags = ws
	return
}
