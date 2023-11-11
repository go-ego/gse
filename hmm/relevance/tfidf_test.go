package relevance_test

import (
	"fmt"
	"testing"

	"github.com/go-ego/gse/hmm/extracker"
	"github.com/vcaesar/tt"
)

func TestExtTFIDFAndRank(t *testing.T) {
	var te extracker.TagExtracter
	te.WithGse(segs)
	err := te.LoadTFIDF()
	tt.Nil(t, err)
	err = te.LoadStopWords()
	tt.Nil(t, err)

	segments := te.ExtractTags(text, 5)
	fmt.Println("segments: ", len(segments), segments)

	var tr extracker.TextRanker
	tr.WithGse(segs)

	results := tr.TextRank(text, 5)
	fmt.Println("results: ", results)
}
