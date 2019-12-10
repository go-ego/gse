package main

import (
	"fmt"

	"github.com/go-ego/gse"
)

func main() {
	var seg gse.Segmenter
	gse.AlphaNum = true
	seg.LoadDict("zh,../../testdata/test_dict3.txt")

	freq, ok := seg.Find("hello")
	fmt.Println(freq, ok)

	freq, ok = seg.Find("world")
	fmt.Println(freq, ok)

	text := "helloworld! 你好世界"

	tx := seg.Cut(text)
	fmt.Println(tx)

	tx = seg.Cut(text, true)
	fmt.Println(tx)
}
