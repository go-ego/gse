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

package gse

import (
	"testing"

	"github.com/vcaesar/tt"
)

// newDagSeg returns a small segmenter, "甲乙" is rarer than "甲" + "乙"
// so the DAG route splits it into single runes
func newDagSeg(t *testing.T) *Segmenter {
	seg := &Segmenter{SkipLog: true, NotLoadHMM: true}
	seg.Dict = NewDict()
	seg.Init()

	tt.Nil(t, seg.AddToken("甲", 1000))
	tt.Nil(t, seg.AddToken("乙", 1000))
	tt.Nil(t, seg.AddToken("甲乙", 3))
	tt.Nil(t, seg.AddToken("丙", 10))
	tt.Nil(t, seg.AddToken("丁戊", 100))
	seg.CalcToken()
	return seg
}

func TestFindTFIDF(t *testing.T) {
	seg := newDagSeg(t)
	freq, idf, ok := seg.FindTFIDF("丁戊")
	tt.Bool(t, ok)
	tt.Equal(t, 100, freq)
	tt.Equal(t, 0, idf)

	_, _, ok = seg.FindTFIDF("不存在")
	tt.False(t, ok)
}

func TestAnalyzeNoText(t *testing.T) {
	seg := newDagSeg(t)
	tt.Equal(t, 0, len(seg.Analyze(nil, "")))

	a := seg.Analyze([]string{"丙", "丁戊"}, "")
	tt.Equal(t, 2, len(a))
	tt.Equal(t, 0, a[0].Start)
	tt.Equal(t, 3, a[0].End)
	tt.Equal(t, 3, a[1].Start)
	tt.Equal(t, 9, a[1].End)
	tt.Equal(t, 100, a[1].Freq)
}

func TestCutDAGBuf(t *testing.T) {
	seg := newDagSeg(t)

	// the buffered "甲乙" is in the dictionary, so hmm keeps single runes
	tt.Equal(t, "[甲 乙]", seg.cutDAG("甲乙"))
	// one buffered rune followed by a word
	tt.Equal(t, "[丙 丁戊]", seg.cutDAG("丙丁戊"))
	// trailing alpha numbers are buffered
	tt.Equal(t, "[丁戊 abc]", seg.cutDAGNoHMM("丁戊abc"))
}

func TestSuggestFreqWords(t *testing.T) {
	seg := newDagSeg(t)

	// the joined word is found and rarer than the words
	tt.Equal(t, 3, seg.SuggestFreq("甲", "乙"))
	tt.Equal(t, 0, seg.SuggestFreq("丙", "不存在"))

	ToLower = false
	tt.Nil(t, seg.AddToken("AB", 100))
	ToLower = true
	defer func() { ToLower = true }()

	// the cut is lower case and not found, the word freq is used
	tt.Equal(t, 100, seg.SuggestFreq("AB"))
}
