package main

import (
	"fmt"

	"github.com/go-ego/gse"
)

func main() {
	seg := gse.New("zh,../../testdata/test_dict3.txt", "alpha")
	// var seg gse.Segmenter
	// seg.AlphaNum = true
	// seg.LoadDict("zh, ../../testdata/test_dict3.txt")
	seg.AddToken("winter is coming!", 100, "n")

	freq, ok := seg.Find("hello")
	fmt.Println(freq, ok)

	freq, ok = seg.Find("world")
	fmt.Println(freq, ok)

	text := "Helloworld, winter is coming! 你好世界."

	tx := seg.Cut(text)
	fmt.Println(tx)

	tx = seg.Cut(text, true)
	fmt.Println(tx)

	tx = seg.Trim(tx)
	fmt.Println(tx)

	a := seg.Analyze(tx)
	fmt.Println(a)
}
