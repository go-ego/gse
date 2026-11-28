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

/*
Package segtool holds the segmentation modes and helpers shared by the
server and benchmark tools.
*/
package segtool

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-ego/gse"
)

// Default is the default mode name
const Default = "segment"

// Func segments the text with a gse mode
type Func func(seg *gse.Segmenter, text string) []gse.SegPos

// Mode is a named gse segmentation mode
type Mode struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
	Func Func   `json:"-"`
}

// Modes lists the supported segmentation modes
var Modes = []Mode{
	{"segment", "shortest path with pos, Pos(text)", func(seg *gse.Segmenter, text string) []gse.SegPos {
		return seg.Pos(text)
	}},
	{"search", "search engine mode with pos, Pos(text, true)", func(seg *gse.Segmenter, text string) []gse.SegPos {
		return seg.Pos(text, true)
	}},
	{"dag", "DAG precise mode, Cut(text, false)", func(seg *gse.Segmenter, text string) []gse.SegPos {
		return toPos(seg.Cut(text, false))
	}},
	{"hmm", "DAG and HMM for the unknown words, Cut(text, true)", func(seg *gse.Segmenter, text string) []gse.SegPos {
		return toPos(seg.Cut(text, true))
	}},
	{"search_hmm", "search engine mode with HMM, CutSearch(text, true)", func(seg *gse.Segmenter, text string) []gse.SegPos {
		return toPos(seg.CutSearch(text, true))
	}},
	{"all", "full mode, all the dictionary words, CutAll(text)", func(seg *gse.Segmenter, text string) []gse.SegPos {
		return toPos(seg.CutAll(text))
	}},
	{"hmm_only", "HMM Viterbi only, HMMCut(text)", func(seg *gse.Segmenter, text string) []gse.SegPos {
		return toPos(seg.HMMCut(text))
	}},
	{"trim", "HMM mode without stop words and punctuation, CutTrim(text, true)", func(seg *gse.Segmenter, text string) []gse.SegPos {
		return toPos(seg.CutTrim(text, true))
	}},
}

func toPos(words []string) []gse.SegPos {
	ps := make([]gse.SegPos, len(words))
	for i, w := range words {
		ps[i].Text = w
	}
	return ps
}

// Get returns the mode by name
func Get(name string) (Mode, error) {
	for _, m := range Modes {
		if m.Name == name {
			return m, nil
		}
	}

	return Mode{}, fmt.Errorf("unknown mode %q, use one of: %s", name, strings.Join(Names(), ", "))
}

// Names returns the mode names
func Names() []string {
	names := make([]string, len(Modes))
	for i, m := range Modes {
		names[i] = m.Name
	}
	return names
}

// LoadDict loads the dict files, or the embedded zh dictionary when empty,
// and the embedded stop words
func LoadDict(seg *gse.Segmenter, dict string) error {
	var err error
	if dict == "" {
		err = seg.LoadDictEmbed()
	} else {
		err = seg.LoadDict(dict)
	}
	if err != nil {
		return err
	}

	return seg.LoadStopEmbed()
}

// ReadLines reads the non-empty lines of the file and their total bytes
func ReadLines(path string) ([]string, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	var (
		lines []string
		size  int
	)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		size += len(line)
		lines = append(lines, line)
	}

	return lines, size, scanner.Err()
}

// Format writes the segments as "text/pos" separated by spaces
func Format(segs []gse.SegPos) string {
	var b strings.Builder
	for i, s := range segs {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s.Text)
		if s.Pos != "" {
			b.WriteByte('/')
			b.WriteString(s.Pos)
		}
	}
	return b.String()
}
