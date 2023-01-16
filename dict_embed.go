//go:build go1.16 && !ne
// +build go1.16,!ne

package gse

import (
	_ "embed"
)

var (
	//go:embed data/dict/jp/dict.txt
	ja string

	//go:embed data/dict/zh/t_1.txt
	zhT string
	//go:embed data/dict/zh/s_1.txt
	zhS string

	//go:embed data/dict/zh/idf.txt
	zhIdf string
)

//go:embed data/dict/zh/stop_tokens.txt
var stopDict string
