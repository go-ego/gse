# [gse](https://github.com/go-ego/gse)

Go 高性能多语言 NLP 和分词, 支持英文、中文、日文等。
支持接入 [riot](https://github.com/vcaesar/riot)、[zedsearch](https://github.com/vcaesar/zedsearch)、[zincsearch](https://github.com/zincsearch/zincsearch)、[elasticsearch](https://github.com/vcaesar/go-gse-elastic) 和 [bleve](https://github.com/vcaesar/gse-bleve)。

<!--<img align="right" src="https://raw.githubusercontent.com/go-ego/ego/master/logo.jpg">-->
<!--<a href="https://circleci.com/gh/go-ego/ego/tree/dev"><img src="https://img.shields.io/circleci/project/go-ego/ego/dev.svg" alt="Build Status"></a>-->

[![Build Status](https://github.com/go-ego/gse/workflows/Go/badge.svg)](https://github.com/go-ego/gse/commits/master)
[![Lint Status](https://github.com/go-ego/gse/workflows/Lint/badge.svg)](https://github.com/go-ego/gse/actions/workflows/lint.yml)
[![CircleCI Status](https://circleci.com/gh/go-ego/gse.svg?style=shield)](https://circleci.com/gh/go-ego/gse)
[![codecov](https://codecov.io/gh/go-ego/gse/branch/master/graph/badge.svg)](https://codecov.io/gh/go-ego/gse)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-ego/gse.svg)](https://pkg.go.dev/github.com/go-ego/gse)
[![GitHub release](https://img.shields.io/github/release/go-ego/gse.svg)](https://github.com/go-ego/gse/releases/latest)
<a href="https://discord.gg/npPb3NzE4A"><img src="https://img.shields.io/discord/1484658282777018551.svg?logo=discord&logoColor=white&label=Discord&color=5865F2" alt="加入 Discord 聊天 https://discord.gg/535uZPakJw"></a>

[English](https://github.com/go-ego/gse/blob/master/README.md) | [日本語](https://github.com/go-ego/gse/blob/master/README_ja.md)

Gse 是结巴分词(jieba)的 golang 实现, 并添加 NLP 功能和更多特性

我目前正在开发 [Codg](https://github.com/vcaesar/codg), 简单易用的编码和工作 AI agent 系统: 自动、异步、并发、高效且高准确率

<p align="center">
<a href="https://github.com/vcaesar/codg" rel="nofollow">
<!-- <img width="800" alt="Codg Demo" src="https://github.com/vcaesar/codg/raw/main/demo/26-04.png" /> -->
<img width="800" alt="Codg Demo" src="https://github.com/vcaesar/codg/raw/main/demo/26-04-1.png" />
</a>
</p>

## 特征:

- 支持普通、搜索引擎、全模式、精确模式和 HMM 模式多种分词模式
- 支持自定义词典、embed 词典、词性标注、分词信息分析、停用词和整理分词
- 多语言支持: 英文, 中文, 日文等
- 支持繁体中文 (使用单一词典)
- 支持使用 Viterbi 算法的 HMM 分词
- 基于 Pytorch 的 NLP 支持 (进行中)
- 命名实体识别 (进行中)
- 支持接入 [riot](https://github.com/vcaesar/riot) 和 [zedsearch](https://github.com/vcaesar/zedsearch), 以及 [elasticsearch](https://github.com/vcaesar/go-gse-elastic)、[bleve](https://github.com/vcaesar/gse-bleve) 等
- 可运行<a href="https://github.com/go-ego/gse/blob/master/tools/server/server.go"> JSON RPC 服务</a>

## 算法:

- [词典](https://github.com/go-ego/gse/blob/master/dictionary.go)默认用双数组 trie（Double-Array Trie, [cedar](https://github.com/vcaesar/cedar)）实现, 使用 `-tags hash` 时用 rune 级哈希 trie 实现
- [分词器](https://github.com/go-ego/gse/blob/master/dag.go)算法为基于词频的最短路径加动态规划, 以及 DAG 和 HMM 算法分词

## 分词速度:

- <a href="https://github.com/go-ego/gse/blob/master/tools/benchmark/benchmark.go">单线程</a> 9.2MB/s
- <a href="https://github.com/go-ego/gse/blob/master/tools/benchmark/goroutines/goroutines.go">goroutines 并发</a> 26.8MB/s
- HMM 模式单线程分词速度 3.2MB/s（双核 4 线程 Macbook Pro）

## Binding:

[gse-bind](https://github.com/vcaesar/gse-bind), 绑定 JavaScript 等, 支持更多语言。

## 安装/更新

支持 Go module (Go 1.11+) 时, 直接导入即可:

```go
import "github.com/go-ego/gse"
```

否则, 运行以下命令安装 gse 包:

```
go get -u github.com/go-ego/gse
```

### 编译标签 (Build tags)

- `-tags hash`: 词典使用 rune 级哈希 trie, 替代默认的 cedar 双数组 trie; 中文词典的 trie 内存约减少 8 MiB (约 40%), 但分词约慢 10%。
- `-tags ne`: 不将默认词典嵌入二进制文件 (体积更小); 需使用 `LoadDict("你的词典路径")` 加载, 而非 `LoadDictEmbed()`。

```
go build -tags hash
go build -tags ne
```

### 低内存设备

默认的中文词典约占用 132 MiB 堆内存。低内存设备可组合使用:

- `-tags hash`: 使用哈希 trie, 约减少 8 MiB。
- 加载前设置 `seg.SkipSubSeg = true`: 跳过搜索模式 (`Slice`, `String`, `Pos` 的 `searchMode`) 预先计算的子词, 约减少 50 MiB; `CutSearch(text)` 也会受影响, `CutSearch(text, true)` 和 `CutSearch(text, false)` 不受影响。
- `seg.LoadDict("zh_s")`: 只加载简体中文词典。

```go
// go build -tags hash
var seg gse.Segmenter
seg.SkipSubSeg = true
err := seg.LoadDict("zh_s") // 约 53 MiB
```

### 代码检查 (Lint)

CI 使用 [golangci-lint](https://golangci-lint.run) v2 检查代码, 配置见 [.golangci.yml](.golangci.yml):

```
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
golangci-lint run ./...
golangci-lint run --build-tags hash .
```

## 使用

```go
package main

import (
	_ "embed"
	"fmt"

	"github.com/go-ego/gse"
)

//go:embed testdata/test_en2.txt
var testDict string

//go:embed testdata/test_en.txt
var testEn string

var (
	text  = "To be or not to be, that's the question!"
	test1 = "Hiworld, Helloworld!"
)

func main() {
	// 保留原始大写字母
	// gse.ToLower = false

	var seg1 gse.Segmenter
	seg1.DictSep = ","
	err := seg1.LoadDict("./testdata/test_en.txt")
	if err != nil {
		fmt.Println("Load dictionary error: ", err)
	}

	s1 := seg1.Cut(text)
	fmt.Println("seg1 Cut: ", s1)
	// seg1 Cut:  [to be   or   not to be ,   that's the question!]

	var seg2 gse.Segmenter
	seg2.AlphaNum = true
	seg2.LoadDict("./testdata/test_en_dict3.txt")

	s2 := seg2.Cut(test1)
	fmt.Println("seg2 Cut: ", s2)
	// seg2 Cut:  [hi world ,   hello world !]

	var seg3 gse.Segmenter
	seg3.AlphaNum = true
	seg3.DictSep = ","
	err = seg3.LoadDictEmbed(testDict + "\n" + testEn)
	if err != nil {
		fmt.Println("loadDictEmbed error: ", err)
	}
	s3 := seg3.Cut(text + test1)
	fmt.Println("seg3 Cut: ", s3)
	// seg3 Cut:  [to be   or   not to be ,   that's the question! hi world ,   hello world !]

	// example2()
}
```

示例2:

```go
package main

import (
	"fmt"
	"regexp"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/pos"
)

var (
	text = "Hello world, Helloworld. Winter is coming! こんにちは世界, 你好世界."

	new, _ = gse.New("zh,testdata/test_en_dict3.txt", "alpha")

	seg gse.Segmenter
	posSeg pos.Segmenter
)

func main() {
	// 加载默认词典
	seg.LoadDict()
	// 加载默认 embed 词典
	// seg.LoadDictEmbed()
	//
	// 加载简体中文词典
	// seg.LoadDict("zh_s")
	// seg.LoadDictEmbed("zh_s")
	//
	// 加载繁体中文词典
	// seg.LoadDict("zh_t")
	//
	// 加载日文词典
	// seg.LoadDict("jp")
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
	fmt.Println("analyze: ", new.Analyze(hmm, text))

	hmm = new.CutAll(text)
	fmt.Println("cut all: ", hmm)

	reg := regexp.MustCompile(`(\d+年|\d+月|\d+日|[\p{Latin}]+|[\p{Hangul}]+|\d+\.\d+|[a-zA-Z0-9]+)`)
	text1 := `헬로월드 헬로 서울, 2021年09月10日, 3.14`
	hmm = seg.CutDAG(text1, reg)
	fmt.Println("Cut with hmm and regexp: ", hmm, hmm[0], hmm[6])
}

func analyzeAndTrim(cut []string) {
	a := seg.Analyze(cut, "")
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

	pos.WithGse(seg)
	po = posSeg.Cut(text, true)
	fmt.Println("pos: ", po)

	po = posSeg.TrimWithPos(po, "zg")
	fmt.Println("trim pos: ", po)
}

func segCut() {
	// 分词文本
	tb := []byte(text)
	fmt.Println(seg.String(text, true))

	segments := seg.Segment(tb)
	// 处理分词结果, 搜索模式
	fmt.Println(gse.ToString(segments, true))
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

//go:embed test_en_dict3.txt
var testDict string

func main() {
	// var seg gse.Segmenter
	// seg.LoadDict("zh, testdata/zh/test_dict.txt, testdata/zh/test_dict1.txt")
	// seg.LoadStop()
	seg, err := gse.NewEmbed("zh, word 20 n"+testDict, "en")
	// seg.LoadDictEmbed()
	seg.LoadStopEmbed()

	text1 := "Hello world, こんにちは世界, 你好世界!"
	s1 := seg.Cut(text1, true)
	fmt.Println(s1)
	fmt.Println("trim: ", seg.Trim(s1))
	fmt.Println("stop: ", seg.Stop(s1))
	fmt.Println(seg.String(text1, true))

	segments := seg.Segment([]byte(text1))
	fmt.Println(gse.ToString(segments))
}
```

[中文分词示例](/examples/main.go)

[日文分词示例](/examples/jp/main.go)

## Elasticsearch

如何与 elasticsearch 一起使用?

[go-gse-elastic](https://github.com/vcaesar/go-gse-elastic)

## 作者

- [维护者](https://github.com/orgs/go-ego/people)
- [贡献者](https://github.com/go-ego/gse/graphs/contributors)

## 许可证

Gse 主要基于 "Apache License (Version 2.0)" 条款分发。
参见 [LICENSE-APACHE](http://www.apache.org/licenses/LICENSE-2.0), [LICENSE](https://github.com/go-vgo/robotgo/blob/master/LICENSE)。

感谢 [jieba](https://github.com/fxsjy/jieba)([jiebago](https://github.com/wangbin/jiebago)) 和 [sego](https://github.com/huichen/sego)。
