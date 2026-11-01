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

package util

import (
	"regexp"
	"testing"

	"github.com/vcaesar/tt"
)

func TestRegexpSplit(t *testing.T) {
	tests := []struct {
		re   string
		s    string
		n    int
		want []string
	}{
		{",", "a,b,c", 0, nil},
		{",", "", -1, []string{""}},
		{"", "", -1, []string{}},
		{",", "abc", -1, []string{"abc"}},
		{",", "a,b,c", -1, []string{"a", "b", "c"}},
		{",", "a,b,c", 2, []string{"a", "b,c"}},
		{",", ",a,", -1, []string{"", "a", ""}},
		{"(,)", "a,b,c", -1, []string{"a", ",", "b", ",", "c"}},
		{"(,)", "a,b,c", 2, []string{"a", ",", "b,c"}},
		{"x*", "ab", -1, []string{"a", "b"}},
	}

	for _, tc := range tests {
		got := RegexpSplit(regexp.MustCompile(tc.re), tc.s, tc.n)
		tt.Equal(t, tc.want, got)
	}
}
