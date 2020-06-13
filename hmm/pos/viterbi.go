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
	"fmt"
	"sort"
)

type probState struct {
	prob  float64
	state uint16
}

func (ps probState) String() string {
	return fmt.Sprintf("(%v: %f)", ps.state, ps.prob)
}

type probStates []probState

func (pss probStates) Len() int {
	return len(pss)
}

func (pss probStates) Less(i, j int) bool {
	if pss[i].prob == pss[j].prob {
		return pss[i].state < pss[j].state
	}

	return pss[i].prob < pss[j].prob
}

func (pss probStates) Swap(i, j int) {
	pss[i], pss[j] = pss[j], pss[i]
}

func viterbi(obs []rune) []tag {
	obsLen := len(obs)
	vtb := make([]map[uint16]float64, obsLen)
	vtb[0] = make(map[uint16]float64)
	memPath := make([]map[uint16]uint16, obsLen)
	memPath[0] = make(map[uint16]uint16)

	ys := charStateTab.get(obs[0]) // default is all_states
	for _, y := range ys {
		vtb[0][y] = probEmit[y].get(obs[0]) + probStart[y]
		memPath[0][y] = 0
	}

	memPath, vtb = probs(obs, vtb, memPath, obsLen)

	last := make(probStates, 0)
	length := len(memPath)
	vlength := len(vtb)
	for y := range memPath[length-1] {
		ps := probState{prob: vtb[vlength-1][y], state: y}
		last = append(last, ps)
	}

	sort.Sort(sort.Reverse(last))
	state := last[0].state
	route := make([]tag, len(obs))

	for i := obsLen - 1; i >= 0; i-- {
		route[i] = tag(state)
		state = memPath[i][state]
	}

	return route
}

func probs(obs []rune, vtb []map[uint16]float64, memPath []map[uint16]uint16,
	obsLen int) ([]map[uint16]uint16, []map[uint16]float64) {
	for t := 1; t < obsLen; t++ {
		var prevStates []uint16
		for x := range memPath[t-1] {
			if len(probTrans[x]) > 0 {
				prevStates = append(prevStates, x)
			}
		}

		prevStatesExpectNext := make(map[uint16]int)
		for _, x := range prevStates {
			for y := range probTrans[x] {
				prevStatesExpectNext[y] = 1
			}
		}
		tmpObsStates := charStateTab.get(obs[t])

		var obsStates []uint16
		for index := range tmpObsStates {
			if _, ok := prevStatesExpectNext[tmpObsStates[index]]; ok {
				obsStates = append(obsStates, tmpObsStates[index])
			}
		}

		if len(obsStates) == 0 {
			for key := range prevStatesExpectNext {
				obsStates = append(obsStates, key)
			}
		}

		if len(obsStates) == 0 {
			obsStates = probTransKeys
		}

		memPath[t] = make(map[uint16]uint16)
		vtb[t] = make(map[uint16]float64)
		for _, y := range obsStates {
			var max, ps probState
			for i, y0 := range prevStates {
				ps = probState{
					prob:  vtb[t-1][y0] + probTrans[y0].Get(y) + probEmit[y].get(obs[t]),
					state: y0}

				if i == 0 || ps.prob > max.prob || (ps.prob == max.prob && ps.state > max.state) {
					max = ps
				}
			}

			vtb[t][y] = max.prob
			memPath[t][y] = max.state
		}
	}

	return memPath, vtb
}
