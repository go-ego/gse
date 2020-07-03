package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/idf"
	"github.com/go-ego/gse/hmm/pos"
)

var (
	seg    gse.Segmenter
	posSeg pos.Segmenter

	text  = "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄制作的的科幻片."
	text1 = flag.String("text", text, "要分词的文本")

	text2 = "西雅图地标建筑, Seattle Space Needle, 西雅图太空针. Sky tree."
)

func main() {
	flag.Parse()

	// 加载默认词典
	seg.LoadDict()
	// seg.LoadDict("../data/dict/dictionary.txt")
	//
	// 使用自定义字典
	// seg.LoadDict("zh,../../testdata/test_dict.txt,../../testdata/test_dict1.txt")

	addToken()

	cut()
	//
	cutPos()
	segCut()

	extAndRank(seg)
}

func addToken() {
	seg.AddToken("《复仇者联盟3：无限战争》", 100, "n")
	seg.AddToken("西雅图中心", 100)
	seg.AddToken("西雅图太空针", 100, "n")
	seg.AddToken("Space Needle", 100, "n")
	// seg.AddTokenForce("上海东方明珠广播电视塔", 100, "n")
	//
	seg.AddToken("太空针", 100)
	freq, ok := seg.Find("太空针")
	fmt.Println("seg.Find: ", freq, ok)

	// seg.CalcToken()
	seg.RemoveToken("太空针")
}

// 使用 DAG 或 HMM 模式分词
func cut() {
	// use DAG and HMM
	hmm := seg.Cut(text, true)
	fmt.Println("cut use hmm: ", hmm)
	//
	cut := seg.Cut(text)
	fmt.Println("cut: ", cut)

	hmm = seg.CutSearch(text, true)
	fmt.Println("cut search use hmm: ", hmm)
	//
	cut = seg.CutSearch(text)
	fmt.Println("cut search: ", cut)

	cut = seg.CutAll(text)
	fmt.Println("cut all: ", cut)

	posAndTrim(cut)
}

func posAndTrim(cut []string) {
	cut = seg.Trim(cut)
	fmt.Println("cut all: ", cut)

	posSeg.WithGse(seg)
	po := posSeg.Cut(text, true)
	fmt.Println("pos: ", po)

	po = posSeg.TrimPos(po)
	fmt.Println("trim pos: ", po)
}

func cutPos() {
	fmt.Println(seg.String(text2, true))
	fmt.Println(seg.Slice(text2, true))

	po := seg.Pos(text2, true)
	fmt.Println("pos: ", po)
	po = seg.TrimPos(po)
	fmt.Println("trim pos: ", po)
}

// 使用最短路径和动态规划分词
func segCut() {
	segments := seg.Segment([]byte(*text1))
	fmt.Println(gse.ToString(segments, true))

	segs := seg.Segment([]byte(text2))
	// log.Println(gse.ToString(segs, false))
	log.Println(gse.ToString(segs))
	// 上海/ns 地标/n 建筑/n ,/x  /x 上海/ns 东方明珠/nr 电视塔/n

	// 搜索模式主要用于给搜索引擎提供尽可能多的关键字
	// segs := seg.ModeSegment(text2, true)
	log.Println("搜索模式: ", gse.ToString(segs, true))
	log.Println("to slice", gse.ToSlice(segs, true))
}

func extAndRank(segs gse.Segmenter) {
	var te idf.TagExtracter
	te.WithGse(segs)
	err := te.LoadIdf()
	fmt.Println(err)

	segments := te.ExtractTags(text, 5)
	fmt.Println("segments: ", len(segments), segments)

	var tr idf.TextRanker
	tr.WithGse(segs)

	results := tr.TextRank(text, 5)
	fmt.Println("results: ", results)
}
