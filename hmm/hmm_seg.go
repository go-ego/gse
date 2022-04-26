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

/*

Package hmm is the Golang HMM cut module
*/
package hmm

import (
	"regexp"
)

var (
	regHan  = regexp.MustCompile(`\p{Han}+`)
	regSkip = regexp.MustCompile(`(\d+\.\d+|[a-zA-Z0-9]+)`)
)

// func LoadFile(filePath string) map[rune]float64 {
//
// }

// LoadModel load the HMM model
func LoadModel(prob ...map[rune]float64) {
	if len(prob) > 3 {
		probEmit['B'] = prob[0]
		probEmit['E'] = prob[1]
		probEmit['M'] = prob[2]
		probEmit['S'] = prob[3]

		return
	}

	loadDefEmit()
}

func internalCut(text string) []string {
	result := make([]string, 0, 10)

	runes := []rune(text)
	_, posList := Viterbi(runes, []byte{'B', 'M', 'E', 'S'})
	begin, next := 0, 0

	for i, char := range runes {
		pos := posList[i]
		switch pos {
		case 'B':
			begin = i
		case 'E':
			result = append(result, string(runes[begin:i+1]))
			next = i + 1
		case 'S':
			result = append(result, string(char))
			next = i + 1
		}
	}

	if next < len(runes) {
		result = append(result, string(runes[next:]))
	}

	return result
}

// Cut cuts text to words using HMM with Viterbi algorithm
func Cut(text string, reg ...*regexp.Regexp) []string {
	result := make([]string, 0, 10)

	var (
		cuts      string
		cutLoc    []int
		nonCutLoc []int
	)

	for {
		// find(text, cuts, cutLoc, nonCutLoc)
		if len(reg) > 1 {
			cutLoc = reg[1].FindStringIndex(text)
		} else {
			cutLoc = regHan.FindStringIndex(text)
		}

		if cutLoc == nil {
			if len(text) == 0 {
				break
			}
		} else if cutLoc[0] == 0 {
			cuts = text[cutLoc[0]:cutLoc[1]]
			text = text[cutLoc[1]:]
			result = append(result, internalCut(cuts)...)
			continue
		}

		if len(reg) > 0 {
			nonCutLoc = reg[0].FindStringIndex(text)
		} else {
			nonCutLoc = regSkip.FindStringIndex(text)
		}
		if nonCutLoc == nil {
			if len(text) == 0 {
				break
			}
		} else if nonCutLoc[0] == 0 {
			nonCuts := text[nonCutLoc[0]:nonCutLoc[1]]
			text = text[nonCutLoc[1]:]
			if nonCuts != "" {
				result = append(result, nonCuts)
				continue
			}
		}

		loc := locJudge(text, cutLoc, nonCutLoc)
		if loc == nil {
			result = append(result, text)
			break
		}

		result = append(result, text[:loc[0]])
		text = text[loc[0]:]
	}

	return result
}

func locJudge(str string, cutLoc, nonCutLoc []int) (loc []int) {
	if cutLoc == nil && nonCutLoc == nil {
		if len(str) > 0 {
			return nil
		}
	} else if cutLoc == nil {
		loc = nonCutLoc
	} else if nonCutLoc == nil || cutLoc[0] < nonCutLoc[0] {
		loc = cutLoc
	} else {
		loc = nonCutLoc
	}

	return
}
