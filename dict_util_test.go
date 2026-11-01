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
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-ego/gse/types"
	"github.com/vcaesar/tt"
)

// writeFile writes the content to dir/name, creating the parent dirs
func writeFile(t *testing.T, dir, name, content string) string {
	file := filepath.Join(dir, name)
	tt.Nil(t, os.MkdirAll(filepath.Dir(file), 0o755))
	tt.Nil(t, os.WriteFile(file, []byte(content), 0o644))
	return file
}

// newDictPathSeg returns a segmenter with the data dir in dir
func newDictPathSeg(dir string) *Segmenter {
	return &Segmenter{
		SkipLog:    true,
		NotLoadHMM: true,
		DictPath:   filepath.Join(dir, "gse.go"),
	}
}

func TestReAddTokenErr(t *testing.T) {
	seg := newDagSeg(t)
	tt.NotNil(t, seg.ReAddToken("不存在", 10))
	tt.Nil(t, seg.ReAddToken("丁戊", 10))
}

func TestLoadDictMapNil(t *testing.T) {
	seg := &Segmenter{SkipLog: true, NotLoadHMM: true}
	err := seg.LoadDictMap([]map[string]string{
		{"text": "to be", "freq": "10"},
		{"text": "no freq"},
	})
	tt.Nil(t, err)

	_, _, ok := seg.Find("to be")
	tt.Bool(t, ok)
	_, _, ok = seg.Find("no freq")
	tt.False(t, ok)
}

func TestLoadDictErr(t *testing.T) {
	seg := &Segmenter{SkipLog: true, NotLoadHMM: true}
	tt.NotNil(t, seg.LoadDict("testdata/not_exist.txt"))

	dir := t.TempDir()
	seg = newDictPathSeg(dir)
	tt.NotNil(t, seg.LoadDict())

	// s_1 exists but t_1 is missing
	writeFile(t, dir, "data/"+zhS1, "测试 10 n\n")
	seg = newDictPathSeg(dir)
	tt.NotNil(t, seg.LoadDict())
	_, _, ok := seg.Find("测试")
	tt.Bool(t, ok)
}

func TestLoadTFIDFDict(t *testing.T) {
	dir := t.TempDir()
	file := writeFile(t, dir, "tfidf.txt", "测试 10 3.5\n分词 20 4\n")

	seg := &Segmenter{NotLoadHMM: true}
	err := seg.LoadTFIDFDict([]*types.LoadDictFile{
		{FilePath: "en"},
		{FilePath: file},
	})
	tt.Nil(t, err)

	freq, idf, ok := seg.FindTFIDF("测试")
	tt.Bool(t, ok)
	tt.Equal(t, 10, freq)
	tt.Equal(t, 3.5, idf)

	err = seg.LoadTFIDFDict([]*types.LoadDictFile{{FilePath: "not_exist.txt"}})
	tt.NotNil(t, err)
}

func TestLoadCorpusAverLen(t *testing.T) {
	dir := t.TempDir()
	seg := newDictPathSeg(dir)

	tt.Equal(t, filepath.Join(dir, "data/dict/zh/tf_idf.txt"), seg.GetTfIdfPath()[0])
	tt.Equal(t, []string{"a.txt", filepath.Join(dir, "data/dict/zh/idf.txt")},
		seg.GetIdfPath("a.txt"))

	// the default corpus file is missing
	_, err := seg.LoadCorpusAverLen()
	tt.NotNil(t, err)

	writeFile(t, dir, "data/dict/zh/tf_idf_origin.txt", "ab\ncdef\n")
	file := writeFile(t, dir, "corpus.txt", "一二三四五六\n")
	seg.SkipLog = false

	avg, err := seg.LoadCorpusAverLen(file)
	tt.Nil(t, err)
	// (6 + (2+4)/2) / 2
	tt.Equal(t, 4.5, avg)
}

func TestSizeNoFreq(t *testing.T) {
	seg := &Segmenter{LoadNoFreq: true, TextFreq: "5", MinTokenFreq: 2}
	tt.Equal(t, 5, seg.Size(1, "测试", ""))
	tt.Equal(t, 0, seg.Size(0, "测试", "10"))
	tt.Equal(t, 0, seg.Size(2, "测试", "abc"))
	tt.Equal(t, 0, seg.Size(2, "测试", "1"))
	tt.Equal(t, 2, seg.Size(2, "测", "10"))
}

func TestReaderMoreLog(t *testing.T) {
	seg := &Segmenter{SkipLog: true, NotLoadHMM: true, MoreLog: true}
	seg.Dict = NewDict()
	seg.Init()

	r := bufio.NewReader(strings.NewReader("测试 10 n extra\n分词 20\n"))
	tt.Nil(t, seg.Reader(r))

	_, pos, ok := seg.Find("测试")
	tt.Bool(t, ok)
	tt.Equal(t, "n", pos)
	_, _, ok = seg.Find("分词")
	tt.Bool(t, ok)
}

func TestReaderTFIDF(t *testing.T) {
	seg := &Segmenter{SkipLog: true, NotLoadHMM: true, MoreLog: true}
	seg.Dict = NewDict()
	seg.Init()

	text := "测试 10 3\n低频 1 3\n\n额外 10 3 x\n"
	tt.Nil(t, seg.ReaderTFIDF(bufio.NewReader(strings.NewReader(text)), "a.txt"))

	freq, idf, ok := seg.FindTFIDF("测试")
	tt.Bool(t, ok)
	tt.Equal(t, 10, freq)
	tt.Equal(t, 3, idf)
	_, _, ok = seg.FindTFIDF("低频")
	tt.False(t, ok)
	_, _, ok = seg.FindTFIDF("额外")
	tt.Bool(t, ok)

	seg.DictSep = ","
	seg.MoreLog = false
	text = "分词, 20, 4\n缺少, 5\n结尾, 30, 6"
	tt.Nil(t, seg.ReaderTFIDF(bufio.NewReader(strings.NewReader(text))))

	freq, idf, ok = seg.FindTFIDF("分词")
	tt.Bool(t, ok)
	tt.Equal(t, 20, freq)
	tt.Equal(t, 4, idf)
	_, _, ok = seg.FindTFIDF("缺少")
	tt.False(t, ok)
	freq, idf, ok = seg.FindTFIDF("结尾")
	tt.Bool(t, ok)
	tt.Equal(t, 30, freq)
	tt.Equal(t, 6, idf)

	seg.MoreLog = true
	tt.Nil(t, seg.ReaderTFIDF(bufio.NewReader(strings.NewReader("末尾, 30, 6"))))
	_, _, ok = seg.FindTFIDF("末尾")
	tt.Bool(t, ok)
}
