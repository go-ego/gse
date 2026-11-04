# gse

Go による効率的な多言語 NLP およびテキスト分割ライブラリ。英語、中国語、日本語などをサポートします。
[riot](https://github.com/vcaesar/riot)、[zedsearch](https://github.com/vcaesar/zedsearch)、[zincsearch](https://github.com/zincsearch/zincsearch)、[elasticsearch](https://github.com/vcaesar/go-gse-elastic)、[bleve](https://github.com/vcaesar/gse-bleve) との連携もサポートしています。

<!--<img align="right" src="https://raw.githubusercontent.com/go-ego/ego/master/logo.jpg">-->
<!--<a href="https://circleci.com/gh/go-ego/ego/tree/dev"><img src="https://img.shields.io/circleci/project/go-ego/ego/dev.svg" alt="Build Status"></a>-->

[![Build Status](https://github.com/go-ego/gse/workflows/Go/badge.svg)](https://github.com/go-ego/gse/commits/master)
[![Lint Status](https://github.com/go-ego/gse/workflows/Lint/badge.svg)](https://github.com/go-ego/gse/actions/workflows/lint.yml)
[![CircleCI Status](https://circleci.com/gh/go-ego/gse.svg?style=shield)](https://circleci.com/gh/go-ego/gse)
[![codecov](https://codecov.io/gh/go-ego/gse/branch/master/graph/badge.svg)](https://codecov.io/gh/go-ego/gse)
[![Go Reference](https://pkg.go.dev/badge/github.com/go-ego/gse.svg)](https://pkg.go.dev/github.com/go-ego/gse)
[![GitHub release](https://img.shields.io/github/release/go-ego/gse.svg)](https://github.com/go-ego/gse/releases/latest)
<a href="https://discord.gg/535uZPakJw"><img src="https://img.shields.io/discord/1484658282777018551.svg?logo=discord&logoColor=white&label=Discord&color=5865F2" alt="Discord チャットに参加 https://discord.gg/535uZPakJw"></a>

[English](https://github.com/go-ego/gse/blob/master/README.md) | [简体中文](https://github.com/go-ego/gse/blob/master/README_zh.md)

Gse は jieba の Golang 実装であり、NLP サポートやより多くの機能を追加しています。

現在 [Codg](https://github.com/vcaesar/codg) を開発中です。手軽にコーディングと作業ができる AI エージェントシステム：自動、非同期、並行、効率的かつ高精度

<p align="center">
<a href="https://github.com/vcaesar/codg" rel="nofollow">
<!-- <img width="800" alt="Codg Demo" src="https://github.com/vcaesar/codg/raw/main/demo/26-04.png" /> -->
<img width="800" alt="Codg Demo" src="https://github.com/vcaesar/codg/raw/main/demo/26-04-1.png" />
</a>
</p>

## 機能:

- 一般、検索エンジン、完全、精密、HMM モードなど複数の分かち書きモードをサポート
- ユーザー辞書と埋め込み辞書、品詞タグ付け、分割情報の解析、ストップワードと単語のトリムをサポート
- 多言語サポート：英語、中国語、日本語など
- 繁体字中国語をサポート（単一の辞書を使用）
- Viterbi アルゴリズムを使った HMM テキスト分割をサポート
- Pytorch による NLP サポート（開発中）
- 固有表現認識（開発中）
- [riot](https://github.com/vcaesar/riot) と [zedsearch](https://github.com/vcaesar/zedsearch)、さらに [elasticsearch](https://github.com/vcaesar/go-gse-elastic)、[bleve](https://github.com/vcaesar/gse-bleve) などをサポート
- <a href="https://github.com/go-ego/gse/blob/master/tools/server/server.go">JSON RPC サービス</a>の実行が可能

## アルゴリズム:

- [辞書](https://github.com/go-ego/gse/blob/master/dictionary.go)はデフォルトで Double Array Trie（ダブル配列トライ、[cedar](https://github.com/vcaesar/cedar)）、`-tags hash` 指定時は rune 単位のハッシュトライで実装
- [分割器](https://github.com/go-ego/gse/blob/master/dag.go)アルゴリズムは最短パス（単語頻度と動的プログラミングに基づく）、DAG および HMM アルゴリズムによる分かち書き

## テキスト分割速度:

- <a href="https://github.com/go-ego/gse/blob/master/tools/benchmark/benchmark.go">シングルスレッド</a> 9.2MB/s
- <a href="https://github.com/go-ego/gse/blob/master/tools/benchmark/goroutines/goroutines.go">ゴルーチン並行</a> 26.8MB/s
- HMM テキスト分割シングルスレッド 3.2MB/s（2コア 4スレッド Macbook Pro）

## バインディング:

[gse-bind](https://github.com/vcaesar/gse-bind)、JavaScript やその他の言語をバインドし、より多くの言語をサポートします。

## インストール / アップデート

Go モジュールサポート（Go 1.11+）では、単純にインポートするだけです：

```go
import "github.com/go-ego/gse"
```

それ以外の場合、gse パッケージをインストールするには以下のコマンドを実行します：

```
go get -u github.com/go-ego/gse
```

### ビルドタグ

- `-tags hash`：辞書にデフォルトの cedar ダブル配列トライではなく、rune 単位ハッシュトライを使用します。中国語辞書のトライメモリを約 8 MiB（約 40%）削減しますが、分かち書きは約 10% 低速になります。
- `-tags ne`：デフォルト辞書をバイナリに埋め込みません（サイズ削減）。`LoadDictEmbed()` ではなく `LoadDict("辞書のパス")` で読み込んでください。

```
go build -tags hash
go build -tags ne
```

### 低メモリデバイス

デフォルトの中国語辞書はヒープを約 132 MiB 使用します。低メモリデバイスでは以下を組み合わせてください：

- `-tags hash`：ハッシュトライを使用し、約 8 MiB 削減。
- 読み込み前に `seg.SkipSubSeg = true`：検索モード（`Slice`、`String`、`Pos` の `searchMode`）用に事前計算するサブワードを省略し、約 50 MiB 削減。`CutSearch(text)` も影響を受けますが、`CutSearch(text, true)` と `CutSearch(text, false)` は引き続き使えます。
- `seg.LoadDict("zh_s")`：簡体字中国語の辞書のみを読み込みます。

```go
// go build -tags hash
var seg gse.Segmenter
seg.SkipSubSeg = true
err := seg.LoadDict("zh_s") // 約 53 MiB
```

### Lint

CI では [golangci-lint](https://golangci-lint.run) v2 でコードを検査します。設定は [.golangci.yml](.golangci.yml) を参照してください：

```
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
golangci-lint run ./...
golangci-lint run --build-tags hash .
```

## 使用方法

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
	// 元の大文字を保持
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

例2:

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
	// デフォルト辞書の読み込み
	seg.LoadDict()
	// 埋め込みデフォルト辞書の読み込み
	// seg.LoadDictEmbed()
	//
	// 簡体字中国語辞書の読み込み
	// seg.LoadDict("zh_s")
	// seg.LoadDictEmbed("zh_s")
	//
	// 繁体字中国語辞書の読み込み
	// seg.LoadDict("zh_t")
	//
	// 日本語辞書の読み込み
	// seg.LoadDict("jp")
	//
	// 辞書の読み込み
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
	// テキスト分割
	tb := []byte(text)
	fmt.Println(seg.String(text, true))

	segments := seg.Segment(tb)
	// 分かち書き結果の処理、検索モード
	fmt.Println(gse.ToString(segments, true))
}

```

[カスタム辞書の例を見る](/examples/dict/main.go)

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

[中国語の例を見る](/examples/main.go)

[日本語の例を見る](/examples/jp/main.go)

## Elasticsearch

elasticsearch と一緒に使用するには？

[go-gse-elastic](https://github.com/vcaesar/go-gse-elastic)

## 作者

- [メンテナー](https://github.com/orgs/go-ego/people)
- [貢献者](https://github.com/go-ego/gse/graphs/contributors)

## ライセンス

Gse は主に「Apache License (Version 2.0)」の条項の下で配布されています。
[LICENSE-APACHE](http://www.apache.org/licenses/LICENSE-2.0)、[LICENSE](https://github.com/go-vgo/robotgo/blob/master/LICENSE) を参照してください。

[jieba](https://github.com/fxsjy/jieba)（[jiebago](https://github.com/wangbin/jiebago)）と [sego](https://github.com/huichen/sego) に感謝します。
