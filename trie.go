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
	"errors"
	"math"
	"math/bits"
	"unicode/utf8"
)

const (
	trieRootLen  = 1 << 16 // root children of BMP runes are cached directly
	trieMinSlots = 1 << 8
	trieNoVal    = -1

	// slot key layout: used(1) | kids(10) | parent(32) | rune(21)
	trieRuneBits   = 21
	trieParentBits = 32
	trieKidsShift  = trieRuneBits + trieParentBits
	trieRuneMask   = 1<<trieRuneBits - 1
	trieKidsMax    = 1<<10 - 1 // saturated counts are never decremented
	trieKidsMask   = trieKidsMax << trieKidsShift
	trieUsed       = 1 << 63
	trieTomb       = math.MaxUint64 // a deleted slot, matches no key
)

var (
	errTrieNoKey      = errors.New("gse: trie not have key")
	errTrieNoVal      = errors.New("gse: trie not have val")
	errTrieInvalidVal = errors.New("gse: trie invalid val")
)

// trie is a rune-level hash trie, used instead of the byte-level cedar
// double array trie (build with `-tags cedar` to switch back).
//
// A CJK word costs one node per rune instead of one per byte. Every edge
// (parent, rune) is one slot of an open addressing table, and the node it
// leads to is identified by that slot (id = index + 1, the root is 0), so
// a node costs 12 bytes: its key, holding its children count, and its value.
type trie struct {
	keys  []uint64
	vals  []int32
	root  []uint32 // node ids of the root children of BMP runes
	val   int32    // value of the root, the empty key
	used  int
	tombs int
	shift uint // 64 - log2(len(keys))
}

func newTrie() *trie {
	t := &trie{root: make([]uint32, trieRootLen), val: trieNoVal}
	t.rehash(trieMinSlots)
	return t
}

// trieRune decodes the next rune of key, mapping each invalid
// byte to its own label above utf8.MaxRune so no key is lost
func trieRune(key []byte) (rune, int) {
	r, n := utf8.DecodeRune(key)
	if r == utf8.RuneError && n == 1 {
		return utf8.MaxRune + 1 + rune(key[0]), 1
	}
	return r, n
}

func trieKey(parent uint32, r rune) uint64 {
	return trieUsed | uint64(parent)<<trieRuneBits | uint64(r)
}

func trieParent(k uint64) uint32 {
	return uint32(k >> trieRuneBits)
}

func (t *trie) home(k uint64) int {
	return int((k * 0x9E3779B97F4A7C15) >> t.shift)
}

// slot returns the slot index of k, or of the empty slot ending its probe
func (t *trie) slot(k uint64) int {
	mask := len(t.keys) - 1
	i := t.home(k)
	for t.keys[i] != 0 && t.keys[i]&^trieKidsMask != k {
		i = (i + 1) & mask
	}
	return i
}

func (t *trie) child(from uint32, r rune) uint32 {
	if from == 0 && r < trieRootLen {
		return t.root[r]
	}

	i := t.slot(trieKey(from, r))
	if t.keys[i] == 0 {
		return 0
	}
	return uint32(i + 1)
}

// Jump follows key from the node `from`, returns the reached node;
// on a missing edge it returns the last node reached and errTrieNoKey
func (t *trie) Jump(key []byte, from int) (int, error) {
	node := uint32(from)
	for len(key) > 0 {
		r, n := trieRune(key)
		next := t.child(node, r)
		if next == 0 {
			return int(node), errTrieNoKey
		}
		node, key = next, key[n:]
	}
	return int(node), nil
}

func (t *trie) value(id int) *int32 {
	if id == 0 {
		return &t.val
	}
	return &t.vals[id-1]
}

// Value returns the value of the node id
func (t *trie) Value(id int) (int, error) {
	if id < 0 || id > len(t.vals) || *t.value(id) == trieNoVal {
		return 0, errTrieNoVal
	}
	return int(*t.value(id)), nil
}

