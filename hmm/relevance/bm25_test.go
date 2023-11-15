package relevance_test

import (
	"fmt"
	"testing"

	"github.com/vcaesar/tt"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/extracker"
)

var (
	segsBM25, _ = gse.New()
	textBM25    = "这里不仅有历史的积淀,还充满活力与想象"
)

func TestExtBM25AndRank(t *testing.T) {
	var te extracker.TagExtracter
	te.WithGse(segsBM25)
	err := te.LoadBM25(nil)
	tt.Nil(t, err)
	err = te.LoadStopWords()
	tt.Nil(t, err)

	segments := te.ExtractTags(textBM25, 5)
	fmt.Println("segments: ", len(segments), segments)

	var tr extracker.TextRanker
	tr.WithGse(segsBM25)

	results := tr.TextRank(textBM25, 5)
	fmt.Println("results: ", results)
}
