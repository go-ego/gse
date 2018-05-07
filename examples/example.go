package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/go-ego/gse"
)

var (
	text = flag.String("text", "中国互联网历史上最大的一笔并购案", "要分词的文本")
)

func main() {
	flag.Parse()

	var seg gse.Segmenter
	// seg.LoadDict("../data/dict/dictionary.txt")
	// 加载默认词典
	seg.LoadDict()

	segments := seg.Segment([]byte(*text))
	fmt.Println(gse.ToString(segments, true))

	text2 := []byte("深圳地标建筑, 深圳地王大厦")
	segs := seg.Segment(text2)

	log.Println(gse.ToString(segs, false))
	// 深圳/ns 地标/n 建筑/n ,/x  /x 深圳/ns 地王大厦/n

	// segs := seg.ModeSegment(text2, true)
	log.Println("搜索模式: ", gse.ToString(segs, true))
	// 搜索模式: 深圳/ns 地标/n 建筑/n ,/x  /x 深圳/ns 地王/n 大厦/n 地王大厦/n
}
