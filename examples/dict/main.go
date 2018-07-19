package main

import (
	"fmt"

	"github.com/go-ego/gse"
)

func main() {
	var seg gse.Segmenter
	seg.LoadDict("zh,../../testdata/test_dict.txt,../../testdata/test_dict1.txt")

	text1 := []byte("深圳地王大厦")
	fmt.Println(seg.String(text1, true))

	segments := seg.Segment(text1)
	// fmt.Println(gse.ToString(segments, false))
	fmt.Println(gse.ToString(segments))
	//"深圳/n 地王大厦/n "

	// 搜索模式主要用于给搜索引擎提供尽可能多的关键字
	segs := seg.ModeSegment(text1, true)
	fmt.Println(gse.ToString(segs, true))
	// "深圳/n 地王大厦/n "
}
