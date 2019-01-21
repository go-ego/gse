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

import "math"

type route struct {
	frequency float64
	index     int
}

func (seg *Segmenter) find(str string) (int, bool) {
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
			freq, ok := seg.find(string(frag))
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
			freq, ok := seg.find(string(runes[idx : i+1]))

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

func (seg *Segmenter) hmm(bufString string,
	buf []rune, prob ...map[rune]float64) (result []string) {

	v, ok := seg.find(bufString)
	if !ok || v == 0 {
		for _, t := range seg.HMMCut(bufString, prob...) {
			result = append(result, t)
		}

		return
	}

	for _, elem := range buf {
		result = append(result, string(elem))
	}
	return
}
