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

package pos

import (
	"math"
	"testing"

	"github.com/go-ego/gse"
	"github.com/vcaesar/tt"
)

type testToken struct {
	text string
	freq float64
	pos  string
}

// newTinySeg builds a segmenter on a small in-memory dictionary.
// The first token gets index 0, which Dict.Pos reports as not found.
func newTinySeg(t *testing.T, tokens ...testToken) *Segmenter {
	var s Segmenter
	s.WithGse(gse.Segmenter{Dict: gse.NewDict(), SkipLog: true})
	for _, tk := range tokens {
		tt.Nil(t, s.dict.AddToken(tk.text, tk.freq, tk.pos))
	}

	return &s
}

// tinyTokens: "甲" is index 0 (Pos not found), "甲乙" is a word whose
// route prefers the two single characters.
var tinyTokens = []testToken{
	{"甲", 100, "a"}, {"乙", 100, "b"}, {"甲乙", 2, "c"}, {"丙丁", 100, "n"},
}

func TestTag(t *testing.T) {
	tt.Equal(t, "B", tag(125).position())
	tt.Equal(t, "E", tag(225).position())
	tt.Equal(t, "M", tag(325).position())
	tt.Equal(t, "S", tag(425).position())
	tt.Equal(t, "", tag(25).position())

	tg, err := newTag("M", "nr")
	tt.Nil(t, err)
	tt.Equal(t, "M", tg.position())
	tt.Equal(t, "nr", tg.pos())

	tg, err = newTag("B", "n")
	tt.Nil(t, err)
	tt.Equal(t, tag(125), tg)

	_, err = newTag("X", "n")
	tt.Equal(t, "failed to convert X n to Tag", err.Error())
	_, err = newTag("B", "zz")
	tt.Equal(t, "failed to convert B zz to Tag", err.Error())
}

func TestTabGet(t *testing.T) {
	tt.Equal(t, probTransKeys, charStateTabMap{}.get('a'))
	tt.Equal(t, []uint16{413}, charStateTab.get('\u9fa0'))

	tt.Equal(t, minFloat, runeFloatMap{}.get('a'))
	tt.Equal(t, 1.5, runeFloatMap{'a': 1.5}.get('a'))
}

func TestDict(t *testing.T) {
	s := newTinySeg(t, tinyTokens...)
	d := &s.dict

	freq, po, ok := d.Freq("乙")
	tt.Bool(t, ok)
	tt.Equal(t, 100.0, freq)
	tt.Equal(t, "b", po)

	po, ok = d.Pos("乙")
	tt.Bool(t, ok)
	tt.Equal(t, "b", po)

	po, ok = d.Pos("戊")
	tt.Bool(t, !ok)
	tt.Equal(t, "", po)

	tt.Nil(t, d.RemoveToken("丙丁"))
	_, ok = d.Pos("丙丁")
	tt.Bool(t, !ok)

	d.total = 100
	d.updateLogTotal()
	tt.Equal(t, math.Log(100), d.logTotal)
}

func TestProbState(t *testing.T) {
	tt.Equal(t, "(100: 1.500000)", probState{prob: 1.5, state: 100}.String())
}

func TestViterbiNoExpectState(t *testing.T) {
	// no transition from the states of '仳' to the only state of '哏'
	route := viterbi([]rune("仳哏"))
	tt.Equal(t, 2, len(route))
}

func TestProbsNoPrevTrans(t *testing.T) {
	// 102 has no outgoing transitions
	tt.Equal(t, 0, len(probTrans[102]))

	vtb := []map[uint16]float64{{102: 0}, nil}
	memPath := []map[uint16]uint16{{102: 0}, nil}
	memPath, vtb = probs([]rune("甲乙"), vtb, memPath, 2)
	tt.Equal(t, len(probTransKeys), len(memPath[1]))
	tt.Equal(t, len(probTransKeys), len(vtb[1]))
}

func TestCutDetail(t *testing.T) {
	r := seg.cutDetailInternal("纽约时代广场")
	tt.Equal(t, "[{纽约 ns} {时代 n} {广场 n}]", r)

	r = seg.cutDetail("abc, 1.2")
	tt.Equal(t, "[{abc,  eng} {1.2 m}]", r)
	tt.Equal(t, "[{,  x}]", seg.cutDetail(", "))
}

func TestCutDAGTiny(t *testing.T) {
	s := newTinySeg(t, tinyTokens...)

	tt.Equal(t, "[{甲 x} {丙丁 n}]", s.cutDAG("甲丙丁"))
	tt.Equal(t, "[{甲 x} {乙 b} {丙丁 n}]", s.cutDAG("甲乙丙丁"))
	tt.Equal(t, "[{甲 x} {乙 b}]", s.cutDAG("甲乙"))
	tt.Equal(t, "[{甲 x}]", s.cutDAG("甲"))
	tt.Equal(t, "[{乙 b}]", s.cutDAG("乙"))
	tt.Equal(t, "[{ab eng}]", s.cutDAG("ab"))

	tt.Equal(t, "[{甲 x} {乙 b}]", s.cutDAGNoHMM("甲乙"))
	tt.Equal(t, "[{丙丁 n} {ab eng}]", s.cutDAGNoHMM("丙丁ab"))

	s0 := newTinySeg(t, testToken{"丙丁", 100, "n"})
	tt.Equal(t, "[{丙丁 x}]", s0.cutDAG("丙丁"))
	tt.Equal(t, "[{丙丁 x}]", s0.cutDAGNoHMM("丙丁"))
}

func TestCutTiny(t *testing.T) {
	s := newTinySeg(t, tinyTokens...)
	tt.Equal(t, "[{甲 x} {乙 b} {, x} {  x} {丙丁 n}]", s.Cut("甲乙, 丙丁", true))
	tt.Equal(t, 0, len(s.Cut("")))
}

func TestTrim(t *testing.T) {
	s := newTinySeg(t, tinyTokens...)
	s.dict.Seg.StopWordMap = map[string]bool{"乙": true}

	se := []gse.SegPos{
		{Text: "", Pos: "x"}, {Text: " ", Pos: "x"}, {Text: ",", Pos: "x"},
		{Text: "甲", Pos: "a"}, {Text: "乙", Pos: "b"}, {Text: "丙丁", Pos: "n"},
	}
	tt.Equal(t, "[{甲 a} {乙 b} {丙丁 n}]", s.TrimPunct(se))
	tt.Equal(t, "[{甲 a} {丙丁 n}]", s.Trim(se))

	s.dict.Seg.NotStop = true
	tt.Equal(t, "[{甲 a} {乙 b} {丙丁 n}]", s.Trim(se))
}
