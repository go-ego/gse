package gse

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/vcaesar/tt"
)

var (
	prodSeg = Segmenter{}
)

func TestGetVer(t *testing.T) {
	fmt.Println("go version: ", runtime.Version())
	ver := GetVersion()

	tt.Expect(t, version, ver)
	expect(t, version, ver)
	tt.Equal(t, version, ver)
}

func TestSplit(t *testing.T) {
	tt.Expect(t, "世/界/有/七/十/亿/人/口/",
		bytesToString(splitTextToWords([]byte(
			"世界有七十亿人口"))))

	tt.Expect(t, "github/ /is/ /a/ /web/-/based/ /hosting/ /service/,/ /for/ /software/ /development/ /projects/./",
		bytesToString(splitTextToWords([]byte(
			"GitHub is a web-based hosting service, for software development projects."))))

	tt.Expect(t, "雅/虎/yahoo/!/ /致/力/于/，/领/先/的/门/户/网/站/。/",
		bytesToString(splitTextToWords([]byte(
			"雅虎Yahoo! 致力于，领先的门户网站。"))))

	tt.Expect(t, "こ/ん/に/ち/は/", bytesToString(splitTextToWords([]byte("こんにちは"))))

	tt.Expect(t, "안/녕/하/세/요/", bytesToString(splitTextToWords([]byte("안녕하세요"))))

	tt.Expect(t, "Я/ /тоже/ /рада/ /Вас/ /видеть/", bytesToString(splitTextToWords([]byte("Я тоже рада Вас видеть"))))

	tt.Expect(t, "¿/cómo/ /van/ /las/ /cosas/", bytesToString(splitTextToWords([]byte("¿Cómo van las cosas"))))

	tt.Expect(t, "wie/ /geht/ /es/ /ihnen/", bytesToString(splitTextToWords([]byte("Wie geht es Ihnen"))))

	tt.Expect(t, "je/ /suis/ /enchanté/ /de/ /cette/ /pièce/",
		bytesToString(splitTextToWords([]byte("Je suis enchanté de cette pièce"))))

	tt.Expect(t, "[[116 111 32 119 111 114 100 115]]", toWords("to words"))
}

func TestSegment(t *testing.T) {
	var seg Segmenter
	seg.LoadDict("testdata/test_dict1.txt,testdata/test_dict2.txt")
	// seg.LoadDict("testdata/test_dict1.txt", "testdata/test_dict2.txt")
	tt.Expect(t, "16", seg.dict.NumTokens())
	// tt.Expect(t, "5", seg.dict.NumTokens())
	segments := seg.Segment([]byte("世界有七十亿人口"))
	tt.Expect(t, "世界/ 有/p3 七十亿/ 人口/p12 ", ToString(segments, false))
	// tt.Expect(t, "世界/ 有/x 七十亿/ 人口/p12 ", ToString(segments, false))

	tt.Expect(t, "4", len(segments))
	tt.Expect(t, "0", segments[0].start)
	tt.Expect(t, "6", segments[0].end)
	tt.Expect(t, "6", segments[1].start)
	tt.Expect(t, "9", segments[1].end)

	tt.Expect(t, "9", segments[2].start)
	tt.Expect(t, "18", segments[2].end)
	tt.Expect(t, "18", segments[3].start)
	tt.Expect(t, "24", segments[3].end)
}

