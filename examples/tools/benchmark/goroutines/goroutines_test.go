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
	"testing"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/examples/tools/internal/segtool"
	"github.com/vcaesar/tt"
)

func TestSegmentParallel(t *testing.T) {
	seg := &gse.Segmenter{SkipLog: true}
	tt.Nil(t, seg.LoadDictStr("你好 2000 l\n世界 2000 n\n"))
	lines := []string{"你好世界", "世界", "你好, 世界"}

	for _, m := range segtool.Modes {
		var want int64
		for _, l := range lines {
			want += int64(len(m.Func(seg, l)))
		}

		for _, workers := range []int{1, 4} {
			tt.Equal(t, want*3, segmentParallel(seg, m.Func, lines, 3, workers), m.Name)
		}
	}
	tt.Equal(t, int64(0), segmentParallel(seg, segtool.Modes[0].Func, nil, 3, 2))
}
