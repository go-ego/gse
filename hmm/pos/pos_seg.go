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

// SegPos type POS struct
type SegPos struct {
	Text, Pos string
}

// Segmenter is a segmentation struct
type Segmenter struct {
	dict Dict
}

// WithGse register the gse segmenter
func (seg *Segmenter) WithGse(segs gse.Segmenter) {
	seg.dict.Seg = segs
}

// LoadDict load dictionary from the file.
func (seg *Segmenter) LoadDict(fileName ...string) error {
	return seg.dict.loadDict(fileName...)
}

func (seg *Segmenter) cutDetailInternal(text string) (result []gse.SegPos) {
	runes := []rune(text)
	posList := viterbi(runes)

	begin := 0
	next := 0
	for i, char := range runes {
		pos := posList[i]
		switch pos.position() {
		case "B":
			begin = i
		case "E":
			result = append(result, gse.SegPos{Text: string(runes[begin : i+1]), Pos: pos.pos()})
			next = i + 1
		case "S":
			result = append(result, gse.SegPos{Text: string(char), Pos: pos.pos()})
			next = i + 1
		}
	}

	if next < len(runes) {
		result = append(result, gse.SegPos{Text: string(runes[next:]), Pos: posList[next].pos()})
	}

	return
}

func (seg *Segmenter) cutDetail(text string) (result []gse.SegPos) {
	for _, blk := range util.RegexpSplit(reHanDetail, text, -1) {
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
				result = append(result, gse.SegPos{Text: x, Pos: "m"})
			case reEng.MatchString(x):
				result = append(result, gse.SegPos{Text: x, Pos: "eng"})
			default:
				result = append(result, gse.SegPos{Text: x, Pos: "x"})
			}
		}
	}

	return
}

func (seg *Segmenter) getDag(runes []rune) map[int][]int {
	dag := make(map[int][]int)
	n := len(runes)
	var frag []rune
	var i int

	for k := 0; k < n; k++ {
		dag[k] = make([]int, 0)
		i = k
		frag = runes[k : k+1]

		for {
			freq, _, ok := seg.dict.Freq(string(frag))
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
	freq  float64
	index int
}

func (seg *Segmenter) calc(runes []rune) map[int]route {
	dag := seg.getDag(runes)
	n := len(runes)

	rs := make(map[int]route)
	rs[n] = route{freq: 0.0, index: 0}
	var r route

	for idx := n - 1; idx >= 0; idx-- {
		for _, i := range dag[idx] {
			if freq, _, ok := seg.dict.Freq(string(runes[idx : i+1])); ok {
				r = route{
					freq:  math.Log(freq) - seg.dict.logTotal + rs[i+1].freq,
					index: i}
			} else {
				r = route{
					freq:  math.Log(1.0) - seg.dict.logTotal + rs[i+1].freq,
					index: i}
			}

			if v, ok := rs[idx]; !ok {
				rs[idx] = r
			} else {
				if v.freq < r.freq ||
					(v.freq == r.freq && v.index < r.index) {
					rs[idx] = r
				}
			}
		}
	}

	return rs
}

type cutFunc func(text string) []gse.SegPos

func (seg *Segmenter) cutDAG(text string) (result []gse.SegPos) {
	runes := []rune(text)
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
					result = append(result, gse.SegPos{Text: bufString, Pos: tag})
				} else {
					result = append(result, gse.SegPos{Text: bufString, Pos: "x"})
				}

				buf = make([]rune, 0)
				continue
			}

			if v, _, ok := seg.dict.Freq(bufString); !ok || v == 0.0 {
				result = append(result, seg.cutDetail(bufString)...)
			} else {
				for _, elem := range buf {
					selem := string(elem)
					if tag, ok := seg.dict.Pos(selem); ok {
						result = append(result, gse.SegPos{Text: selem, Pos: tag})
					} else {
						result = append(result, gse.SegPos{Text: selem, Pos: "x"})
					}
				}
			}
			buf = make([]rune, 0)
		}

		word := string(frag)
		if tag, ok := seg.dict.Pos(word); ok {
			result = append(result, gse.SegPos{Text: word, Pos: tag})
		} else {
			result = append(result, gse.SegPos{Text: word, Pos: "x"})
		}
		x = y
	}

	if len(buf) > 0 {
		result = append(result, seg.bufn(buf)...)
	}

	return
}

