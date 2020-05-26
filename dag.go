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
	"math"
	"regexp"
)

const (
	// RatioWord ratio words and letters
	RatioWord float32 = 1.5
	// RatioWordFull full ratio words and letters
	RatioWordFull float32 = 1
)

var reEng = regexp.MustCompile(`[[:alnum:]]`)

type route struct {
	frequency float64
	index     int
}

// Find find word in dictionary return word's frequency and existence
func (seg *Segmenter) Find(str string) (int, bool) {
	return seg.dict.Find([]byte(str))
}

func (seg *Segmenter) getDag(runes []rune) map[int][]int {
	dag := make(map[int][]int)
	n := len(runes)

	var (
		frag []rune
		i    int
	)

	for k := 0; k < n; k++ {
		dag[k] = make([]int, 0)
		i = k
		frag = runes[k : k+1]

		for {
			freq, ok := seg.Find(string(frag))
			if !ok {
				break
			}

			if freq > 0 {
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

func (seg *Segmenter) calc(runes []rune) map[int]route {
	dag := seg.getDag(runes)

	n := len(runes)
	rs := make(map[int]route)

	rs[n] = route{frequency: 0.0, index: 0}
	var r route

	logT := math.Log(float64(seg.dict.totalFrequency))
	for idx := n - 1; idx >= 0; idx-- {
		for _, i := range dag[idx] {
			freq, ok := seg.Find(string(runes[idx : i+1]))

			if ok {
				f := math.Log(float64(freq)) - logT + rs[i+1].frequency
				r = route{frequency: f, index: i}
			} else {
				f := math.Log(1.0) - logT + rs[i+1].frequency
				r = route{frequency: f, index: i}
			}

			if v, ok := rs[idx]; !ok {
				rs[idx] = r
			} else {
				f := v.frequency == r.frequency && v.index < r.index
				if v.frequency < r.frequency || f {
					rs[idx] = r
				}
			}
		}
	}

	return rs
}

func (seg *Segmenter) hmm(bufString string, buf []rune) (result []string) {

	v, ok := seg.Find(bufString)
	if !ok || v == 0 {
		for _, t := range seg.HMMCut(bufString) {
			result = append(result, t)
		}

		return
	}

	for _, elem := range buf {
		result = append(result, string(elem))
	}
	return
}

func (seg *Segmenter) cutDAG(str string) []string {

	mLen := int(float32(len(str))/RatioWord) + 1
	result := make([]string, 0, mLen)

	runes := []rune(str)
	routes := seg.calc(runes)

	var y int
	length := len(runes)
	var buf []rune

	for x := 0; x < length; {
		y = routes[x].index + 1
		frag := runes[x:y]

		if y-x == 1 {
			buf = append(buf, frag...)
		} else {
			if len(buf) > 0 {
				bufString := string(buf)

				if len(buf) == 1 {
					result = append(result, bufString)
				} else {
					result = append(result, seg.hmm(bufString, buf)...)
				}

				buf = make([]rune, 0)
			}

			result = append(result, string(frag))
		}

		x = y
	}

	if len(buf) > 0 {
		bufString := string(buf)

		if len(buf) == 1 {
			result = append(result, bufString)
		} else {
			result = append(result, seg.hmm(bufString, buf)...)
		}
	}

	return result
}

func (seg *Segmenter) cutDAGNoHMM(str string) []string {
	mLen := int(float32(len(str))/RatioWord) + 1
	result := make([]string, 0, mLen)

	runes := []rune(str)
	routes := seg.calc(runes)
	length := len(runes)

	var y int
	var buf []rune

	for x := 0; x < length; {
		y = routes[x].index + 1
		frag := runes[x:y]

		if reEng.MatchString(string(frag)) && len(frag) == 1 {
			buf = append(buf, frag...)
			x = y
			continue
		}

		if len(buf) > 0 {
			result = append(result, string(buf))
			buf = make([]rune, 0)
		}

		result = append(result, string(frag))
		x = y
	}

	if len(buf) > 0 {
		result = append(result, string(buf))
		// buf = make([]rune, 0)
	}

	return result
}

func (seg *Segmenter) cutAll(str string) []string {
	mLen := int(float32(len(str))/RatioWord) + 1
	result := make([]string, 0, mLen)

	runes := []rune(str)
	dag := seg.getDag(runes)
	start := -1
	ks := make([]int, len(dag))

	for k := range dag {
		ks[k] = k
	}

	var l []int
	for k := range ks {
		l = dag[k]

		if len(l) == 1 && k > start {
			result = append(result, string(runes[k:l[0]+1]))
			start = l[0]
			continue
		}

		for _, j := range l {
			if j > k {
				result = append(result, string(runes[k:j+1]))
				start = j
			}
		}
	}

	return result
}

func (seg *Segmenter) cutForSearch(str string, hmm ...bool) []string {

	mLen := int(float32(len(str))/RatioWordFull) + 1
	result := make([]string, 0, mLen)

	ws := seg.Cut(str, hmm...)
	for _, word := range ws {
		runes := []rune(word)
		for _, incr := range []int{2, 3} {
			if len(runes) <= incr {
				continue
			}

			var gram string
			for i := 0; i < len(runes)-incr+1; i++ {
				gram = string(runes[i : i+incr])
				v, ok := seg.Find(gram)
				if ok && v > 0 {
					result = append(result, gram)
				}
			}
		}

		result = append(result, word)
	}

	return result
}
