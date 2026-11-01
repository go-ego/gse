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

//go:build !cedar

package gse

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/vcaesar/tt"
)

func TestTrieChurn(t *testing.T) {
	tr := newTrie()
	tt.Nil(t, tr.Insert([]byte("中国"), 1))
	for i := 0; i < 10000; i++ {
		tt.Nil(t, tr.Insert([]byte("西雅图太空针"), i))
		tt.Nil(t, tr.Delete([]byte("西雅图太空针")))
	}

	// the deleted slots are reused instead of growing the table
	tt.Equal(t, trieMinSlots, len(tr.keys))
	tt.Equal(t, 2, tr.used)
	val, err := tr.Get([]byte("中国"))
	tt.Nil(t, err)
	tt.Equal(t, 1, val)
	_, err = tr.Jump([]byte("西"), 0)
	tt.NotNil(t, err)

	// distinct keys leave tombstones, compacted by a same size rehash
	for i := 0; i < 10000; i++ {
		key := []byte("词" + strconv.Itoa(i))
		tt.Nil(t, tr.Insert(key, i))
		tt.Nil(t, tr.Delete(key))
	}
	tt.Equal(t, trieMinSlots, len(tr.keys))
	tt.Equal(t, 2, tr.used)
}

func TestTrieReinsert(t *testing.T) {
	tr := newTrie()
	key := []byte(strings.Repeat("长", 200))
	tt.Nil(t, tr.Insert(key, 1))
	size := len(tr.keys)
	id, err := tr.Jump(key, 0)
	tt.Nil(t, err)

	// an existing key needs no room and keeps its node id
	tt.Nil(t, tr.Insert(key, 2))
	tt.Equal(t, size, len(tr.keys))
	id2, err := tr.Jump(key, 0)
	tt.Nil(t, err)
	tt.Equal(t, id, id2)
	val, err := tr.Value(id2)
	tt.Nil(t, err)
	tt.Equal(t, 2, val)
}

func TestTrieDeepRehash(t *testing.T) {
	tr := newTrie()
	key := []byte(strings.Repeat("深", 100000))
	tt.Nil(t, tr.Insert(key, 1))
	tt.Nil(t, tr.Insert(key[:3], 2))
	tr.rehash(len(tr.keys) * 2)

	val, err := tr.Get(key)
	tt.Nil(t, err)
	tt.Equal(t, 1, val)
	tt.Nil(t, tr.Delete(key))
	tt.Equal(t, 1, tr.used)
}

func TestTrieValueEmptySlot(t *testing.T) {
	tr := newTrie()
	tt.Nil(t, tr.Insert([]byte("网"), 0))
	id, err := tr.Jump([]byte("网"), 0)
	tt.Nil(t, err)

	empty := id%len(tr.keys) + 1
	_, err = tr.Value(empty)
	tt.NotNil(t, err)
}

// TestTrieModel checks random inserts and deletes against a map,
// with the children counts and node pruning
func TestTrieModel(t *testing.T) {
	alphabet := []string{"大", "小", "a", "𦡦", "\xff"}
	rnd := rand.New(rand.NewSource(1))
	randKey := func() string {
		var b strings.Builder
		for n := rnd.Intn(4); n >= 0; n-- {
			b.WriteString(alphabet[rnd.Intn(len(alphabet))])
		}
		return b.String()
	}

	tr, model := newTrie(), map[string]int{}
	for i := 0; i < 50000; i++ {
		key := randKey()
		if rnd.Intn(3) == 0 {
			_, ok := model[key]
			err := tr.Delete([]byte(key))
			tt.Equal(t, ok, err == nil)
			delete(model, key)
			continue
		}
		tt.Nil(t, tr.Insert([]byte(key), i))
		model[key] = i
	}

	// every node is a prefix of a key, so used counts the distinct prefixes
	nodes := map[string]bool{}
	for key, val := range model {
		got, err := tr.Get([]byte(key))
		tt.Nil(t, err)
		tt.Equal(t, val, got)
		for pos := 0; pos < len(key); {
			_, n := trieRune([]byte(key[pos:]))
			pos += n
			nodes[key[:pos]] = true
		}
	}
	tt.Equal(t, len(nodes), tr.used)

	for key := range model {
		tt.Nil(t, tr.Delete([]byte(key)))
	}
	tt.Equal(t, 0, tr.used)
	tt.Equal(t, 0, len(tr.more))
}

