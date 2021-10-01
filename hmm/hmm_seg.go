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
func Cut(text string) []string {
	result := make([]string, 0, 10)

	var (
		hans      string
		hanLoc    []int
		nonHanLoc []int
	)

	for {
		// find(text, hans, hanLoc, nonHanLoc)

		hanLoc = regHan.FindStringIndex(text)
		if hanLoc == nil {
			if len(text) == 0 {
				break
			}
		} else if hanLoc[0] == 0 {
			hans = text[hanLoc[0]:hanLoc[1]]
			text = text[hanLoc[1]:]
			result = append(result, internalCut(hans)...)
			continue
		}

		nonHanLoc = regSkip.FindStringIndex(text)
		if nonHanLoc == nil {
			if len(text) == 0 {
				break
			}
		} else if nonHanLoc[0] == 0 {
			nonHans := text[nonHanLoc[0]:nonHanLoc[1]]
			text = text[nonHanLoc[1]:]
			if nonHans != "" {
				result = append(result, nonHans)
				continue
			}
		}

		loc := locJudge(text, hanLoc, nonHanLoc)
		if loc == nil {
			result = append(result, text)
			break
		}

		result = append(result, text[:loc[0]])
		text = text[loc[0]:]
	}

	return result
}

func locJudge(str string, hanLoc, nonHanLoc []int) (loc []int) {
	if hanLoc == nil && nonHanLoc == nil {
		if len(str) > 0 {
			return nil
		}
	} else if hanLoc == nil {
		loc = nonHanLoc
	} else if nonHanLoc == nil || hanLoc[0] < nonHanLoc[0] {
		loc = hanLoc
	} else {
		loc = nonHanLoc
	}

	return
}
