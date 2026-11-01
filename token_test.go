package gse

import (
	"fmt"
	"strings"
	"testing"

	"github.com/vcaesar/tt"
)

var token = Token{text: Text("onetwo"), size: 2}

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

func TestToTokenText(t *testing.T) {
	var seg Segmenter
	token := seg.ToToken("Hello世界", 10, "n")
	tt.Equal(t, "hello世界", token.Text())
	tt.Equal(t, 3, int(token.size))
	tt.Equal(t, 3, len(seg.SplitTextToWords(token.text)))

	seg.AlphaNum = true
	token = seg.ToToken("ab世", 10)
	tt.Equal(t, "ab世", token.Text())
	tt.Equal(t, 3, int(token.size))

	ToLower = false
	defer func() { ToLower = true }()
	token = seg.ToToken("Ab世", 10)
	tt.Equal(t, "Ab世", token.Text())
	tt.Equal(t, 3, int(token.size))
}

func TestFirstIsJp(t *testing.T) {
	tt.True(t, firstIsJp(Text("こんにちは")))
	tt.True(t, firstIsJp(Text("カタカナ")))
	tt.False(t, firstIsJp(Text("世界")))
	tt.False(t, firstIsJp(Text("abc")))
	tt.False(t, firstIsJp(nil))
}

func TestDictTokenText(t *testing.T) {
	var seg Segmenter
	seg.SkipLog = true
	seg.Dict = NewDict()
	tt.Nil(t, seg.AddToken("西雅图", 100, "ns"))
	tt.Nil(t, seg.AddToken("西雅", 10))
	seg.CalcToken()

	freq, pos, ok := seg.Find("西雅图")
	tt.True(t, ok)
	tt.Equal(t, 100, freq)
	tt.Equal(t, "ns", pos)
	tt.Equal(t, 3, seg.Dict.MaxTokenLen())
	tt.Equal(t, "[西雅 图]", seg.Slice("西雅图", true))
	tt.Equal(t, "西雅/ 图/x ", seg.String("西雅图", true))
	tt.Equal(t, 2, len(seg.Dict.Tokens[0].Segments()))

	tt.Nil(t, seg.RemoveToken("西雅图"))
	_, _, err := seg.Dict.Value([]byte("西雅图"))
	tt.NotNil(t, err)
}

func TestSplitOptionChanged(t *testing.T) {
	var seg Segmenter
	seg.SkipLog = true
	seg.Dict = NewDict()
	seg.AlphaNum = true
	tt.Nil(t, seg.AddToken("ab世", 10))
	tt.Nil(t, seg.AddToken("ab", 10))

	// the tokens are stored with 3 and 2 words, now split to 2 and 1
	seg.AlphaNum = false
	seg.CalcToken()
	tt.Equal(t, "[ab世]", seg.Slice("ab世"))
	tt.Equal(t, "[ab 世]", seg.Slice("ab世", true))
	tt.Equal(t, 2, len(seg.Dict.Tokens[0].Segments()))

	seg.AlphaNum = true
	seg.CalcToken()
	tt.Equal(t, "[ab世]", seg.Slice("ab世"))
	tt.Equal(t, "[ab 世]", seg.Slice("ab世", true))
}

func TestLookupTokensSize(t *testing.T) {
	var seg Segmenter
	seg.Dict = NewDict()
	tt.Nil(t, seg.AddToken("西雅", 10))
	tt.Nil(t, seg.AddToken("西雅图", 10))

	words := seg.SplitTextToWords([]byte("西雅图人"))
	tokens := make([]*Token, 4)
	sizes := make([]int, 4)
	tt.Equal(t, 2, seg.Dict.lookupTokens(words, tokens, sizes))
	tt.Equal(t, "[2 3]", sizes[:2])
	tt.Equal(t, "西雅图", tokens[1].Text())
	tt.Equal(t, 2, seg.Dict.LookupTokens(words, tokens))
}

func TestSegmentLongToken(t *testing.T) {
	var seg Segmenter
	seg.Dict = NewDict()
	long := strings.Repeat("长", 40) // above the 32 stack buffer
	tt.Nil(t, seg.AddToken(long, 100))
	seg.CalcToken()

	tt.Equal(t, 40, seg.Dict.MaxTokenLen())
	tt.Equal(t, "[短 "+long+" 短]", seg.Slice("短"+long+"短"))
	tt.Equal(t, 40, len(seg.Dict.Tokens[0].Segments()))
}
