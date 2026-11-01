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
	"sort"
	"testing"

	"github.com/vcaesar/tt"
)

func TestEdgesSort(t *testing.T) {
	es := edges{
		{start: "a", end: "b", weight: 3},
		{start: "a", end: "c", weight: 1},
		{start: "b", end: "c", weight: 2},
	}
	tt.Equal(t, 3, es.Len())
	tt.Bool(t, es.Less(1, 0))

	sort.Sort(es)
	tt.Equal(t, []float64{1, 2, 3},
		[]float64{es[0].weight, es[1].weight, es[2].weight})
}

func TestGraphAddEdge(t *testing.T) {
	g := newUndirectWeightedGraph()
	g.addEdge("b", "a", 1)
	g.addEdge("b", "c", 2)
	g.addEdge("c", "a", 3)

	tt.Equal(t, 3, len(g.keys))
	tt.Equal(t, 2, len(g.graph["a"]))
	tt.Equal(t, 2, len(g.graph["b"]))
	tt.Equal(t, 2, len(g.graph["c"]))
	tt.Equal(t, edge{start: "a", end: "b", weight: 1}, g.graph["a"][0])
	tt.Equal(t, edge{start: "b", end: "c", weight: 2}, g.graph["b"][1])
}

func TestGraphRank(t *testing.T) {
	tt.Equal(t, 0, len(newUndirectWeightedGraph().rank()))

	g := newUndirectWeightedGraph()
	g.addEdge("c", "a", 1)
	g.addEdge("a", "b", 1)
	g.addEdge("a", "d", 1)

	ws := g.rank()
	tt.Bool(t, sort.IsSorted(g.keys))
	tt.Equal(t, 4, len(ws))
	// the leaves tie, so the hub "a" always ranks first
	tt.Equal(t, "a", ws[0].Text)
	for i := 1; i < len(ws); i++ {
		tt.Bool(t, ws[i].Weight < ws[0].Weight)
	}
}

func TestTextRankerLoadDict(t *testing.T) {
	var tr TextRanker
	tr.WithGse(newPosSeg(t))
	tt.Nil(t, tr.LoadDict(testDict))
	tt.NotNil(t, tr.LoadDict(notExistDic))
}

func TestTextRank(t *testing.T) {
	var tr TextRanker
	tr.WithGse(newPosSeg(t))

	all := tr.TextRank(extText, 0)
	texts := make(map[string]bool)
	for _, s := range all {
		texts[s.Text] = true
	}
	tt.Equal(t, map[string]bool{
		"喜欢": true, "吃": true, "苹果": true, "香蕉": true, "水果": true,
	}, texts)

	// rank order depends on map iteration in rank's min/max search,
	// so only the members are checked
	top := tr.TextRank(extText, 2)
	tt.Equal(t, 2, len(top))
	tt.Bool(t, texts[top[0].Text] && texts[top[1].Text])

	nouns := tr.TextRankWithPOS(extText, 10, []string{"n"})
	tt.Equal(t, 3, len(nouns))
	for _, s := range nouns {
		tt.Bool(t, s.Text == "苹果" || s.Text == "香蕉" || s.Text == "水果")
	}

	tt.Equal(t, 0, len(tr.TextRankWithPOS(extText, 5, nil)))
}
