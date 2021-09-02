package main

import (
	"fmt"

	"github.com/go-ego/gse"
)

var (
	text = "纽约时代广场, 纽约帝国大厦"

	seg gse.Segmenter
)

func main() {
	err := seg.LoadDict()
	fmt.Println("load dictionary error: ", err)
	// seg.LoadModel() // load the hmm model

	hmm := seg.Cut(text, true)
	fmt.Println("hmm cut: ", hmm)

	hmm = seg.CutSearch(text, true)
	fmt.Println("hmm cut: ", hmm)

	hmm = seg.CutAll(text)
	fmt.Println("cut all: ", hmm)

	//
	// hmm = seg.HMMCutMod(text)
	// fmt.Println("hmm cut: ", hmm)
}