func TestSegmentS(t *testing.T) {
	var seg Segmenter
	seg.LoadDict("zh,testdata/test_dict.txt")
	// seg.LoadDict()

	dict := seg.Dictionary()
	tt.Expect(t, "16", dict.maxTokenLen)
	tt.Expect(t, "53250733", dict.totalFrequency)

	tt.Expect(t, "587880", seg.dict.NumTokens())
	text1 := []byte("纽约帝国大厦, 深圳地王大厦")
	segStr := "纽约/ns 帝国大厦/nr ,/x  /x 深圳/ns 地王大厦/n "

	tt.Expect(t, "纽约/ns 帝国大厦/nr ,/x  /x 深圳/ns 地王大厦/n ", seg.String(text1))
	tt.Expect(t, "[纽约 帝国大厦 ,   深圳 地王大厦]", seg.Slice(text1))

	tt.Expect(t, "纽约/ns 帝国/n 大厦/n 帝国大厦/nr ,/x  /x 深圳/ns 地王/n 大厦/n 地王大厦/n ",
		seg.String(text1, true))
	tt.Expect(t, "[纽约 帝国 大厦 帝国大厦 ,   深圳 地王 大厦 地王大厦]", seg.Slice(text1, true))

	segments := seg.Segment([]byte(text1))
	tt.Expect(t, segStr, ToString(segments))
	tt.Expect(t, segStr, ToString(segments, false))

	segs := seg.ModeSegment([]byte(text1), true)
	tt.Expect(t, segStr, ToString(segs, false))

	tt.Expect(t, "6", len(segments))
	tt.Expect(t, "0", segments[0].start)
	tt.Expect(t, "6", segments[0].end)
	tt.Expect(t, "6", segments[1].start)
	tt.Expect(t, "18", segments[1].end)

	text2 := []byte("留给真爱你的人")
	segments2 := seg.Segment([]byte(text2))
	tt.Expect(t, "留给/v 真爱/nr 你/r 的/uj 人/n ", ToString(segments2, false))

	tt.Expect(t, "5", len(segments2))
	tt.Expect(t, "0", segments2[0].start)
	tt.Expect(t, "6", segments2[0].end)
	tt.Expect(t, "6", segments2[1].start)
	tt.Expect(t, "12", segments2[1].end)
}

func TestSegmentJp(t *testing.T) {
	var seg Segmenter
	seg.LoadDict("data/dict/jp/dict.txt")
	text2 := []byte("こんにちは世界")
	segments := seg.Segment([]byte(text2))

	tt.Expect(t, "こんにちは/感動詞 世界/名詞 ", ToString(segments, false))
	tt.Expect(t, "こん/名詞 こんにちは/感動詞 世界/名詞 ", ToString(segments, true))
	tt.Expect(t, "[こん こんにちは 世界]", ToSlice(segments, true))
	tt.Expect(t, "[こんにちは 世界]", ToSlice(segments, false))
	tt.True(t, IsJp(ToSlice(segments)[0]))

	tt.Expect(t, "2", len(segments))
	tt.Expect(t, "0", segments[0].start)
	tt.Expect(t, "15", segments[0].end)
}

func TestDictPaths(t *testing.T) {
	paths := DictPaths("./dictDir", "zh,jp")
	tt.Expect(t, "2", len(paths))

	if paths[0] != "dictDir/dict/dictionary.txt" {
		t.Errorf("what=\"%s\", got=\"%s\"", "dictDir/dict/dictionary.txt", paths[0])
	}
	if paths[1] != "dictDir/dict/jp/dict.txt" {
		t.Errorf("what=\"%s\", got=\"%s\"", "dictDir/dict/jp/dict.txt", paths[1])
	}
}

func TestSegmentDicts(t *testing.T) {
	var seg Segmenter
	// seg.LoadDict("zh,jp")
	seg.LoadDict("./data/dict/dictionary.txt,./data/dict/jp/dict.txt")

	text1 := []byte("深圳地王大厦")
	segments := seg.Segment([]byte(text1))
	tt.Expect(t, "深圳/ns 地王大厦/n ", ToString(segments, false))

	tt.Expect(t, "2", len(segments))
	tt.Expect(t, "0", segments[0].start)
	tt.Expect(t, "6", segments[0].end)
	tt.Expect(t, "6", segments[1].start)
	tt.Expect(t, "18", segments[1].end)

	text2 := []byte("こんにちは世界")
	segments = seg.Segment([]byte(text2))
	tt.Expect(t, "こんにちは/感動詞 世界/n ", ToString(segments, false))
	tt.Expect(t, "2", len(segments))
	tt.Expect(t, "こん/名詞 こんにちは/感動詞 世界/n ", ToString(segments, true))
	tt.Expect(t, "2", len(segments))
	tt.Expect(t, "0", segments[0].start)
	tt.Expect(t, "15", segments[0].end)

	tt.Expect(t, "0", segments[0].Start())
	tt.Expect(t, "15", segments[0].End())

	token := segments[0].Token()
	tt.Expect(t, "こんにちは", token.Text())
	tt.Expect(t, "5704", token.Frequency())
	tt.Expect(t, "感動詞", token.Pos())

	var tokenArr []*Token
	for i := 0; i < len(segments); i++ {
		tokenArr = append(tokenArr, segments[i].Token())
	}
	tt.Expect(t, "こんにちは 世界 ", printTokens(tokenArr, 2))

	tseg := token.Segments()
	tt.Expect(t, "0", tseg[0].Start())
	tt.Expect(t, "6", tseg[0].End())
}