// Get returns the value of key
func (t *trie) Get(key []byte) (int, error) {
	id, err := t.Jump(key, 0)
	if err != nil {
		return 0, err
	}
	return t.Value(id)
}

// Insert sets the value of key, creating its nodes as needed
func (t *trie) Insert(key []byte, val int) error {
	if val < 0 || val > math.MaxInt32 {
		return errTrieInvalidVal
	}
	// rehash changes node ids, so make room before walking
	t.reserve(utf8.RuneCount(key))

	node := uint32(0)
	for len(key) > 0 {
		r, n := trieRune(key)
		next := t.child(node, r)
		if next == 0 {
			next = t.link(node, r)
		}
		node, key = next, key[n:]
	}

	*t.value(int(node)) = int32(val)
	return nil
}

// link adds a new child node of parent on r
func (t *trie) link(parent uint32, r rune) uint32 {
	if parent != 0 && t.keys[parent-1]&trieKidsMask != trieKidsMask {
		t.keys[parent-1] += 1 << trieKidsShift
	}

	k := trieKey(parent, r)
	i := t.slot(k)
	t.keys[i], t.vals[i] = k, trieNoVal
	t.used++

	id := uint32(i + 1)
	if parent == 0 && r < trieRootLen {
		t.root[r] = id
	}
	return id
}

// reserve makes room for n more slots under the 7/8 load limit
func (t *trie) reserve(n int) {
	size := len(t.keys)
	if (t.used+t.tombs+n)*8 <= size*7 {
		return
	}
	for (t.used+n)*4 > size*3 {
		size *= 2
	}
	t.rehash(size)
}

// rehash moves the slots to a new table of size n, dropping
// the tombstones and remapping the parent ids of the keys
func (t *trie) rehash(n int) {
	keys, vals := t.keys, t.vals
	t.keys, t.vals = make([]uint64, n), make([]int32, n)
	t.shift = uint(64 - bits.Len(uint(n-1)))
	t.used, t.tombs = 0, 0
	for i := range t.root {
		t.root[i] = 0
	}

	ids := make([]uint32, len(keys)) // old slot index -> new node id
	var move func(i int) uint32
	move = func(i int) uint32 {
		if ids[i] != 0 {
			return ids[i]
		}

		parent := trieParent(keys[i])
		if parent != 0 {
			parent = move(int(parent - 1))
		}
		r := rune(keys[i] & trieRuneMask)
		k := trieKey(parent, r)
		j := t.slot(k)
		t.keys[j], t.vals[j] = k|keys[i]&trieKidsMask, vals[i]
		t.used++

		ids[i] = uint32(j + 1)
		if parent == 0 && r < trieRootLen {
			t.root[r] = ids[i]
		}
		return ids[i]
	}

	for i, k := range keys {
		if k != 0 && k != trieTomb {
			move(i)
		}
	}
}

// Delete removes the value of key and prunes the nodes left without
// value and children
func (t *trie) Delete(key []byte) error {
	var path []uint32
	node := uint32(0)
	for len(key) > 0 {
		r, n := trieRune(key)
		next := t.child(node, r)
		if next == 0 {
			return errTrieNoKey
		}
		path = append(path, next)
		node, key = next, key[n:]
	}
	if *t.value(int(node)) == trieNoVal {
		return errTrieNoKey
	}

	*t.value(int(node)) = trieNoVal
	for i := len(path) - 1; i >= 0; i-- {
		k := t.keys[path[i]-1]
		if t.vals[path[i]-1] != trieNoVal || k&trieKidsMask != 0 {
			break
		}

		t.keys[path[i]-1] = trieTomb
		t.used--
		t.tombs++

		parent := trieParent(k)
		r := rune(k & trieRuneMask)
		if parent == 0 && r < trieRootLen {
			t.root[r] = 0
		}
		if parent != 0 && t.keys[parent-1]&trieKidsMask != trieKidsMask {
			t.keys[parent-1] -= 1 << trieKidsShift
		}
	}
	return nil
}
