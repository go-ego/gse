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

package segment

import (
	"sort"
	"testing"

	"github.com/vcaesar/tt"
)

func TestSegment(t *testing.T) {
	s := Segment{Text: "gse", Weight: 1.5}
	tt.Equal(t, "gse", s.GetText())
	tt.Equal(t, 1.5, s.GetWeight())
}

func TestSegmentsSort(t *testing.T) {
	ss := Segments{
		{Text: "c", Weight: 2},
		{Text: "b", Weight: 1},
		{Text: "a", Weight: 2},
	}
	tt.Equal(t, 3, ss.Len())
	tt.Bool(t, ss.Less(1, 0))
	tt.Bool(t, ss.Less(2, 0))
	tt.Bool(t, !ss.Less(0, 2))

	sort.Sort(ss)
	tt.Equal(t, Segments{
		{Text: "b", Weight: 1},
		{Text: "a", Weight: 2},
		{Text: "c", Weight: 2},
	}, ss)

	sort.Sort(sort.Reverse(ss))
	tt.Equal(t, "c", ss[0].Text)
	tt.Equal(t, "b", ss[2].Text)
}
