package main

import (
	"fmt"
	"regexp"

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

	reg := regexp.MustCompile(`(\d+年|\d+月|\d+日|[\p{Latin}]+|[\p{Hangul}]+|\d+\.\d+|[a-zA-Z0-9]+)`)
	text1 := `헬로월드 헬로 서울, 2021年09月10日, 3.14`
	hmm = seg.CutDAG(text1, reg)
	fmt.Println("Cut with hmm and regexp: ", hmm, hmm[0], hmm[6])

	//
	// hmm = seg.HMMCutMod(text)
	// fmt.Println("hmm cut: ", hmm)
}
