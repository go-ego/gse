package gse

import (
	"fmt"
	"testing"

	"github.com/vcaesar/tt"
)

var token = Token{
	text: []Text{
		[]byte("one"),
		[]byte("two"),
	},
}

func TestTokenEquals(t *testing.T) {
	tt.True(t, token.Equals("onetwo"))
}

func TestTokenNotEquals(t *testing.T) {
	tt.False(t, token.Equals("one-two"))
}

var strs = []Text{
	Text("one"),
	Text("two"),
	Text("three"),
	Text("four"),
}

func TextSliceToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		textSliceToString(strs)
	}
}

func TextToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		textToString(strs)
	}
}

func TestBenchmark(t *testing.T) {
	fmt.Println("textToString: ")
	fmt.Println(testing.Benchmark(TextToString))

	fmt.Println("textSliceToString: ")
	fmt.Println(testing.Benchmark(TextSliceToString))
}

func BenchmarkEquals(t *testing.B) {
	fn := func() {
		token.Equals("onetwo")
	}

	tt.BM(t, fn)
}
