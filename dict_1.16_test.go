//go:build go1.16
// +build go1.16

package gse

import (
	_ "embed"
	"testing"

	"github.com/go-ego/gse/types"
	"github.com/vcaesar/tt"
)

//go:embed testdata/test_en_dict3.txt
var testDict string

//go:embed testdata/test_en.txt
var testEn string

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

func TestLoadDictZhEmbed(t *testing.T) {
	var seg1 Segmenter
	err := seg1.LoadDictEmbed("zh")
	tt.Nil(t, err)
	tt.Equal(t, 587207, len(seg1.Dict.Tokens))
	// grown once to the total line count of both dicts
	tt.Equal(t, 589035, cap(seg1.Dict.Tokens))
	tt.Equal(t, 5.3226742e+07, seg1.Dict.totalFreq)

	f, pos, ok := seg1.Find("共和国")
	tt.Bool(t, ok)
	tt.Equal(t, "ns", pos)
	tt.Equal(t, 2389, f)
	val, _, err := seg1.Dict.Value([]byte("共和国"))
	tt.Nil(t, err)
	tt.Equal(t, 2, len(seg1.Dict.Tokens[val].segments))
	tt.Equal(t, "[新星 共和 国 共和国]", seg1.CutSearch("新星共和国"))

	var seg2 Segmenter
	seg2.SkipSubSeg = true
	err = seg2.LoadDictEmbed("zh")
	tt.Nil(t, err)
	tt.Equal(t, 587207, len(seg2.Dict.Tokens))
	for i := range seg2.Dict.Tokens {
		tt.Equal(t, 0, len(seg2.Dict.Tokens[i].segments))
	}
	tt.Equal(t, "[新星 共和国]", seg2.CutSearch("新星共和国"))
	tt.Equal(t, seg1.Cut("新星共和国是一个国家"), seg2.Cut("新星共和国是一个国家"))
}

func TestSplitDictLine(t *testing.T) {
	size, text, freq, pos := splitDictLine("word", " ")
	tt.Equal(t, 1, size)
	tt.Equal(t, "word", text)
	tt.Equal(t, "", freq)
	tt.Equal(t, "", pos)

	size, text, freq, pos = splitDictLine("word 20", " ")
	tt.Equal(t, 2, size)
	tt.Equal(t, "20", freq)
	tt.Equal(t, "", pos)

	size, text, freq, pos = splitDictLine("word 20 n \r", " ")
	tt.Equal(t, 3, size)
	tt.Equal(t, "word", text)
	tt.Equal(t, "20", freq)
	tt.Equal(t, "n", pos)

	size, text, freq, pos = splitDictLine("word, 20, n, x", ", ")
	tt.Equal(t, 3, size)
	tt.Equal(t, "word", text)
	tt.Equal(t, "20", freq)
	tt.Equal(t, "n", pos)

	size, text, _, _ = splitDictLine("", " ")
	tt.Equal(t, 1, size)
	tt.Equal(t, "", text)
}

func TestLoadTFIDFDictStrEmbed(t *testing.T) {
	var seg Segmenter
	seg.MinTokenFreq = 0.1
	// a trailing newline and a short line must be skipped, not panic
	dict := "不中 0.38 8.8\n中国 0.5 9.1\nbad\nbad2 3\n"
	err := seg.LoadTFIDFDictStr(&types.LoadDictFile{
		FilePath: dict, FileType: types.LoadDictTypeTFIDF})
	tt.Nil(t, err)
	tt.Equal(t, 2, seg.Dict.NumTokens())

	freq, idf, ok := seg.Dict.FindTFIDF([]byte("中国"))
	tt.Bool(t, ok)
	tt.Equal(t, 0.5, freq)
	tt.Equal(t, 9.1, idf)
}

func TestLoadStopEmbed(t *testing.T) {
	var seg1 Segmenter
	err := seg1.LoadStopEmbed("zh, " + testStop)
	tt.Nil(t, err)
	tt.Bool(t, seg1.IsStop("比如"))
	tt.Bool(t, seg1.IsStop("离开"))
}

func TestDictSep(t *testing.T) {
	var seg1 Segmenter
	seg1.DictSep = ","
	err := seg1.LoadDictEmbed(testEn)
	tt.Nil(t, err)

	f, pos, ok := seg1.Find("to be")
	tt.Bool(t, ok)
	tt.Equal(t, "x", pos)
	tt.Equal(t, 10, f)
}

func TestLoadTFIDFDictStr(t *testing.T) {
	var seg Segmenter
	a := []*types.LoadDictFile{}
	a = append(a, &types.LoadDictFile{
		FilePath: "/workspaces/gse/data/dict/zh/tf_idf.txt",
		FileType: types.LoadDictTypeTFIDF,
	})
	seg.LoadTFIDFDict(a)
}

func TestLoadDictEmbedLangs(t *testing.T) {
	seg := Segmenter{SkipSubSeg: true, NotLoadHMM: true}
	tt.Nil(t, seg.LoadDictEmbed("ja"))
	_, _, ok := seg.Find("こん")
	tt.Bool(t, ok)
	// words before the full-width space line (trimmed to "") must survive
	freq, _, ok := seg.Find("犠牲")
	tt.Bool(t, ok)
	tt.Equal(t, 5622.0, freq)

	empty := Segmenter{SkipSubSeg: true, NotLoadHMM: true}
	tt.Nil(t, empty.LoadDictStr("犠牲 5622 n\n\u3000 1287 x\n運命 5578 n"))
	tt.Equal(t, 2, empty.Dict.NumTokens())
	freq, _, ok = empty.Find("犠牲")
	tt.Bool(t, ok)
	tt.Equal(t, 5622.0, freq)

	tt.Nil(t, seg.LoadDictEmbed("zh_t"))
	tt.Nil(t, seg.LoadDictEmbed("zh_s, word1 20 n"))
	_, pos, ok := seg.Find("word1")
	tt.Bool(t, ok)
	tt.Equal(t, "n", pos)

	seg1 := Segmenter{SkipSubSeg: true, NotLoadHMM: true}
	tt.Nil(t, seg1.LoadDictEmbed())
	tt.Equal(t, 587207, seg1.Dict.NumTokens())
}

func TestLoadStopEmbedDefault(t *testing.T) {
	var seg Segmenter
	tt.Nil(t, seg.LoadStopEmbed())
	tt.Bool(t, seg.IsStop("比如"))

	var seg1 Segmenter
	tt.Nil(t, seg1.LoadStopEmbed("停止\n词语"))
	tt.Bool(t, seg1.IsStop("停止"))
	tt.Bool(t, seg1.IsStop("词语"))
	tt.False(t, seg1.IsStop("比如"))
}
