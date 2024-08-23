package gse_test

import (
	"fmt"
	"testing"
	"unicode/utf8"
)

func Test_ReadCorpus(t *testing.T) {
	var a = "天气在太湖广场集结TEST"
	number := utf8.RuneCountInString(a)
	fmt.Println(number)
	fmt.Println(len(a))
}
