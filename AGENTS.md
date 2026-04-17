# gse

Go efficient multilingual NLP and text segmentation library — a Go port/extension of jieba supporting English, Chinese (Simplified/Traditional), Japanese, with DAG, HMM (Viterbi), POS tagging, BM25/TF-IDF, and embedded dictionaries.

Module: `github.com/go-ego/gse` · Go: `1.17` (`go.mod`) · CI tested against Go 1.25.x / 1.26.x on macOS/Linux/Windows.

## Build/Test/Lint Commands

- **Build**: `go build -v .`
- **Test (root package)**: `go test -v .`
- **Test (all packages)**: `go test -v ./...`
- **Test single test**: `go test -v -run TestHMM .`
- **Test single subpackage**: `go test -v ./hmm/...`
- **Coverage**: `go test -v -covermode=count -coverprofile=coverage.out`
- **Benchmark**: `go test -v -bench=. -benchmem .` (see `gse_bm_test.go`)
- **Benchmark tool**: `go run ./tools/benchmark/benchmark.go` (goroutines variant: `./tools/benchmark/goroutines/goroutines.go`)
- **Run server demo**: `go run ./tools/server/server.go`
- **Build without embedded dicts**: `go build -tags ne .` (excludes `dict_embed.go`, uses `dict_no_embed.go` stubs)
- **Format**: `gofmt -s -w .`
- **Vet**: `go vet ./...`

No Makefile / Taskfile / justfile exists. CI (`.github/workflows/go.yml`) runs only `go build -v .` + `go test -v .` (root package). CircleCI (`circle.yml`) runs `go test -v ./...` + coverage upload.

## Architecture

Root package `gse` is the public segmenter API. Subpackages under `hmm/` and `tf/` provide HMM cut, POS tagging, relevance scoring, and NLP extensions. Dictionary data lives under `data/dict/` and is `//go:embed`-ed at build time.

```
.
├── gse.go                  # Package entry; New, NewEmbed, Cut, Version
├── segmenter.go            # Segmenter struct + Segment (shortest path / Viterbi)
├── dag.go                  # DAG construction and cutDAG / cutDAGNoHMM
├── dictionary.go           # Dictionary (cedar double-array trie wrapper)
├── dict_util.go            # LoadDict / LoadDictMap / dict parsing, token freq
├── dict_1.16.go            # LoadDictEmbed, embed-aware loaders (go1.16+)
├── dict_embed.go           # //go:embed of data/dict/{zh,jp} files (build !ne)
├── dict_no_embed.go        # Empty stubs when built with -tags ne
├── seg_utils.go            # Low-level rune/token helpers
├── stop.go, trim.go        # Stop words & trim-by-pos helpers
├── token.go                # Token / Segment types
├── types/dict_file.go      # Shared LoadDictFile, BM25Setting, const types
├── crf/                    # CRF placeholder
├── tf/                     # TF extensions (nlp subpkg)
├── gonn/                   # Toy rnn/cnn stubs
├── hmm/
│   ├── hmm_seg.go          # HMM segmentation entry
│   ├── prob_emit.go, prob_trans.go, viterbi.go
│   ├── pos/                # POS tagging (Viterbi over char_state_tab)
│   ├── segment/            # hmm/segment wrapper
│   ├── stopwords/          # stop_words list + loader
│   ├── relevance/          # BM25, TF-IDF, IDF
│   └── extracker/          # TextRank, tag extractor
├── data/dict/
│   ├── en/dict.txt
│   ├── zh/{s_1,t_1,idf,tf_idf,stop_tokens,stop_word}.txt
│   └── jp/dict.txt
├── examples/               # Runnable example main.go (en, jp, hmm, dict, embed)
├── testdata/               # Test fixtures (en + zh/*)
├── tools/
│   ├── benchmark/          # single + goroutines benchmark
│   └── server/             # JSON RPC server demo + static UI
└── .github/workflows/go.yml
```

## Code Style

