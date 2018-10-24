package hmm

import (
	"math"
	"testing"

	"github.com/vcaesar/tt"
)

var testText = "纽约时代广场"

func init() {
	loadDefEmit()
}

func TestViterbi(t *testing.T) {
	states := []byte{'B', 'M', 'E', 'S'}
	prob, path := Viterbi([]rune(testText), states)
	if math.Abs(prob+43.124804533979976) > 1e-10 {
		t.Fatal(prob)
	}

	for index, state := range []byte{'B', 'E', 'B', 'M', 'M', 'E'} {
		tt.Equal(t, state, path[index])
	}
}

func TestCutHan(t *testing.T) {
	result := internalCut(testText)
	tt.Equal(t, 2, len(result))

	tt.Equal(t, "纽约", result[0])
	tt.Equal(t, "时代广场", result[1])
}

func TestCut(t *testing.T) {
	result := Cut(testText)
	tt.Equal(t, 2, len(result))

	result2 := Cut("New York City.")
	tt.Equal(t, 6, len(result2))
}

func Benchmark_Hmm(b *testing.B) {
	fn := func() {
		states := []byte{'B', 'M', 'E', 'S'}
		Viterbi([]rune(testText), states)
	}

	tt.BM(b, fn)
}

func Benchmark_Hmm_Cut(b *testing.B) {
	fn := func() {
		Cut(testText)
	}

	tt.BM(b, fn)
}
