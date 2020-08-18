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
	"regexp"
	"unicode"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/util"
)

var (
	reHanDetail  = regexp.MustCompile(`(\p{Han}+)`)
	reSkipDetail = regexp.MustCompile(`([[\.[:digit:]]+|[:alnum:]]+)`)

	reEng  = regexp.MustCompile(`[[:alnum:]]`)
	reNum  = regexp.MustCompile(`[\.[:digit:]]+`)
	reEng1 = regexp.MustCompile(`[[:alnum:]]$`)

	reHanInternal  = regexp.MustCompile(`([\p{Han}+[:alnum:]+#&\._]+)`)
	reSkipInternal = regexp.MustCompile(`(\r\n|\s)`)
)

// SegPos represents a word with it's POS
type SegPos struct {
	Text, Pos string
}

// Segmenter is a words segmentation struct.
type Segmenter struct {
	dict Dict
}

// WithGse register gse segmenter
func (seg *Segmenter) WithGse(segs gse.Segmenter) {
	seg.dict.seg = segs
}

// LoadDict loads dictionary from given file name.
func (seg *Segmenter) LoadDict(fileName ...string) error {
	return seg.dict.loadDict(fileName...)
}

func (seg *Segmenter) cutDetailInternal(sentence string) (result []SegPos) {
	runes := []rune(sentence)
	posList := viterbi(runes)

	begin := 0
	next := 0
	for i, char := range runes {
		pos := posList[i]
		switch pos.position() {
		case "B":
			begin = i
		case "E":
			result = append(result, SegPos{string(runes[begin : i+1]), pos.pos()})
			next = i + 1
		case "S":
			result = append(result, SegPos{string(char), pos.pos()})
			next = i + 1
		}
	}

	if next < len(runes) {
		result = append(result, SegPos{string(runes[next:]), posList[next].pos()})
	}

	return
}

func (seg *Segmenter) cutDetail(sentence string) (result []SegPos) {
	for _, blk := range util.RegexpSplit(reHanDetail, sentence, -1) {
		if reHanDetail.MatchString(blk) {
			result = append(result, seg.cutDetailInternal(blk)...)
			continue
		}

		for _, x := range util.RegexpSplit(reSkipDetail, blk, -1) {
			if len(x) == 0 {
				continue
			}

			switch {
			case reNum.MatchString(x):
				result = append(result, SegPos{x, "m"})
			case reEng.MatchString(x):
				result = append(result, SegPos{x, "eng"})
			default:
				result = append(result, SegPos{x, "x"})
			}
		}
	}

	return
}

func (seg *Segmenter) dag(runes []rune) map[int][]int {
	dag := make(map[int][]int)
	n := len(runes)
	var frag []rune
	var i int

	for k := 0; k < n; k++ {
		dag[k] = make([]int, 0)
		i = k
		frag = runes[k : k+1]

		for {
			freq, ok := seg.dict.Frequency(string(frag))
			if !ok {
				break
			}

			if freq > 0.0 {
				dag[k] = append(dag[k], i)
			}

			i++
			if i >= n {
				break
			}

			frag = runes[k : i+1]
		}

		if len(dag[k]) == 0 {
			dag[k] = append(dag[k], k)
		}
	}

	return dag
}

type route struct {
	frequency float64
	index     int
}

func (seg *Segmenter) calc(runes []rune) map[int]route {
	dag := seg.dag(runes)
	n := len(runes)

	rs := make(map[int]route)
	rs[n] = route{frequency: 0.0, index: 0}
	var r route

	for idx := n - 1; idx >= 0; idx-- {
		for _, i := range dag[idx] {
			if freq, ok := seg.dict.Frequency(string(runes[idx : i+1])); ok {
				r = route{
					frequency: math.Log(freq) - seg.dict.logTotal + rs[i+1].frequency,
					index:     i}
			} else {
				r = route{
					frequency: math.Log(1.0) - seg.dict.logTotal + rs[i+1].frequency,
					index:     i}
			}

			if v, ok := rs[idx]; !ok {
				rs[idx] = r
			} else {
				if v.frequency < r.frequency ||
					(v.frequency == r.frequency && v.index < r.index) {
					rs[idx] = r
				}
			}
		}
	}

	return rs
}

type cutFunc func(sentence string) []SegPos

