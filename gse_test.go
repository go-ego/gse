package gse

import (
	"testing"

	"github.com/vcaesar/tt"
)

func init() {
	prodSeg.LoadDict()
	prodSeg.LoadStop("zh")
}

func TestHMM(t *testing.T) {
	tt.Equal(t, 587878, len(prodSeg.Dict.Tokens))
	tt.Equal(t, 5.3250719e+07, prodSeg.Dict.totalFrequency)

	hmm := prodSeg.HMMCutMod("纽约时代广场")
	tt.Equal(t, 2, len(hmm))
	tt.Equal(t, "纽约", hmm[0])
	tt.Equal(t, "时代广场", hmm[1])

	// text := "纽约时代广场, 纽约帝国大厦, 旧金山湾金门大桥"
	tx := prodSeg.Cut(text, true)
	tt.Equal(t, 7, len(tx))
	tt.Equal(t, "[纽约时代广场 ,  纽约 帝国大厦 ,  旧金山湾 金门大桥]", tx)

	tx = prodSeg.TrimPunct(tx)
	tt.Equal(t, 5, len(tx))
	tt.Equal(t, "[纽约时代广场 纽约 帝国大厦 旧金山湾 金门大桥]", tx)

	tx = prodSeg.cutDAGNoHMM(text)
	tt.Equal(t, 9, len(tx))
	tt.Equal(t, "[纽约时代广场 ,   纽约 帝国大厦 ,   旧金山湾 金门大桥]", tx)

	tx = append(tx, " 广场")
	tx = append(tx, "ok👌", "的")
	tt.Bool(t, prodSeg.IsStop("的"))

	tx = prodSeg.Trim(tx)
	tt.Equal(t, 7, len(tx))
	tt.Equal(t, "[纽约时代广场 纽约 帝国大厦 旧金山湾 金门大桥 广场 ok]", tx)

	tx1 := prodSeg.CutTrim(text, true)
	tt.Equal(t, 5, len(tx1))
	tt.Equal(t, "[纽约时代广场 纽约 帝国大厦 旧金山湾 金门大桥]", tx1)

	s := prodSeg.CutStr(tx, ", ")
	tt.Equal(t, 80, len(s))
	tt.Equal(t, "纽约时代广场, 纽约, 帝国大厦, 旧金山湾, 金门大桥, 广场, ok", s)

	tx = prodSeg.CutAll(text)
	tt.Equal(t, 21, len(tx))
	tt.Equal(t,
		"[纽约 纽约时代广场 时代 时代广场 广场 ,   纽约 帝国 帝国大厦 国大 大厦 ,   旧金山 旧金山湾 金山 山湾 金门 金门大桥 大桥]",
		tx)

	tx = prodSeg.CutSearch(text, false)
	tt.Equal(t, 20, len(tx))
	tt.Equal(t,
		"[纽约 时代 广场 纽约时代广场 ,   纽约 帝国 国大 大厦 帝国大厦 ,   金山 山湾 旧金山 旧金山湾 金门 大桥 金门大桥]",
		tx)

	tx = prodSeg.CutSearch(text, true)
	tt.Equal(t, 18, len(tx))
	tt.Equal(t,
		"[纽约 时代 广场 纽约时代广场 ,  纽约 帝国 国大 大厦 帝国大厦 ,  金山 山湾 旧金山 旧金山湾 金门 大桥 金门大桥]",
		tx)

	f1 := prodSeg.SuggestFreq("西雅图")
	tt.Equal(t, 79, f1)

	f1 = prodSeg.SuggestFreq("西雅图", "西雅图都会区", "旧金山湾")
	tt.Equal(t, 0, f1)
}

