// +build go1.16

package main

import (
	_ "embed"
	"fmt"

	"github.com/go-ego/gse"
)

//go:embed test_dict3.txt
var testDict string

var (
	text = "沙漠的那边还是沙漠, hello world"
	seg  gse.Segmenter
)

func main() {
	var err error
	seg, err = gse.NewEmbed("zh, "+testDict, "en")
	fmt.Println("gse NewEmbed error: ", err)

	s := seg.Cut(text, true)
	fmt.Println(s)
}

func load1() {
	err := seg.LoadDictEmbed()
	fmt.Println(err)
	err = seg.LoadDictStr(testDict)
	fmt.Println(err)

	err = seg.LoadStopEmbed()
	fmt.Println(err)
}
