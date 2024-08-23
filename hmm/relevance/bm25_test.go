package relevance_test

import (
	"fmt"
	"testing"

	"github.com/vcaesar/tt"

	"github.com/go-ego/gse/hmm/extracker"
)

var (
	textBM25 = "这里有历史的积淀,充满活力与想象"
)

func TestExtBM25AndRank(t *testing.T) {
	var te extracker.TagExtracter
	te.WithGse(segs)
	err := te.LoadBM25(nil, nil)
	tt.Nil(t, err)
	err = te.LoadStopWords()
	tt.Nil(t, err)

	segments := te.ExtractTags(textBM25, 5)
	fmt.Println("segments: ", len(segments), segments)

	var tr extracker.TextRanker
	tr.WithGse(segs)

	results := tr.TextRank(textBM25, 5)
	fmt.Println("results: ", results)
	// output:
	// segments:  5 [{想象 13.489829905298084} {活力 12.86320693643856} {充满 12.480977334559475} {这里 9.56153393671824} {历史 8.738605467373437}]
	// results:  [{积淀 1} {活力 0.7380261680439799} {有 0.6602549059736358} {历史 0.6573229314364966} {想象 0.39804353825110805}]
}
