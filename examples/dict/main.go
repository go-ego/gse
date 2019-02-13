package main

import (
	"fmt"

	"github.com/go-ego/gse"
)

func main() {
	var seg gse.Segmenter
	seg.LoadDict("zh,../../testdata/test_dict.txt,../../testdata/test_dict1.txt")

	text1 := []byte("旧金山湾金门大桥")
	fmt.Println(seg.String(text1, true))
	// 金山/nr 旧金山/ns 湾/zg 旧金山湾/ns 金门/n 大桥/ns 金门大桥/nz

	segments := seg.Segment(text1)
	// fmt.Println(gse.ToString(segments, false))
	fmt.Println(gse.ToString(segments))
	//"旧金山湾/n 金门大桥/nz "

	// 搜索模式主要用于给搜索引擎提供尽可能多的关键字
	segs := seg.ModeSegment(text1, true)
	fmt.Println(gse.ToString(segs, true))
	// "金山/nr 旧金山/ns 湾/zg 旧金山湾/ns 金门/n 大桥/ns 金门大桥/nz "
}
