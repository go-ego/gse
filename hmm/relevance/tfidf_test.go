package relevance_test

import (
	"fmt"
	"testing"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/extracker"
	"github.com/vcaesar/tt"
)

var (
	segsTFIDF, _ = gse.New()
	textTFIDF    = "油价的下跌将刺激汽车新一轮消费，增强消费者的购车欲望"
)

func TestExtTFIDFAndRank(t *testing.T) {
	var te extracker.TagExtracter
	te.WithGse(segsTFIDF)
	err := te.LoadTFIDF()
	tt.Nil(t, err)
	err = te.LoadStopWords()
	tt.Nil(t, err)

	segments := te.ExtractTags(textTFIDF, 5)
	fmt.Println("segments: ", len(segments), segments)

	var tr extracker.TextRanker
	tr.WithGse(segs)

	results := tr.TextRank(textTFIDF, 5)
	fmt.Println("results: ", results)
}
