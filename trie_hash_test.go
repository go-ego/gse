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

	// tombstones are dropped by rehash instead of growing the table
	tt.Equal(t, trieMinSlots, len(tr.keys))
	tt.Equal(t, 2, tr.used)
	val, err := tr.Get([]byte("中国"))
	tt.Nil(t, err)
	tt.Equal(t, 1, val)
	_, err = tr.Jump([]byte("西"), 0)
	tt.NotNil(t, err)
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
}
