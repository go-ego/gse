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

func TestDictFindEdge(t *testing.T) {
	seg := newDagSeg(t)

	// the empty word jumps to the root, which has no value
	_, _, ok := seg.Dict.Find([]byte(""))
	tt.False(t, ok)
	_, _, ok = seg.Dict.FindTFIDF([]byte(""))
	tt.False(t, ok)

	// "丁" is only a prefix of "丁戊"
	freq, idf, ok := seg.Dict.FindTFIDF([]byte("丁"))
	tt.Bool(t, ok)
	tt.Equal(t, 0, freq)
	tt.Equal(t, 0, idf)

	_, _, ok = seg.Dict.FindTFIDF([]byte("不存在"))
	tt.False(t, ok)
}

func TestTokenToBytes(t *testing.T) {
	sub := &Token{text: []byte("丁"), pos: "n"}
	token := &Token{text: []byte("丁戊"), pos: "nz",
		segments: []*Segment{{token: sub}}}
	tt.Equal(t, "丁/n 丁戊/nz ", string(tokenToBytes(token)))
}

func TestTextJoin(t *testing.T) {
	tt.Equal(t, "", Join(nil))
	tt.Equal(t, "a", Join([]Text{Text("a")}))
	tt.Equal(t, "ab", Join([]Text{Text("a"), Text("b")}))
	tt.Equal(t, "abc", Join([]Text{Text("a"), Text("b"), Text("c")}))
	tt.Equal(t, "abcd", Join([]Text{Text("a"), Text("b"), Text("c"), Text("d")}))

	tt.Equal(t, "a", string(textSliceToBytes([]Text{Text("a")})))
	tt.Equal(t, "ab", string(textSliceToBytes([]Text{Text("a"), Text("b")})))
}

func TestSegmentEdge(t *testing.T) {
	var seg Segmenter
	tt.Equal(t, 0, len(seg.Segment(nil)))
	// no dictionary loaded
	tt.Equal(t, 0, len(seg.segmentWords([]Text{Text("a"), Text("b")}, false)))

	tt.Equal(t, 2, maxInt(1, 2))
	tt.Equal(t, 2, maxInt(2, 1))
}

func TestLoadStopArrNil(t *testing.T) {
	var seg Segmenter
	seg.LoadStopArr([]string{"停止", "词语"})
	tt.Bool(t, seg.IsStop("停止"))
	tt.Bool(t, seg.IsStop("词语"))

	tt.NotNil(t, seg.LoadStop("testdata/not_exist.txt"))
}

func TestTrimEdge(t *testing.T) {
	var seg Segmenter
	tt.Equal(t, "[a 。a]", seg.TrimPunct([]string{"", "a", "，。", "。a"}))

	seg.LoadStopArr([]string{"的"})
	r := seg.TrimPos([]SegPos{{Text: "的", Pos: "u"}, {Text: "猫", Pos: "n"}})
	tt.Equal(t, 1, len(r))
	tt.Equal(t, "猫", r[0].Text)
}
