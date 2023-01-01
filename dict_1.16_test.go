//go:build go1.16
// +build go1.16

package gse

import (
	_ "embed"
	"testing"

	"github.com/vcaesar/tt"
)

//go:embed testdata/test_en_dict3.txt
var testDict string

//go:embed testdata/zh/test_zh_dict2.txt
var testDict2 string

//go:embed testdata/stop.txt
var testStop string

func TestLoadDictEmbed(t *testing.T) {
	var seg2 Segmenter
	err := seg2.LoadDictEmbed(testDict)
	tt.Nil(t, err)

	seg1, err := NewEmbed("zh, word1 20 n, "+testDict+", "+testDict2, "en")
	tt.Nil(t, err)

	f, pos, ok := seg1.Find("1号店")
	tt.Bool(t, ok)
	tt.Equal(t, "n", pos)
	tt.Equal(t, 3, f)

	f, pos, ok = seg1.Find("hello")
	tt.Bool(t, ok)
	tt.Equal(t, "", pos)
	tt.Equal(t, 20, f)

	f, pos, ok = seg1.Find("world")
	tt.Bool(t, ok)
	tt.Equal(t, "n", pos)
	tt.Equal(t, 20, f)

	f, pos, ok = seg1.Find("word1")
	tt.Bool(t, ok)
	tt.Equal(t, "n", pos)
	tt.Equal(t, 20, f)

	f, pos, ok = seg1.Find("新星共和国")
	tt.Bool(t, ok)
	tt.Equal(t, "ns", pos)
	tt.Equal(t, 32, f)

	f, _, ok = seg1.Find("八千一百三十七万七千二百三十六口")
	tt.Bool(t, ok)
	tt.Equal(t, 2, f)
}

func TestLoadDictSTEmbed(t *testing.T) {
	var seg1 Segmenter
	err := seg1.LoadDictEmbed("zh_s")
	tt.Nil(t, err)
	tt.Equal(t, 352275, len(seg1.Dict.Tokens))
	tt.Equal(t, 3.3335153e+07, seg1.Dict.totalFreq)

	err = seg1.LoadDictEmbed("zh_t, word1 20 n, " + testDict)
	tt.Nil(t, err)
	tt.Equal(t, 587211, len(seg1.Dict.Tokens))
	tt.Equal(t, 5.3226834e+07, seg1.Dict.totalFreq)
}

func TestLoadStopEmbed(t *testing.T) {
	var seg1 Segmenter
	err := seg1.LoadStopEmbed("zh, " + testStop)
	tt.Nil(t, err)
	tt.Bool(t, seg1.IsStop("比如"))
	tt.Bool(t, seg1.IsStop("离开"))
}