- **Package naming**: short, lowercase (`gse`, `hmm`, `pos`, `relevance`).
- **File naming**: `snake_case.go`. Tests use `_test.go`; benchmarks live in `gse_bm_test.go`.
- **Imports**: grouped stdlib → third-party → internal (`github.com/go-ego/gse/...`), separated by a blank line (see `hmm/relevance/bm25.go`, `dict_util.go`).
- **Copyright header**: every source file starts with an Apache-2.0 header. New files must include it — either the "ego authors" short form (see `gse.go`, `segmenter.go`) or the "go-ego Project Developers" long form (see `CONTRIBUTING.md § Copyright`, `dictionary.go`).
- **Doc comments**: every exported identifier has a `// Name ...` godoc comment. Use block `/* Package xxx ... */` at the top of the package-entry file.
- **Error handling**: return `error` as last value; never discard with `_ = err`. Load-dict APIs return and log via `log.Println`; examples check and print.
- **Receivers**: pointer receivers on `*Segmenter`, `*Dictionary`, etc. Keep consistent per type.
- **Zero-value friendly**: `Segmenter` is usable as `var seg gse.Segmenter` then `seg.LoadDict(...)`. Preserve this.
- **Types package** (`types/`): shared DTOs/consts (`LoadDictFile`, `BM25Setting`, `LoadDictType*` iota constants). Put cross-subpackage constants here rather than duplicating.
- Prefer `any` over `interface{}` for new code (Go 1.18+).

## Testing

- **Framework**: stdlib `testing` + assertion helpers from `github.com/vcaesar/tt` (`tt.Equal`, `tt.Nil`, `tt.Bool`). Match existing assertion style.
- **Location**: tests sit next to the code they cover (`gse_test.go`, `segmenter_test.go`, `dict_1.16_test.go`, `token_test.go`, `hmm/hmm_seg_test.go`, `hmm/pos/pos_seg_test.go`, `hmm/relevance/*_test.go`).
- **Shared fixture**: root tests share a package-level `prodSeg` segmenter initialized in `gse_test.go`'s `init()` via `prodSeg.LoadDict(); prodSeg.LoadStop("zh")`. New root-package tests should reuse `prodSeg` rather than reloading the full dict (~587K tokens, slow).
- **Test data**: `./testdata/` for root; per-subpackage testdata kept local. Examples use `./examples/.../testdata/`.
- **Single test**: `go test -v -run TestAnalyze .` (root); `go test -v -run TestBM25 ./hmm/relevance`.
- **Benchmarks**: `go test -bench=. -benchmem .` — see `gse_bm_test.go`.
- **Assertions expect exact token counts / stringified slices** (e.g. `tt.Equal(t, 23, len(s))`). Changes to default dict/frequency tables WILL break these — update tests deliberately, not casually.

## Key Patterns

- **Build tags**:
  - `dict_embed.go` uses `//go:build go1.16 && !ne` — embeds ~several-MB dictionaries into the binary.
  - `dict_no_embed.go` uses `//go:build ne` — empty stubs for size-sensitive builds. Use `-tags ne` to opt out of embedding. Keep both files in sync when adding new embedded vars.
  - gopls will report "No packages found" on `dict_no_embed.go` without the `ne` tag — this is expected, not a bug.
- **Two constructors**:
  - `gse.New(files...)` → loads from disk via `LoadDict`.
  - `gse.NewEmbed(dict...)` → loads from embedded/in-memory strings via `LoadDictEmbed`.
  - Second arg `"alpha"` (or `"en"` for Embed) enables `AlphaNum` mode.
- **Dict file selectors**: `LoadDict("zh")`, `"zh_s"`, `"zh_t"`, `"jp"`, `"en"` resolve to embedded/bundled dicts; anything else is treated as a file path. Comma-separated lists load multiple dicts in order.
- **HMM**: `seg.Cut(text, true)` enables HMM fallback for OOV tokens; `seg.Cut(text, false)` uses DAG only; `seg.Cut(text)` uses shortest-path `Slice`.
- **Concurrency**: `dict_util.go` guards the global `ToLower` with `toLowerMu sync.RWMutex`. Avoid introducing unsynchronized package-level mutable state.
- **Trie**: `Dictionary` wraps `github.com/vcaesar/cedar` (double-array trie). Do not replace without strong reason — public API exposes `Dict.Tokens`, `Dict.TotalFreq()`, etc.
- **Version constant**: `gse.Version` in `gse.go` — bump on releases.
- **Do not commit** generated dicts listed in `.gitignore`: `data/dict/zh/dict.txt`, `data/dict/zh/dictionary.txt`, vendored deps, example binaries.
- **Branching** (`CONTRIBUTING.md`): `master` is tip; release branches are `vX.Y.Z`. Bugfixes for a release go to the release branch, then forward-port.

## Dependencies

- `github.com/vcaesar/cedar` — double-array trie backing `Dictionary`.
- `github.com/vcaesar/tt` — test assertion helpers (test-only usage, but declared in main `require` block).
- Standard library only otherwise (`regexp`, `unicode`, `unicode/utf8`, `bufio`, `embed`, `math`, `sync`, `log`).

External integrations (not deps, just referenced): `go-gse-elastic`, `gse-bleve`, `gse-bind`.