func (seg *Segmenter) cutDAG(sentence string) (result []SegPos) {
	runes := []rune(sentence)
	routes := seg.calc(runes)
	length := len(runes)

	var y int
	var buf []rune

	for x := 0; x < length; {
		y = routes[x].index + 1
		frag := runes[x:y]
		if y-x == 1 {
			buf = append(buf, frag...)
			x = y
			continue
		}

		if len(buf) > 0 {
			bufString := string(buf)
			if len(buf) == 1 {
				if tag, ok := seg.dict.Pos(bufString); ok {
					result = append(result, SegPos{bufString, tag})
				} else {
					result = append(result, SegPos{bufString, "x"})
				}

				buf = make([]rune, 0)
				continue
			}

			if v, ok := seg.dict.Frequency(bufString); !ok || v == 0.0 {
				result = append(result, seg.cutDetail(bufString)...)
			} else {
				for _, elem := range buf {
					selem := string(elem)
					if tag, ok := seg.dict.Pos(selem); ok {
						result = append(result, SegPos{selem, tag})
					} else {
						result = append(result, SegPos{selem, "x"})
					}
				}
			}
			buf = make([]rune, 0)
		}

		word := string(frag)
		if tag, ok := seg.dict.Pos(word); ok {
			result = append(result, SegPos{word, tag})
		} else {
			result = append(result, SegPos{word, "x"})
		}
		x = y
	}

	if len(buf) > 0 {
		result = seg.bufn(buf)
	}

	return
}

func (seg *Segmenter) bufn(buf []rune) (result []SegPos) {
	bufString := string(buf)
	if len(buf) == 1 {
		if tag, ok := seg.dict.Pos(bufString); ok {
			result = append(result, SegPos{bufString, tag})
		} else {
			result = append(result, SegPos{bufString, "x"})
		}

		return
	}

	if v, ok := seg.dict.Frequency(bufString); !ok || v == 0.0 {
		result = append(result, seg.cutDetail(bufString)...)
	} else {
		for _, elem := range buf {
			selem := string(elem)
			if tag, ok := seg.dict.Pos(selem); ok {
				result = append(result, SegPos{selem, tag})
			} else {
				result = append(result, SegPos{selem, "x"})
			}
		}
	}

	return
}

func (seg *Segmenter) cutDAGNoHMM(sentence string) (result []SegPos) {
	runes := []rune(sentence)
	routes := seg.calc(runes)
	var y int

	length := len(runes)
	var buf []rune
	for x := 0; x < length; {
		y = routes[x].index + 1
		frag := runes[x:y]
		if reEng1.MatchString(string(frag)) && len(frag) == 1 {
			buf = append(buf, frag...)
			x = y
			continue
		}

		if len(buf) > 0 {
			result = append(result, SegPos{string(buf), "eng"})
			buf = make([]rune, 0)
		}

		word := string(frag)
		if tag, ok := seg.dict.Pos(word); ok {
			result = append(result, SegPos{word, tag})
		} else {
			result = append(result, SegPos{word, "x"})
		}
		x = y
	}

	if len(buf) > 0 {
		result = append(result, SegPos{string(buf), "eng"})
		// buf = make([]rune, 0)
	}

	return
}

// Cut cuts a sentence into words.
// Parameter hmm controls whether to use the HMM.
func (seg *Segmenter) Cut(sentence string, hmm bool) (result []SegPos) {
	var cut cutFunc
	if hmm {
		cut = seg.cutDAG
	} else {
		cut = seg.cutDAGNoHMM
	}

	for _, blk := range util.RegexpSplit(reHanInternal, sentence, -1) {
		if reHanInternal.MatchString(blk) {
			result = append(result, cut(blk)...)
			continue
		}

		for _, x := range util.RegexpSplit(reSkipInternal, blk, -1) {
			if reSkipInternal.MatchString(x) {
				result = append(result, SegPos{x, "x"})
				continue
			}

			for _, xx := range x {
				s := string(xx)
				switch {
				case reNum.MatchString(s):
					result = append(result, SegPos{s, "m"})
				case reEng.MatchString(x):
					result = append(result, SegPos{x, "eng"})
				default:
					result = append(result, SegPos{s, "x"})
				}
			}
		}

	}

	return
}

// TrimPunct not space and punct
func (seg *Segmenter) TrimPunct(se []SegPos) (re []SegPos) {
	for i := 0; i < len(se); i++ {
		if se[i].Text != "" && len(se[i].Text) > 0 {
			ru := []rune(se[i].Text)[0]
			if !unicode.IsSpace(ru) && !unicode.IsPunct(ru) {
				re = append(re, se[i])
			}
		}
	}

	return
}

// Trim not space and punct
func (seg *Segmenter) Trim(se []SegPos) (re []SegPos) {
	for i := 0; i < len(se); i++ {
		if !seg.dict.seg.IsStop(se[i].Text) {
			si := gse.FilterSymbol(se[i].Text)
			if si != "" {
				re = append(re, se[i])
			}
		}
	}

	return
}

// TrimWithPos trim some pos
func (seg *Segmenter) TrimWithPos(se []SegPos, pos ...string) (re []SegPos) {
	for h := 0; h < len(pos); h++ {
		if h > 0 {
			se = re
			re = nil
		}

		for i := 0; i < len(se); i++ {
			if se[i].Pos != pos[h] {
				re = append(re, se[i])
			}
		}
	}

	return
}
