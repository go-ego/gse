package gse

import (
	"testing"

	"github.com/vcaesar/tt"
)

func TestHMM(t *testing.T) {
	prodSeg.LoadDict()

	hmm := prodSeg.HMMCutMod("纽约时代广场")
	tt.Equal(t, 2, len(hmm))
	tt.Equal(t, "纽约", hmm[0])
	tt.Equal(t, "时代广场", hmm[1])

	// text := "纽约时代广场, 纽约帝国大厦, 旧金山湾金门大桥"
	tx := prodSeg.Cut(text, true)
	tt.Equal(t, 7, len(tx))
	tt.Equal(t, "[纽约时代广场 ,  纽约 帝国大厦 ,  旧金山湾 金门大桥]", tx)

	tx = prodSeg.cutDAGNoHMM(text)
	tt.Equal(t, 9, len(tx))
	tt.Equal(t, "[纽约时代广场 ,   纽约 帝国大厦 ,   旧金山湾 金门大桥]", tx)

	tx = prodSeg.Trim(tx)
	tt.Equal(t, 5, len(tx))
	tt.Equal(t, "[纽约时代广场 纽约 帝国大厦 旧金山湾 金门大桥]", tx)

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
}

func TestPos(t *testing.T) {
	pos := prodSeg.Pos(text, false)
	tt.Equal(t, 7, len(pos))
	tt.Equal(t,
		"[{纽约时代广场 nt} {, x} {纽约 ns} {帝国大厦 nr} {, x} {旧金山湾 ns} {金门大桥 nz}]", pos)

	pos = prodSeg.Pos(text, true)
	tt.Equal(t, 18, len(pos))
	tt.Equal(t,
		"[{纽约 ns} {时代 n} {广场 n} {时代广场 n} {纽约时代广场 nt} {, x} {纽约 ns} {帝国 n} {大厦 n} {帝国大厦 nr} {, x} {金山 nr} {旧金山 ns} {湾 zg} {旧金山湾 ns} {金门 n} {大桥 ns} {金门大桥 nz}]", pos)

	pos = prodSeg.TrimPos(pos)
	tt.Equal(t, 16, len(pos))
	tt.Equal(t,
		"[{纽约 ns} {时代 n} {广场 n} {时代广场 n} {纽约时代广场 nt} {纽约 ns} {帝国 n} {大厦 n} {帝国大厦 nr} {金山 nr} {旧金山 ns} {湾 zg} {旧金山湾 ns} {金门 n} {大桥 ns} {金门大桥 nz}]", pos)
}
