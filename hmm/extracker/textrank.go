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

package extracker

import (
	"math"
	"sort"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/pos"
	"github.com/go-ego/gse/hmm/segment"
)

const dampingFactor = 0.85

var (
	defaultAllowPOS = []string{"ns", "n", "vn", "v"}
)

// TextRanker is extract tags struct.
type TextRanker struct {
	seg pos.Segmenter
	HMM bool
}

// WithGse register the gse segmenter
func (t *TextRanker) WithGse(segs gse.Segmenter) {
	t.seg.WithGse(segs)
}

// LoadDict load and create a new dictionary from the file for Textranker
func (t *TextRanker) LoadDict(fileName ...string) error {
	// t.seg = new(pos.Segmenter)
	return t.seg.LoadDict(fileName...)
}

type edge struct {
	start, end string
	weight     float64
}

type edges []edge

func (es edges) Len() int {
	return len(es)
}

func (es edges) Less(i, j int) bool {
	return es[i].weight < es[j].weight
}

func (es edges) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}

type undirectWeightedGraph struct {
	graph map[string]edges
	keys  sort.StringSlice
}

func newUndirectWeightedGraph() *undirectWeightedGraph {
	u := new(undirectWeightedGraph)
	u.graph = make(map[string]edges)
	u.keys = make(sort.StringSlice, 0)
	return u
}

func (u *undirectWeightedGraph) addEdge(start, end string, weight float64) {
	// # use a tuple (start, end, weight) instead of a Edge object
	if _, ok := u.graph[start]; !ok {
		u.keys = append(u.keys, start)
		u.graph[start] = edges{edge{start: start, end: end, weight: weight}}
	} else {
		u.graph[start] = append(u.graph[start],
			edge{start: start, end: end, weight: weight})
	}

	if _, ok := u.graph[end]; !ok {
		u.keys = append(u.keys, end)
		u.graph[end] = edges{edge{start: end, end: start, weight: weight}}
		return
	}
	u.graph[end] = append(u.graph[end],
		edge{start: end, end: start, weight: weight})
}

func (u *undirectWeightedGraph) rank() segment.Segments {
	if !sort.IsSorted(u.keys) {
		sort.Sort(u.keys)
	}

	ws := make(map[string]float64)
	outSum := make(map[string]float64)

	wsDef := 1.0
	if len(u.graph) > 0 {
		wsDef /= float64(len(u.graph))
	}

	for n, out := range u.graph {
		ws[n] = wsDef
		sum := 0.0
		for _, e := range out {
			sum += e.weight
		}
		outSum[n] = sum
	}

	for x := 0; x < 10; x++ {
		for _, n := range u.keys {
			s := 0.0
			inedges := u.graph[n]
			for _, e := range inedges {
				s += e.weight / outSum[e.end] * ws[e.end]
			}
			ws[n] = (1 - dampingFactor) + dampingFactor*s
		}
	}

	minRank := math.MaxFloat64
	maxRank := math.SmallestNonzeroFloat64
	for _, w := range ws {
		if w < minRank {
			minRank = w
		} else if w > maxRank {
			maxRank = w
		}
	}

	result := make(segment.Segments, 0)
	for n, w := range ws {
		result = append(result,
			segment.Segment{Text: n, Weight: (w - minRank/10.0) / (maxRank - minRank/10.0)},
		)
	}

	sort.Sort(sort.Reverse(result))
	return result
}

// TextRankWithPOS extracts keywords from text using TextRank algorithm.
// Parameter allowPOS allows a []string pos list.
func (t *TextRanker) TextRankWithPOS(text string, topK int, allowPOS []string) segment.Segments {
	posFilt := make(map[string]int)
	for _, pos1 := range allowPOS {
		posFilt[pos1] = 1
	}

	g := newUndirectWeightedGraph()
	cm := make(map[[2]string]float64)
	span := 5

	var pairs []gse.SegPos
	pairs = append(pairs, t.seg.Cut(text, true)...)

	for i := range pairs {
		_, ok := posFilt[pairs[i].Pos]
		if ok {
			for j := i + 1; j < i+span && j < len(pairs); j++ {
				if _, ok := posFilt[pairs[j].Pos]; !ok {
					continue
				}

				if _, ok := cm[[2]string{pairs[i].Text, pairs[j].Text}]; !ok {
					cm[[2]string{pairs[i].Text, pairs[j].Text}] = 1.0
				} else {
					cm[[2]string{pairs[i].Text, pairs[j].Text}] += 1.0
				}
			}
		}
	}

	for startEnd, weight := range cm {
		g.addEdge(startEnd[0], startEnd[1], weight)
	}

	tags := g.rank()
	if topK > 0 && len(tags) > topK {
		tags = tags[:topK]
	}

	return tags
}

// TextRank extract keywords from text using TextRank algorithm.
// Parameter topK specify how many top keywords to be returned at most.
func (t *TextRanker) TextRank(text string, topK int) segment.Segments {
	return t.TextRankWithPOS(text, topK, defaultAllowPOS)
}
