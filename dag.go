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
	"bytes"
	"math"
	"regexp"
	"strings"
)

const (
	// RatioWord ratio words and letters
	RatioWord float32 = 1.5
	// RatioWordFull full ratio words and letters
	RatioWordFull float32 = 1
)

var reEng = regexp.MustCompile(`[[:alnum:]]`)

type route struct {
	freq  float64
	index int
}

// Find find word in dictionary return word's freq, pos and existence
func (seg *Segmenter) Find(str string) (float64, string, bool) {
	return seg.Dict.Find([]byte(str))
}

// FindTFIDF find word in dictionary return word's freq, inverseFreq and existence
func (seg *Segmenter) FindTFIDF(str string) (float64, float64, bool) {
	return seg.Dict.FindTFIDF([]byte(str))
}

// Value find word in dictionary return word's value
func (seg *Segmenter) Value(str string) (int, int, error) {
	return seg.Dict.Value([]byte(str))
}

// FindAllOccs find the all search byte start in data
func FindAllOccs(data []byte, searches []string) map[string][]int {
	results := make(map[string][]int, 0)
	tmp := data
	for _, search := range searches {
		index := len(data)
		for {
			match := bytes.LastIndex(tmp[0:index], []byte(search))
			if match == -1 {
				break
			}

			index = match
			results[search] = append(results[search], match)
		}
	}

	return results
}

// Analyze analyze the token segment info
func (seg *Segmenter) Analyze(text []string, t1 string, by ...bool) (az []AnalyzeToken) {
	if len(text) <= 0 {
		return
	}

	start, end := 0, 0
	if t1 == "" {
		if len(by) > 0 {
			end = len([]rune(text[0]))
		} else {
			end = len([]byte(text[0]))
		}
	}

	isEx := make(map[string]int, 0)
	if ToLower {
		t1 = strings.ToLower(t1)
	}
	all := FindAllOccs([]byte(t1), text)
	for k, v := range text {
		if k > 0 && t1 == "" {
			start = az[k-1].End
			if len(by) > 0 {
				end = az[k-1].End + len([]rune(v))
			} else {
				end = az[k-1].End + len([]byte(v))
			}
		}

		if t1 != "" {
			if _, ok := isEx[v]; ok {
				isEx[v]++
			} else {
				isEx[v] = 0
			}

			if len(all[v]) > 0 {
				start = all[v][isEx[v]]
			}
			end = start + len([]byte(v))
		}

		freq, pos, _ := seg.Find(v)
		az = append(az, AnalyzeToken{
			Position: k,
			Start:    start,
			End:      end,

			Text: v,
			Freq: freq,
			Pos:  pos,
		})
	}

	return
}

// getDag get a directed acyclic graph (DAG) from slice of runes(containing Unicode characters)
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
			freq, _, ok := seg.Find(string(frag))
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

	rs[n] = route{freq: 0.0, index: 0}
	var r route

	logT := math.Log(seg.Dict.totalFreq)
	for idx := n - 1; idx >= 0; idx-- {
		for _, i := range dag[idx] {
			freq, _, ok := seg.Find(string(runes[idx : i+1]))

			if ok {
				f := math.Log(freq) - logT + rs[i+1].freq
				r = route{freq: f, index: i}
			} else {
				f := math.Log(1.0) - logT + rs[i+1].freq
				r = route{freq: f, index: i}
			}

			if v, ok := rs[idx]; !ok {
				rs[idx] = r
			} else {
				f := v.freq == r.freq && v.index < r.index
				if v.freq < r.freq || f {
					rs[idx] = r
				}
			}
		}
	}

	return rs
}

func (seg *Segmenter) hmm(bufString string, buf []rune, reg ...*regexp.Regexp) (result []string) {

	v, _, ok := seg.Find(bufString)
	if !ok || v == 0 {
		result = append(result, seg.HMMCut(bufString, reg...)...)
		return
	}

	for _, elem := range buf {
		result = append(result, string(elem))
	}
	return
}

func (seg *Segmenter) cutDAG(str string, reg ...*regexp.Regexp) []string {

	mLen := int(float32(len(str))/RatioWord) + 1
	result := make([]string, 0, mLen)

	if ToLower {
		str = strings.ToLower(str)
	}
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
					result = append(result, seg.hmm(bufString, buf, reg...)...)
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
			result = append(result, seg.hmm(bufString, buf, reg...)...)
		}
	}

	return result
}

func (seg *Segmenter) cutDAGNoHMM(str string) []string {
	mLen := int(float32(len(str))/RatioWord) + 1
	result := make([]string, 0, mLen)

	if ToLower {
		str = strings.ToLower(str)
	}
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

	if ToLower {
		str = strings.ToLower(str)
	}
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
				v, _, ok := seg.Find(gram)
				if ok && v > 0 {
					result = append(result, gram)
				}
			}
		}

		result = append(result, word)
	}

	return result
}

// SuggestFreq suggest the words frequency
// return a suggested frequency of a word cutted to short words.
func (seg *Segmenter) SuggestFreq(words ...string) float64 {
	freq := 1.0
	total := seg.Dict.totalFreq

	if len(words) > 1 {
		for _, word := range words {
			v, _, ok := seg.Find(word)
			if ok {
				freq *= v
			}

			freq /= total
		}

		freq, _ = math.Modf(freq * total)
		wordFreq := 0.0
		v, _, ok := seg.Find(strings.Join(words, ""))
		if ok {
			wordFreq = v
		}

		if wordFreq < freq {
			freq = wordFreq
		}

		return freq
	}

	word := words[0]
	for _, segment := range seg.Cut(word, false) {
		v, _, ok := seg.Find(segment)
		if ok {
			freq *= v
		}

		freq /= total
	}

	freq, _ = math.Modf(freq * total)
	freq += 1.0
	wordFreq := 1.0

	v, _, ok := seg.Find(word)
	if ok {
		wordFreq = v
	}

	if wordFreq > freq {
		freq = wordFreq
	}

	return freq
}
