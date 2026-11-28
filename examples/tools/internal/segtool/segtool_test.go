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

package segtool

import (
	"strings"
	"testing"

	"github.com/go-ego/gse"
	"github.com/vcaesar/tt"
)

const testDict = "纽约 1758 ns\n帝国 3655 n\n帝国大厦 3000 nr\n大厦 1000 n\n你好 2000 l\n世界 2000 n\n的 50000 uj\n"

func newSeg(t *testing.T) *gse.Segmenter {
	t.Helper()
	seg := &gse.Segmenter{SkipLog: true}
	tt.Nil(t, seg.LoadDictStr(testDict))
	tt.Nil(t, seg.LoadStopStr("的"))
	return seg
}

func texts(segs []gse.SegPos) []string {
	r := make([]string, len(segs))
	for i, s := range segs {
		r[i] = s.Text
	}
	return r
}

func TestModes(t *testing.T) {
	seg := newSeg(t)
	text := "纽约的帝国大厦, 你好世界"

	want := map[string]string{
		"segment":    "[纽约 的 帝国大厦 ,   你好 世界]",
		"search":     "[纽约 的 帝国 大厦 帝国大厦 ,   你好 世界]",
		"dag":        "[纽约 的 帝国大厦 ,   你好 世界]",
		"hmm":        "[纽约 的 帝国大厦 ,  你好 世界]",
		"search_hmm": "[纽约 的 帝国 大厦 帝国大厦 ,  你好 世界]",
		"all":        "[纽约 的 帝国 帝国大厦 大厦 ,   你好 世界]",
		"hmm_only":   "[纽约 的 帝国大厦 ,  你 好 世界]",
		"trim":       "[纽约 帝国大厦 你好 世界]",
	}
	for _, m := range Modes {
		tt.Equal(t, want[m.Name], texts(m.Func(seg, text)), m.Name)
	}
	tt.Equal(t, len(want), len(Names()))

	segs := Modes[0].Func(seg, text)
	tt.Equal(t, "ns", segs[0].Pos)
	tt.Equal(t, "nr", segs[2].Pos)
}

func TestGet(t *testing.T) {
	m, err := Get(Default)
	tt.Nil(t, err)
	tt.Equal(t, Default, m.Name)

	_, err = Get("nope")
	tt.NotNil(t, err)
	tt.Bool(t, strings.Contains(err.Error(), "segment, search"))
}

func TestFormat(t *testing.T) {
	segs := []gse.SegPos{{Text: "你好", Pos: "l"}, {Text: ","}, {Text: "世界", Pos: "n"}}
	tt.Equal(t, "你好/l , 世界/n", Format(segs))
	tt.Equal(t, "", Format(nil))
}

func TestReadLines(t *testing.T) {
	lines, size, err := ReadLines("testdata/lines.txt")
	tt.Nil(t, err)
	tt.Equal(t, "[纽约的帝国大厦 hello world 你好世界]", lines)
	tt.Equal(t, len("纽约的帝国大厦 hello world")+len("你好世界"), size)

	_, _, err = ReadLines("testdata/none.txt")
	tt.NotNil(t, err)
}

func TestLoadDict(t *testing.T) {
	seg := &gse.Segmenter{SkipLog: true}
	tt.NotNil(t, LoadDict(seg, "testdata/none_dict.txt"))

	seg = &gse.Segmenter{SkipLog: true}
	tt.Nil(t, LoadDict(seg, "../../../../testdata/zh/test_dict.txt"))
	_, pos, ok := seg.Find("帝国大厦")
	tt.Bool(t, ok)
	tt.Equal(t, "nr", pos)
	tt.Bool(t, len(seg.StopWordMap) > 0)
}