func TestTrieKidsSaturate(t *testing.T) {
	tr := newTrie()
	n := trieKidsMax + 100
	for i := 0; i < n; i++ {
		tt.Nil(t, tr.Insert([]byte("大"+string(rune(0x4E00+i))), i))
	}
	tt.Nil(t, tr.Insert([]byte("大小"), n))

	id, err := tr.Jump([]byte("大"), 0)
	tt.Nil(t, err)
	tt.Equal(t, uint64(trieKidsMask), tr.keys[id-1]&trieKidsMask)

	for i := 0; i < n; i++ {
		tt.Nil(t, tr.Delete([]byte("大"+string(rune(0x4E00+i)))))
	}
	val, err := tr.Get([]byte("大小"))
	tt.Nil(t, err)
	tt.Equal(t, n, val)
	_, err = tr.Get([]byte("大一"))
	tt.NotNil(t, err)

	// a saturated node is pruned once all its children are deleted
	tt.Nil(t, tr.Delete([]byte("大小")))
	_, err = tr.Jump([]byte("大"), 0)
	tt.NotNil(t, err)
	tt.Equal(t, 0, len(tr.more))
	tt.Equal(t, 0, tr.used)
}

func TestTrieKidsRehash(t *testing.T) {
	tr := newTrie()
	n := trieKidsMax + 10
	for i := 0; i < n; i++ {
		tt.Nil(t, tr.Insert([]byte("大"+string(rune(0x4E00+i))), i))
	}
	size := len(tr.keys)
	// grow the table so the overflow counts are moved to new node ids
	for i := 0; i < 4*size; i++ {
		tt.Nil(t, tr.Insert([]byte(string(rune(0x10000+i))), i))
	}
	tt.Bool(t, len(tr.keys) > size)

	id, err := tr.Jump([]byte("大"), 0)
	tt.Nil(t, err)
	tt.Equal(t, 10, tr.more[uint32(id)])

	for i := 0; i < n; i++ {
		tt.Nil(t, tr.Delete([]byte("大"+string(rune(0x4E00+i)))))
	}
	_, err = tr.Jump([]byte("大"), 0)
	tt.NotNil(t, err)
	tt.Equal(t, 0, len(tr.more))
}

func TestTrieTombReuse(t *testing.T) {
	tr := newTrie()
	tt.Nil(t, tr.Insert([]byte("西雅图太空针"), 1))
	tt.Nil(t, tr.Delete([]byte("西雅图太空针")))
	tt.Equal(t, 6, tr.tombs)

	// deleted slots are reused, not left to lengthen the probes
	tt.Nil(t, tr.Insert([]byte("西雅图太空针"), 2))
	tt.Equal(t, 0, tr.tombs)
	tt.Equal(t, 6, tr.used)
	val, err := tr.Get([]byte("西雅图太空针"))
	tt.Nil(t, err)
	tt.Equal(t, 2, val)
}

func TestDictSaturatedPrefix(t *testing.T) {
	dict := NewDict()
	n := trieKidsMax + 1
	for i := 0; i < n; i++ {
		tok := Token{text: []Text{Text("大"), Text(string(rune(0x4E00 + i)))}, freq: 1}
		tt.Nil(t, dict.AddToken(tok))
	}
	for i := 0; i < n; i++ {
		tok := Token{text: []Text{Text("大"), Text(string(rune(0x4E00 + i)))}}
		tt.Nil(t, dict.RemoveToken(tok))
	}

	_, _, ok := dict.Find([]byte("大"))
	tt.False(t, ok)
}
