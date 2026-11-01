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

package hmm

import (
	"testing"

	"github.com/vcaesar/tt"
)

var bmes = []byte{'B', 'M', 'E', 'S'}

func TestLoadModel(t *testing.T) {
	defer LoadModel()

	b := map[rune]float64{'甲': -1}
	e := map[rune]float64{'乙': -2}
	m := map[rune]float64{'丙': -3}
	s := map[rune]float64{'丁': -4}
	LoadModel(b, e, m, s)
	tt.Equal(t, 1, len(probEmit['B']))
	tt.Equal(t, -1.0, probEmit['B']['甲'])
	tt.Equal(t, -2.0, probEmit['E']['乙'])
	tt.Equal(t, -3.0, probEmit['M']['丙'])
	tt.Equal(t, -4.0, probEmit['S']['丁'])

	LoadModel()
	tt.Bool(t, len(probEmit['B']) > 100)
	tt.Bool(t, len(probEmit['S']) > 100)
}

func TestInternalCutSingle(t *testing.T) {
	tt.Equal(t, []string{"我"}, internalCut("我"))
}

func TestCutWithReg(t *testing.T) {
	text := "西雅图太空针, 纽约时代广场, New York City."
	tt.Equal(t, Cut(text), Cut(text, regSkip, regHan))
	tt.Equal(t, Cut(text), Cut(text, regSkip))
}

func TestCutNonCutFirst(t *testing.T) {
	tt.Equal(t, []string{", ", "a", "我"}, Cut(", a我"))
}

func TestLocJudge(t *testing.T) {
	tt.Bool(t, locJudge("a", nil, nil) == nil)
	tt.Bool(t, locJudge("", nil, nil) == nil)
	tt.Equal(t, []int{1, 2}, locJudge("abc", nil, []int{1, 2}))
	tt.Equal(t, []int{1, 2}, locJudge("abc", []int{1, 2}, nil))
	tt.Equal(t, []int{1, 2}, locJudge("abc", []int{1, 2}, []int{2, 3}))
	tt.Equal(t, []int{1, 2}, locJudge("abc", []int{2, 3}, []int{1, 2}))
}

func TestProbState(t *testing.T) {
	tt.Equal(t, "(1.500000, 42)", probState{prob: 1.5, state: 'B'}.String())

	ps := probStates{{prob: 1, state: 'B'}, {prob: 1, state: 'E'}, {prob: 2, state: 'B'}}
	tt.Equal(t, 3, ps.Len())
	tt.Bool(t, ps.Less(0, 1))
	tt.Bool(t, !ps.Less(1, 0))
	tt.Bool(t, ps.Less(1, 2))

	ps.Swap(0, 2)
	tt.Equal(t, 2.0, ps[0].prob)
}

func TestViterbiUnknown(t *testing.T) {
	_, path := Viterbi([]rune("a"), bmes)
	tt.Equal(t, []byte{'S'}, path)

	_, path = Viterbi([]rune("我a"), bmes)
	tt.Equal(t, 2, len(path))
}

func TestProbsNoTrans(t *testing.T) {
	prevStatus['X'] = []byte{'B'}
	defer delete(prevStatus, 'X')

	vtb := []map[byte]float64{{'B': -1}, {}}
	ps := probs([]rune("我们"), vtb, 'X', 1)
	tt.Equal(t, 1, len(ps))
	tt.Equal(t, byte('B'), ps[0].state)
	tt.Equal(t, -1+minFloat+minFloat, ps[0].prob)
}
