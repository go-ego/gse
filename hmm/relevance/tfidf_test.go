package relevance_test

import (
	"fmt"
	"testing"

	"github.com/vcaesar/tt"

	"github.com/go-ego/gse/hmm/extracker"
)

var (
	textTFIDF = "油价的下跌将刺激汽车新一轮消费，增强消费者的购车欲望"
)

func TestExtTFIDFAndRank(t *testing.T) {
	var te extracker.TagExtracter
	te.WithGse(segs)
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
	// output:
	// segments:  5 [{消费者 135.35978394451678} {汽车 132.5431762274668} {消费 99.74972568967256} {增强 96.4479152517576} {下跌 62.99878978351253}]
	// results:  [{消费 1} {刺激 0.5486451492724487} {下跌 0.4311204839551169} {汽车 0.4095437392771989} {购车 0.4064546007671519}]
}
