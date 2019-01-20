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
