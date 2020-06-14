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

// Segment represents a word with it's POS
type Segment struct {
	Text, Pos string
}

// Segmenter is a words segmentation struct.
type Segmenter struct {
	dict Dict
}

// LoadDict loads dictionary from given file name.
func (seg *Segmenter) LoadDict(fileName ...string) error {
	return seg.dict.loadDict(fileName...)
}

func (seg *Segmenter) cutDetailInternal(sentence string) (result []Segment) {
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
			result = append(result, Segment{string(runes[begin : i+1]), pos.pos()})
			next = i + 1
		case "S":
			result = append(result, Segment{string(char), pos.pos()})
			next = i + 1
		}
	}

	if next < len(runes) {
		result = append(result, Segment{string(runes[next:]), posList[next].pos()})
	}

	return
}

func (seg *Segmenter) cutDetail(sentence string) (result []Segment) {
	for _, blk := range util.RegexpSplit(reHanDetail, sentence, -1) {
		if reHanDetail.MatchString(blk) {
			for _, segment := range seg.cutDetailInternal(blk) {
				result = append(result, segment)
			}
			continue
		}

		for _, x := range util.RegexpSplit(reSkipDetail, blk, -1) {
			if len(x) == 0 {
				continue
			}

			switch {
			case reNum.MatchString(x):
				result = append(result, Segment{x, "m"})
			case reEng.MatchString(x):
				result = append(result, Segment{x, "eng"})
			default:
				result = append(result, Segment{x, "x"})
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
				r = route{frequency: math.Log(float64(freq)) - seg.dict.logTotal + rs[i+1].frequency,
					index: i}
			} else {
				r = route{frequency: math.Log(1.0) - seg.dict.logTotal + rs[i+1].frequency,
					index: i}
			}

			if v, ok := rs[idx]; !ok {
				rs[idx] = r
			} else {
				if v.frequency < r.frequency || (v.frequency == r.frequency && v.index < r.index) {
					rs[idx] = r
				}
			}
		}
	}

	return rs
}

type cutFunc func(sentence string) []Segment

func (seg *Segmenter) cutDAG(sentence string) (result []Segment) {
	runes := []rune(sentence)
	routes := seg.calc(runes)
	var y int
	length := len(runes)
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
					result = append(result, Segment{bufString, tag})
				} else {
					result = append(result, Segment{bufString, "x"})
				}

				buf = make([]rune, 0)
				continue
			}

			if v, ok := seg.dict.Frequency(bufString); !ok || v == 0.0 {
				for _, t := range seg.cutDetail(bufString) {
					result = append(result, t)
				}

			} else {
				for _, elem := range buf {
					selem := string(elem)
					if tag, ok := seg.dict.Pos(selem); ok {
						result = append(result, Segment{selem, tag})
					} else {
						result = append(result, Segment{selem, "x"})
					}
				}
			}
			buf = make([]rune, 0)
		}

		word := string(frag)
		if tag, ok := seg.dict.Pos(word); ok {
			result = append(result, Segment{word, tag})
		} else {
			result = append(result, Segment{word, "x"})
		}
		x = y
	}

	if len(buf) > 0 {
		bufString := string(buf)
		if len(buf) == 1 {
			if tag, ok := seg.dict.Pos(bufString); ok {
				result = append(result, Segment{bufString, tag})
			} else {
				result = append(result, Segment{bufString, "x"})
			}
		} else {
			if v, ok := seg.dict.Frequency(bufString); !ok || v == 0.0 {
				for _, t := range seg.cutDetail(bufString) {
					result = append(result, t)
				}
			} else {
				for _, elem := range buf {
					selem := string(elem)
					if tag, ok := seg.dict.Pos(selem); ok {
						result = append(result, Segment{selem, tag})
					} else {
						result = append(result, Segment{selem, "x"})
					}
				}
			}
		}
	}

	return
}

func (seg *Segmenter) cutDAGNoHMM(sentence string) (result []Segment) {
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
			result = append(result, Segment{string(buf), "eng"})
			buf = make([]rune, 0)
		}

		word := string(frag)
		if tag, ok := seg.dict.Pos(word); ok {
			result = append(result, Segment{word, tag})
		} else {
			result = append(result, Segment{word, "x"})
		}
		x = y
	}

	if len(buf) > 0 {
		result = append(result, Segment{string(buf), "eng"})
		// buf = make([]rune, 0)
	}

	return
}

// Cut cuts a sentence into words.
// Parameter hmm controls whether to use the HMM.
func (seg *Segmenter) Cut(sentence string, hmm bool) (result []Segment) {
	var cut cutFunc
	if hmm {
		cut = seg.cutDAG
	} else {
		cut = seg.cutDAGNoHMM
	}

	for _, blk := range util.RegexpSplit(reHanInternal, sentence, -1) {
		if reHanInternal.MatchString(blk) {
			for _, wordTag := range cut(blk) {
				result = append(result, wordTag)
			}
			continue
		}

		for _, x := range util.RegexpSplit(reSkipInternal, blk, -1) {
			if reSkipInternal.MatchString(x) {
				result = append(result, Segment{x, "x"})
				continue
			}

			for _, xx := range x {
				s := string(xx)
				switch {
				case reNum.MatchString(s):
					result = append(result, Segment{s, "m"})
				case reEng.MatchString(x):
					result = append(result, Segment{x, "eng"})
				default:
					result = append(result, Segment{s, "x"})
				}
			}
		}

	}

	return
}
