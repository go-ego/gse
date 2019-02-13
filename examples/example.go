package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/go-ego/gse"
)

var (
	text = flag.String("text",
		"《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄", "要分词的文本")

	text2 = []byte("上海地标建筑, 上海东方明珠电视塔")
)

func main() {
	flag.Parse()

	var seg gse.Segmenter
	// seg.LoadDict("../data/dict/dictionary.txt")
	// 加载默认词典
	seg.LoadDict()

	segments := seg.Segment([]byte(*text))
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
