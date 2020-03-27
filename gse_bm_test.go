package gse

import (
	"testing"

	"github.com/vcaesar/tt"
)

var text = "纽约时代广场, 纽约帝国大厦, 旧金山湾金门大桥"

func init() {
	prodSeg.LoadDict()
}

func BenchmarkCut(t *testing.B) {
	fn := func() {
		prodSeg.Cut(text)
	}

	tt.BM(t, fn)
}

func BenchmarkCutHMM(t *testing.B) {
	fn := func() {
		prodSeg.Cut(text, true)
	}

	tt.BM(t, fn)
}

func BenchmarkCutAll(t *testing.B) {
	fn := func() {
		prodSeg.CutAll(text)
	}

	tt.BM(t, fn)
}

func BenchmarkCutSearch(t *testing.B) {
	fn := func() {
		prodSeg.CutSearch(text)
	}

	tt.BM(t, fn)
}

func BenchmarkCutSearchHMM(t *testing.B) {
	fn := func() {
		prodSeg.CutSearch(text, true)
	}

	tt.BM(t, fn)
}

func BenchmarkSlice(b *testing.B) {
	fn := func() {
		prodSeg.Slice([]byte(text))
	}

	tt.BM(b, fn)
}

func BenchmarkString(b *testing.B) {
	fn := func() {
		prodSeg.String([]byte(text))
	}

	tt.BM(b, fn)
}

func BenchmarkAddToken(b *testing.B) {
	fn := func() {
		prodSeg.AddToken(text, 10)
	}

	tt.BM(b, fn)
}

func BenchmarkRemoveToken(b *testing.B) {
	fn := func() {
		prodSeg.RemoveToken(text)
	}

	tt.BM(b, fn)
}
