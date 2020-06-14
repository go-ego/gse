package idf

import (
	"math"
	"sort"

	"github.com/go-ego/gse/hmm/pos"
)

const dampingFactor = 0.85

var (
	defaultAllowPOS = []string{"ns", "n", "vn", "v"}
)

type edge struct {
	start  string
	end    string
	weight float64
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
	} else {
		u.graph[end] = append(u.graph[end],
			edge{start: end, end: start, weight: weight})
	}
}

func (u *undirectWeightedGraph) rank() Segments {
	if !sort.IsSorted(u.keys) {
		sort.Sort(u.keys)
	}

	ws := make(map[string]float64)
	outSum := make(map[string]float64)

	wsdef := 1.0
	if len(u.graph) > 0 {
		wsdef /= float64(len(u.graph))
	}

	for n, out := range u.graph {
		ws[n] = wsdef
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

	result := make(Segments, 0)
	for n, w := range ws {
		result = append(result,
			Segment{text: n,
				weight: (w - minRank/10.0) / (maxRank - minRank/10.0)})
	}

	sort.Sort(sort.Reverse(result))
	return result
}

// TextRankWithPOS extracts keywords from sentence using TextRank algorithm.
// Parameter allowPOS allows a customized pos list.
func (t *TextRanker) TextRankWithPOS(sentence string, topK int, allowPOS []string) Segments {
	posFilt := make(map[string]int)
	for _, pos1 := range allowPOS {
		posFilt[pos1] = 1
	}

	g := newUndirectWeightedGraph()
	cm := make(map[[2]string]float64)
	span := 5

	var pairs []pos.Segment
	for _, pair := range t.seg.Cut(sentence, true) {
		pairs = append(pairs, pair)
	}

	for i := range pairs {
		_, ok := posFilt[pairs[i].Pos]
		if ok {
			for j := i + 1; j < i+span && j <= len(pairs); j++ {
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

// TextRank extract keywords from sentence using TextRank algorithm.
// Parameter topK specify how many top keywords to be returned at most.
func (t *TextRanker) TextRank(sentence string, topK int) Segments {
	return t.TextRankWithPOS(sentence, topK, defaultAllowPOS)
}

// TextRanker is used to extract tags from sentence.
type TextRanker struct {
	seg *pos.Segmenter
}

// LoadDict reads a given file and create a new dictionary file for Textranker.
func (t *TextRanker) LoadDict(fileName ...string) error {
	t.seg = new(pos.Segmenter)
	return t.seg.LoadDict(fileName...)
}