func TestPos(t *testing.T) {
	s := prodSeg.String(text, true)
	tt.Equal(t, 206, len(s))
	tt.Equal(t,
		"纽约/ns 时代/n 广场/n 时代广场/n 纽约时代广场/nt ,/x  /x 纽约/ns 帝国/n 大厦/n 帝国大厦/nr ,/x  /x 金山/nr 旧金山/ns 湾/zg 旧金山湾/ns 金门/n 大桥/ns 金门大桥/nz ", s)

	c := prodSeg.Slice(text, true)
	tt.Equal(t, 20, len(c))

	pos := prodSeg.Pos(text, false)
	tt.Equal(t, 9, len(pos))
	tt.Equal(t,
		"[{纽约时代广场 nt} {, x} {  x} {纽约 ns} {帝国大厦 nr} {, x} {  x} {旧金山湾 ns} {金门大桥 nz}]", pos)

	pos = prodSeg.TrimPosPunct(pos)
	tt.Equal(t, 5, len(pos))
	tt.Equal(t, "[{纽约时代广场 nt} {纽约 ns} {帝国大厦 nr} {旧金山湾 ns} {金门大桥 nz}]", pos)

	pos = prodSeg.Pos(text, true)
	tt.Equal(t, 20, len(pos))
	tt.Equal(t,
		"[{纽约 ns} {时代 n} {广场 n} {时代广场 n} {纽约时代广场 nt} {, x} {  x} {纽约 ns} {帝国 n} {大厦 n} {帝国大厦 nr} {, x} {  x} {金山 nr} {旧金山 ns} {湾 zg} {旧金山湾 ns} {金门 n} {大桥 ns} {金门大桥 nz}]", pos)

	pos1 := prodSeg.PosTrim(text, true, "zg")
	tt.Equal(t, 15, len(pos1))
	tt.Equal(t,
		"[{纽约 ns} {时代 n} {广场 n} {时代广场 n} {纽约时代广场 nt} {纽约 ns} {帝国 n} {大厦 n} {帝国大厦 nr} {金山 nr} {旧金山 ns} {旧金山湾 ns} {金门 n} {大桥 ns} {金门大桥 nz}]", pos1)

	pos = append(pos, SegPos{Text: "👌", Pos: "x"})
	pos = prodSeg.TrimPos(pos)
	tt.Equal(t, 16, len(pos))
	tt.Equal(t,
		"[{纽约 ns} {时代 n} {广场 n} {时代广场 n} {纽约时代广场 nt} {纽约 ns} {帝国 n} {大厦 n} {帝国大厦 nr} {金山 nr} {旧金山 ns} {湾 zg} {旧金山湾 ns} {金门 n} {大桥 ns} {金门大桥 nz}]", pos)

	s = prodSeg.PosStr(pos, ", ")
	tt.Equal(t, 204, len(s))
	tt.Equal(t,
		"纽约/ns, 时代/n, 广场/n, 时代广场/n, 纽约时代广场/nt, 纽约/ns, 帝国/n, 大厦/n, 帝国大厦/nr, 金山/nr, 旧金山/ns, 湾/zg, 旧金山湾/ns, 金门/n, 大桥/ns, 金门大桥/nz", s)

	prodSeg.SkipPos = true
	s = prodSeg.PosStr(pos, ", ")
	tt.Equal(t, 162, len(s))
	tt.Equal(t,
		"纽约, 时代, 广场, 时代广场, 纽约时代广场, 纽约, 帝国, 大厦, 帝国大厦, 金山, 旧金山, 湾, 旧金山湾, 金门, 大桥, 金门大桥", s)

	pos = prodSeg.TrimWithPos(pos, "n", "zg")
	tt.Equal(t, 9, len(pos))
	tt.Equal(t,
		"[{纽约 ns} {纽约时代广场 nt} {纽约 ns} {帝国大厦 nr} {金山 nr} {旧金山 ns} {旧金山湾 ns} {大桥 ns} {金门大桥 nz}]", pos)

	pos2 := prodSeg.PosTrimArr(text, false, "n", "zg")
	tt.Equal(t, 5, len(pos2))
	tt.Equal(t,
		"[纽约时代广场 纽约 帝国大厦 旧金山湾 金门大桥]", pos2)

	pos3 := prodSeg.PosTrimStr(text, false, "n", "zg")
	tt.Equal(t, 64, len(pos3))
	tt.Equal(t,
		"纽约时代广场 纽约 帝国大厦 旧金山湾 金门大桥", pos3)
}

func TestStop(t *testing.T) {
	var seg Segmenter
	err := seg.LoadStop()
	tt.Nil(t, err)
	tt.Equal(t, 88, len(seg.StopWordMap))

	err = seg.LoadStop("testdata/stop.txt")
	tt.Nil(t, err)
	tt.Equal(t, 89, len(seg.StopWordMap))
	tt.Bool(t, seg.IsStop("离开"))

	// err := prodSeg.LoadStop("zh")
	// tt.Nil(t, err)
	tt.Equal(t, 1161, len(prodSeg.StopWordMap))

	b := prodSeg.IsStop("阿")
	tt.True(t, b)

	tt.True(t, prodSeg.IsStop("哎"))
	b = prodSeg.IsStop("的")
	tt.True(t, b)

	prodSeg.AddStop("lol")
	b = prodSeg.IsStop("lol")
	tt.True(t, b)

	prodSeg.RemoveStop("lol")
	b = prodSeg.IsStop("lol")
	tt.False(t, b)

	t1 := `hi, bot, 123; 🤖, 机器人; 👌^_^😆`
	s := FilterEmoji(t1)
	tt.Equal(t, "hi, bot, 123; , 机器人; ^_^", s)

	s = FilterSymbol(t1)
	tt.Equal(t, "hibot123机器人", s)

	s = FilterLang(t1, "Han")
	tt.Equal(t, "hibot机器人", s)

	t2 := `<p>test: </p> <div class="bot"> bot 机器人 <<银河系漫游指南>> </div>`
	s = FilterHtml(t2)
	tt.Equal(t, "test:   bot 机器人 <<银河系漫游指南>> ", s)

	s = prodSeg.CutTrimHtmls(t2, true)
	tt.Equal(t, "test bot 机器人 银河系 漫游 指南", s)
}
