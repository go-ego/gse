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
	"runtime"
	"strconv"
	"testing"

	"github.com/vcaesar/tt"
)

// the tests run on both tries: `go test` and `go test -tags cedar`

func TestTrieInsertGet(t *testing.T) {
	tr := newTrie()
	words := []string{"网", "网球", "网球拍", "拍卖会", "a", "ab", "𦡦", "𦡦𦡦", "\xff", "\xfe"}
	for i, w := range words {
		tt.Nil(t, tr.Insert([]byte(w), i))
	}
	for i, w := range words {
		val, err := tr.Get([]byte(w))
		tt.Nil(t, err)
		tt.Equal(t, i, val)
	}

	tt.NotNil(t, tr.Insert([]byte("x"), -1))

	// a prefix node exists but has no value
	id, err := tr.Jump([]byte("拍卖"), 0)
	tt.Nil(t, err)
	tt.Bool(t, id != 0)
	_, err = tr.Value(id)
	tt.NotNil(t, err)

	for _, w := range []string{"球", "网球拍卖", "abc", "x", "𦡦x", "\xfd"} {
		_, err = tr.Get([]byte(w))
		tt.NotNil(t, err, w)
	}

	tt.Nil(t, tr.Insert([]byte("网球"), 100))
	val, err := tr.Get([]byte("网球"))
	tt.Nil(t, err)
	tt.Equal(t, 100, val)
}

func TestTrieJumpPieces(t *testing.T) {
	tr := newTrie()
	tt.Nil(t, tr.Insert([]byte("to be"), 7))

	id := 0
	var err error
	for _, piece := range []string{"to", " ", "be"} {
		id, err = tr.Jump([]byte(piece), id)
		tt.Nil(t, err)
	}
	val, err := tr.Value(id)
	tt.Nil(t, err)
	tt.Equal(t, 7, val)

	_, err = tr.Jump([]byte("tx"), 0)
	tt.NotNil(t, err)
}

func TestTrieDelete(t *testing.T) {
	tr := newTrie()
	tt.Nil(t, tr.Insert([]byte("西雅图"), 1))
	tt.Nil(t, tr.Insert([]byte("西雅图太空针"), 2))

	tt.Nil(t, tr.Delete([]byte("西雅图太空针")))
	_, err := tr.Jump([]byte("西雅图太"), 0)
	tt.NotNil(t, err)
	val, err := tr.Get([]byte("西雅图"))
	tt.Nil(t, err)
	tt.Equal(t, 1, val)

	tt.NotNil(t, tr.Delete([]byte("东京")))

	tt.Nil(t, tr.Insert([]byte("西雅图太空针"), 3))
	val, err = tr.Get([]byte("西雅图太空针"))
	tt.Nil(t, err)
	tt.Equal(t, 3, val)
}

func TestTrieGrowDelete(t *testing.T) {
	tr := newTrie()
	n := 20000
	for i := 0; i < n; i++ {
		tt.Nil(t, tr.Insert([]byte("词"+strconv.Itoa(i)), i))
	}
	for i := 0; i < n; i += 2 {
		tt.Nil(t, tr.Delete([]byte("词"+strconv.Itoa(i))))
	}

	for i := 0; i < n; i++ {
		val, err := tr.Get([]byte("词" + strconv.Itoa(i)))
		if i%2 == 0 {
			if err == nil {
				t.Fatalf("deleted key %d still found", i)
			}
			continue
		}
		if err != nil || val != i {
			t.Fatalf("key %d: got %d, %v", i, val, err)
		}
	}
}

func heapAlloc() uint64 {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// BenchmarkTrieLoad builds a trie from the loaded zh dictionary
// and reports its retained heap size
func BenchmarkTrieLoad(b *testing.B) {
	keys := make([][]byte, len(prodSeg.Dict.Tokens))
	for i, token := range prodSeg.Dict.Tokens {
		keys[i] = textSliceToBytes(token.text)
	}

	var tr *trie
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		before := heapAlloc()
		b.StartTimer()

		tr = newTrie()
		for id, key := range keys {
			if err := tr.Insert(key, id); err != nil {
				b.Fatal(err)
			}
		}

		b.StopTimer()
		b.ReportMetric(float64(heapAlloc()-before)/(1<<20), "MB-heap")
		b.StartTimer()
	}
	runtime.KeepAlive(tr)
}
