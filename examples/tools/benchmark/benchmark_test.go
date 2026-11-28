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

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/examples/tools/internal/segtool"
	"github.com/vcaesar/tt"
)

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write error") }

func TestSegmentLines(t *testing.T) {
	seg := &gse.Segmenter{SkipLog: true}
	tt.Nil(t, seg.LoadDictStr("你好 2000 l\n世界 2000 n\n"))
	m, err := segtool.Get(segtool.Default)
	tt.Nil(t, err)

	var b strings.Builder
	tt.Nil(t, segmentLines(seg, m.Func, []string{"你好世界", "世界"}, 2, &b))
	tt.Equal(t, "你好/l 世界/n\n世界/n\n你好/l 世界/n\n世界/n\n", b.String())

	tt.Nil(t, segmentLines(seg, m.Func, []string{"你好世界"}, 1, nil))
	tt.NotNil(t, segmentLines(seg, m.Func, []string{"你好世界"}, 1, errWriter{}))
}

func TestWriteMemProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mem.prof")
	tt.Nil(t, writeMemProfile(path))
	info, err := os.Stat(path)
	tt.Nil(t, err)
	tt.Bool(t, info.Size() > 0)

	tt.NotNil(t, writeMemProfile(filepath.Join(t.TempDir(), "none", "mem.prof")))
}
