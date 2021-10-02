# [gse](https://github.com/go-ego/gse)

Go 高性能多语言 NLP 和分词, 支持英文、中文、日文等, 支持接入 [elasticsearch](https://github.com/vcaesar/go-gse-elastic) 和 bleve

<!--<img align="right" src="https://raw.githubusercontent.com/go-ego/ego/master/logo.jpg">-->
<!--<a href="https://circleci.com/gh/go-ego/ego/tree/dev"><img src="https://img.shields.io/circleci/project/go-ego/ego/dev.svg" alt="Build Status"></a>-->

[![Build Status](https://github.com/go-ego/gse/workflows/Go/badge.svg)](https://github.com/go-ego/gse/commits/master)
[![CircleCI Status](https://circleci.com/gh/go-ego/gse.svg?style=shield)](https://circleci.com/gh/go-ego/gse)
[![codecov](https://codecov.io/gh/go-ego/gse/branch/master/graph/badge.svg)](https://codecov.io/gh/go-ego/gse)
[![Build Status](https://travis-ci.org/go-ego/gse.svg)](https://travis-ci.org/go-ego/gse)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-ego/gse)](https://goreportcard.com/report/github.com/go-ego/gse)
[![GoDoc](https://godoc.org/github.com/go-ego/gse?status.svg)](https://godoc.org/github.com/go-ego/gse)
[![GitHub release](https://img.shields.io/github/release/go-ego/gse.svg)](https://github.com/go-ego/gse/releases/latest)
[![Join the chat at https://gitter.im/go-ego/ego](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/go-ego/ego?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

<!--<a href="https://github.com/go-ego/ego/releases"><img src="https://img.shields.io/badge/%20version%20-%206.0.0%20-blue.svg?style=flat-square" alt="Releases"></a>-->

Gse 是结巴分词(jieba)的 golang 实现, 并尝试添加 NLP 功能和更多属性

## 特征:
- 支持普通、搜索引擎、全模式、精确模式和 HMM 模式多种分词模式
- 支持自定义词典、embed 词典、词性标注、停用词、整理分析分词、支持繁体字
- NLP 和 TensorFlow 支持 (进行中)
- 命名实体识别 (进行中)
- 支持接入 Elasticsearch 和 bleve
- 多语言支持: 英文, 中文, 日文等
- 可运行<a href="https://github.com/go-ego/gse/blob/master/server/server.go"> JSON RPC 服务</a>

## 算法: 
- [词典](https://github.com/go-ego/gse/blob/master/dictionary.go)用双数组 trie（Double-Array Trie）实现，
- [分词器](https://github.com/go-ego/gse/blob/master/segmenter.go)算法为基于词频的最短路径加动态规划, 以及 DAG 和 HMM 算法分词.
- 支持 HMM 分词, 使用 viterbi 算法.

## 分词速度:
- <a href="https://github.com/go-ego/gse/blob/master/benchmark/benchmark.go">单线程</a> 9.2MB/s
- <a href="https://github.com/go-ego/gse/blob/master/benchmark/goroutines/goroutines.go">goroutines 并发</a> 26.8MB/s. 
- HMM 模式单线程分词速度 3.2MB/s.（双核 4 线程 Macbook Pro）。

## Binding:

[gse-bind](https://github.com/vcaesar/gse-bind), binding JavaScript and other, support more language.

## 安装/更新

```
go get -u github.com/go-ego/gse
```

## 使用

```go
package main

import (
	"fmt"
	"regexp"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/pos"
)

var (
	seg gse.Segmenter
	posSeg pos.Segmenter

	new, _ = gse.New("zh,testdata/test_dict3.txt", "alpha")

	text = "你好世界, Hello world, Helloworld."
)

func main() {
	// 加载默认字典
	seg.LoadDict()
	// seg.LoadDictEmbed()
	// 
	// 载入词典
	// seg.LoadDict("your gopath"+"/src/github.com/go-ego/gse/data/dict/dictionary.txt")

	cut()

	segCut()
}


func cut() {
	hmm := new.Cut(text, true)
	fmt.Println("cut use hmm: ", hmm)

	hmm = new.CutSearch(text, true)
	fmt.Println("cut search use hmm: ", hmm)

	hmm = new.CutAll(text)
	fmt.Println("cut all: ", hmm)

	reg := regexp.MustCompile(`(\d+年|\d+月|\d+日|[\p{Latin}]+|[\p{Hangul}]+|\d+\.\d+|[a-zA-Z0-9]+)`)
	text1 := `헬로월드 헬로 서울, 2021年09月10日, 3.14`
	hmm = seg.CutDAG(text1, reg)
	fmt.Println("Cut with hmm and regexp: ", hmm, hmm[0], hmm[6])
}

func analyzeAndTrim(cut []string) {
	a := seg.Analyze(cut)
	fmt.Println("analyze the segment: ", a)

	cut = seg.Trim(cut)
	fmt.Println("cut all: ", cut)

	fmt.Println(seg.String(text, true))
	fmt.Println(seg.Slice(text, true))
}

func cutPos() {
	po := seg.Pos(text, true)
	fmt.Println("pos: ", po)
	po = seg.TrimPos(po)
	fmt.Println("trim pos: ", po)

	posSeg.WithGse(seg)
	po = posSeg.Cut(text, true)
	fmt.Println("pos: ", po)

	po = posSeg.TrimWithPos(po, "zg")
	fmt.Println("trim pos: ", po)
}

func segCut() {
	// 分词文本
	tb := []byte("山达尔星联邦共和国联邦政府")

	// 处理分词结果
	fmt.Println("输出分词结果, 类型为字符串, 使用搜索模式: ", seg.String(tb, true))
	fmt.Println("输出分词结果, 类型为 slice: ", seg.Slice(tb))

	segments := seg.Segment(tb)
	// 处理分词结果, 普通模式
	fmt.Println(gse.ToString(segments))

	segments1 := seg.Segment([]byte(text))
	// 搜索模式
	fmt.Println(gse.ToString(segments1, true))
}

```

[自定义词典分词示例](/examples/dict/main.go)

```Go

package main

import (
	"fmt"
	_ "embed"

	"github.com/go-ego/gse"
)

//go:embed test_dict3.txt
var testDict string

func main() {
	// var seg gse.Segmenter
	// seg.LoadDict("zh, testdata/test_dict.txt, testdata/test_dict1.txt")
	// seg.LoadStop()
	seg, err := gse.NewEmbed("zh, word 20 n"+testDict, "en")
	// seg.LoadDictEmbed()
	seg.LoadStopEmbed()

	text1 := "所以, 你好, 再见"
	fmt.Println(seg.Cut(text1, true))
	fmt.Println(seg.String(text1, true))

	segments := seg.Segment([]byte(text1))
	fmt.Println(gse.ToString(segments))
}
```

[中文分词示例](/examples/main.go)

[日文分词示例](/examples/jp/main.go)

## Elasticsearch
How to use it with elasticsearch?

[go-gse-elastic](https://github.com/vcaesar/go-gse-elastic)


## [Build-tools](https://github.com/go-ego/re)

```
go get -u github.com/go-ego/re
```

### re gse

创建一个新的 gse 程序

```
$ re gse my-gse
```

### re run

运行我们刚刚创建的应用程序, CD 到程序文件夹并执行:

```
$ cd my-gse && re run
```

## Authors

- [Maintainers](https://github.com/orgs/go-ego/people)
- [Contributors](https://github.com/go-ego/gse/graphs/contributors)

## License

Gse is primarily distributed under the terms of "both the MIT license and the Apache License (Version 2.0)".
See [LICENSE-APACHE](http://www.apache.org/licenses/LICENSE-2.0), [LICENSE-MIT](https://github.com/go-vgo/robotgo/blob/master/LICENSE).

Thanks for [sego](https://github.com/huichen/sego) and [jieba](https://github.com/fxsjy/jieba)([jiebago](https://github.com/wangbin/jiebago)).