func TestLargeDictionary(t *testing.T) {
	err := prodSeg.LoadDict("zh,testdata/test_dict2.txt")
	tt.Nil(t, err)

	tt.Expect(t, "世界/n 人口/n ", ToString(prodSeg.Segment(
		[]byte("世界人口")), false))

	tt.Expect(t, "世界/n 人口/n ", ToString(prodSeg.internalSegment(
		[]byte("世界人口"), false), false))

	tt.Expect(t, "世界/n 人口/n ", ToString(prodSeg.internalSegment(
		[]byte("世界人口"), true), false))

	tt.Expect(t, "山达尔星新星联邦共和国/ns 联邦政府/nt ", ToString(prodSeg.internalSegment(
		[]byte("山达尔星新星联邦共和国联邦政府"), true), false))

	tt.Expect(t, "山达尔星新星联邦共和国联邦政府/nt ", ToString(prodSeg.internalSegment(
		[]byte("山达尔星新星联邦共和国联邦政府"), false), false))

	tt.Expect(t, "达尔/nrt 星/n 山达尔星/nz 新星/nz 联邦/n 共和/nz 国/n 共和国/ns 联邦共和国/nt 山达尔星新星联邦共和国/ns 联邦/n 政府/n 联邦政府/nt 山达尔星新星联邦共和国联邦政府/nt ",
		ToString(prodSeg.Segment([]byte("山达尔星新星联邦共和国联邦政府")), true))
}

func TestLoadDictionary(t *testing.T) {
	// 	var seg Segmenter
	// 	seg.LoadDict()
	err := prodSeg.LoadDict()
	tt.Nil(t, err)
	// 	tt.Expect(t, "世界/ns 人口/n ", ToString(prodSeg.Segment(
	// 		[]byte("世界人口")), false))

	// 	tt.Expect(t, "世界/ns 人口/n ", ToString(prodSeg.internalSegment(
	// 		[]byte("世界人口"), false), false))

	// 	tt.Expect(t, "世界/ns 人口/n ", ToString(prodSeg.internalSegment(
	// 		[]byte("世界人口"), true), false))

	// 	tt.Expect(t, "山达尔星新星联邦共和国/ns 联邦政府/nt ", ToString(prodSeg.internalSegment(
	// 		[]byte("山达尔星新星联邦共和国联邦政府"), true), false))

	// 	tt.Expect(t, "山达尔星新星联邦共和国联邦政府/nt ", ToString(prodSeg.internalSegment(
	// 		[]byte("山达尔星新星联邦共和国联邦政府"), false), false))

	// 	tt.Expect(t, "山达尔星/nz 新星/nz 联邦/n 共和/nz 国/n 共和国/ns 联邦共和国/nt 新星联邦共和国/ns 联邦/n 政府/n 联邦政府/nt 联邦联邦政府/nt 新星联邦共和国联邦政府/nt ", ToString(prodSeg.Segment(
	// 		[]byte("山达尔星新星联邦共和国联邦政府")), true))
}

func TestTokenEquals(t *testing.T) {
	token := Token{
		text: []Text{
			[]byte("one"),
			[]byte("two"),
		},
	}

	tt.True(t, token.Equals("onetwo"))
}

func TestTokenNotEquals(t *testing.T) {
	token := Token{
		text: []Text{
			[]byte("one"),
			[]byte("two"),
		},
	}

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
