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

// LoadModel load HMM model
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

func internalCut(sentence string) []string {
	result := make([]string, 0, 10)

	runes := []rune(sentence)
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

// Cut cuts sentence into words using HMM with Viterbi algorithm
func Cut(sentence string) []string {
	result := make([]string, 0, 10)

	var (
		hans      string
		hanLoc    []int
		nonHanLoc []int
	)

	for {
		// find(sentence, hans, hanLoc, nonHanLoc)

		hanLoc = regHan.FindStringIndex(sentence)
		if hanLoc == nil {
			if len(sentence) == 0 {
				break
			}
		} else if hanLoc[0] == 0 {
			hans = sentence[hanLoc[0]:hanLoc[1]]
			sentence = sentence[hanLoc[1]:]
			for _, han := range internalCut(hans) {
				result = append(result, han)
			}
			continue
		}

		nonHanLoc = regSkip.FindStringIndex(sentence)
		if nonHanLoc == nil {
			if len(sentence) == 0 {
				break
			}
		} else if nonHanLoc[0] == 0 {
			nonHans := sentence[nonHanLoc[0]:nonHanLoc[1]]
			sentence = sentence[nonHanLoc[1]:]
			if nonHans != "" {
				result = append(result, nonHans)
				continue
			}
		}

		loc := locJudge(sentence, hanLoc, nonHanLoc)
		if loc == nil {
			result = append(result, sentence)
			break
		}

		result = append(result, sentence[:loc[0]])
		sentence = sentence[loc[0]:]
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

	return loc
}