func (seg *Segmenter) bufn(buf []rune) (result []gse.SegPos) {
	bufString := string(buf)
	if len(buf) == 1 {
		if tag, ok := seg.dict.Pos(bufString); ok {
			result = append(result, gse.SegPos{Text: bufString, Pos: tag})
		} else {
			result = append(result, gse.SegPos{Text: bufString, Pos: "x"})
		}

		return
	}

	if v, _, ok := seg.dict.Freq(bufString); !ok || v == 0.0 {
		result = append(result, seg.cutDetail(bufString)...)
		return
	}

	for _, elem := range buf {
		selem := string(elem)
		if tag, ok := seg.dict.Pos(selem); ok {
			result = append(result, gse.SegPos{Text: selem, Pos: tag})
		} else {
			result = append(result, gse.SegPos{Text: selem, Pos: "x"})
		}
	}

	return
}

func (seg *Segmenter) cutDAGNoHMM(text string) (result []gse.SegPos) {
	runes := []rune(text)
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
			result = append(result, gse.SegPos{Text: string(buf), Pos: "eng"})
			buf = make([]rune, 0)
		}

		word := string(frag)
		if tag, ok := seg.dict.Pos(word); ok {
			result = append(result, gse.SegPos{Text: word, Pos: tag})
		} else {
			result = append(result, gse.SegPos{Text: word, Pos: "x"})
		}
		x = y
	}

	if len(buf) > 0 {
		result = append(result, gse.SegPos{Text: string(buf), Pos: "eng"})
		// buf = make([]rune, 0)
	}

	return
}

// Cut cuts a text into words.
// Parameter hmm controls whether to use the HMM
func (seg *Segmenter) Cut(text string, hmm ...bool) (result []gse.SegPos) {
	var cut cutFunc
	if len(hmm) > 0 && hmm[0] {
		cut = seg.cutDAG
	} else {
		cut = seg.cutDAGNoHMM
	}

	for _, blk := range util.RegexpSplit(reHanInternal, text, -1) {
		if reHanInternal.MatchString(blk) {
			result = append(result, cut(blk)...)
			continue
		}

		for _, x := range util.RegexpSplit(reSkipInternal, blk, -1) {
			if reSkipInternal.MatchString(x) {
				result = append(result, gse.SegPos{Text: x, Pos: "x"})
				continue
			}

			for _, xx := range x {
				s := string(xx)
				switch {
				case reNum.MatchString(s):
					result = append(result, gse.SegPos{Text: s, Pos: "m"})
				case reEng.MatchString(x):
					result = append(result, gse.SegPos{Text: x, Pos: "eng"})
				default:
					result = append(result, gse.SegPos{Text: s, Pos: "x"})
				}
			}
		}

	}

	return
}

// TrimPunct not space and punct
func (seg *Segmenter) TrimPunct(se []gse.SegPos) (re []gse.SegPos) {
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
func (seg *Segmenter) Trim(se []gse.SegPos) (re []gse.SegPos) {
	for i := 0; i < len(se); i++ {
		si := gse.FilterSymbol(se[i].Text)
		if !seg.dict.Seg.NotStop && seg.dict.Seg.IsStop(si) {
			si = ""
		}

		if si != "" {
			re = append(re, se[i])
		}
	}

	return
}

// TrimWithPos trim some pos
func (seg *Segmenter) TrimWithPos(se []gse.SegPos, pos ...string) (re []gse.SegPos) {
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
