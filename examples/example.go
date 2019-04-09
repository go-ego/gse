package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/go-ego/gse"
)

var (
	seg gse.Segmenter

	text  = "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄"
	text1 = flag.String("text", text, "要分词的文本")

	text2 = []byte("上海地标建筑, 上海东方明珠广播电视塔")
)

// 使用 DAG 或 HMM 模式分词
func cut() {
	// use DAG and HMM
	hmm := seg.Cut(text, true)
	fmt.Println("hmm cut: ", hmm)

	hmm = seg.Cut(text)
	fmt.Println("hmm cut: ", hmm)

	hmm = seg.CutSearch(text, true)
	fmt.Println("hmm cut: ", hmm)

	hmm = seg.CutSearch(text)
	fmt.Println("hmm cut: ", hmm)

	hmm = seg.CutAll(text)
	fmt.Println("cut all: ", hmm)
}

func main() {
	flag.Parse()

	// 加载默认词典
	seg.LoadDict()
	// seg.LoadDict("../data/dict/dictionary.txt")

	// 使用自定义字典
	// seg.LoadDict("zh,../../testdata/test_dict.txt,../../testdata/test_dict1.txt")

	seg.AddToken("《复仇者联盟3：无限战争》", 100, "n")
	seg.AddToken("上海中心大厦", 100)
	seg.AddTokenForce("上海东方明珠广播电视塔", 100, "n")
	//
	seg.AddToken("东方明珠广播电视塔", 100)
	seg.CalcToken()

	cut()

	segCut()
}

// 使用最短路径和动态规划分词
func segCut() {
	segments := seg.Segment([]byte(*text1))
	fmt.Println(gse.ToString(segments, true))

	segs := seg.Segment(text2)
	// log.Println(gse.ToString(segs, false))
	log.Println(gse.ToString(segs))
	// 上海/ns 地标/n 建筑/n ,/x  /x 上海/ns 东方明珠/nr 电视塔/n

	// 搜索模式主要用于给搜索引擎提供尽可能多的关键字
	// segs := seg.ModeSegment(text2, true)
	log.Println("搜索模式: ", gse.ToString(segs, true))
	// 搜索模式: 上海/ns 地标/n 建筑/n ,/x  /x 上海/ns 东方/s 明珠/nr 东方明珠/nr 电视/n 塔/j 电视塔/n

	log.Println("to slice", gse.ToSlice(segs, true))

	fmt.Println(seg.String(text2, true))
	fmt.Println(seg.Slice(text2, true))
}
