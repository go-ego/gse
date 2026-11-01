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

package stopwords

import (
	"testing"

	"github.com/vcaesar/tt"
)

func TestStopWord(t *testing.T) {
	s := NewStopWord()
	tt.Bool(t, s.IsStopWord("the"))
	tt.Bool(t, !s.IsStopWord("gse-word"))

	s.AddStop("gse-word")
	tt.Bool(t, s.IsStopWord("gse-word"))
	tt.Bool(t, StopWordMap["gse-word"])

	s.RemoveStop("gse-word")
	tt.Bool(t, !s.IsStopWord("gse-word"))
}

func TestLoadDict(t *testing.T) {
	s := NewStopWord()
	s.seg.SkipLog = true

	err := s.LoadDict("../../testdata/stop.txt")
	tt.Nil(t, err)
	tt.Bool(t, s.IsStopWord("leave"))
	tt.Bool(t, s.IsStopWord("离开"))

	err = s.LoadDict()
	tt.Nil(t, err)
	tt.Bool(t, s.IsStopWord("?"))
}

func TestLoadDictErr(t *testing.T) {
	s := NewStopWord()
	s.seg.SkipLog = true

	err := s.LoadDict("../../testdata/not_exist_stop.txt")
	tt.NotNil(t, err)
}
