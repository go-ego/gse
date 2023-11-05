package idf

import (
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/segment"
)

// TagExtracter is extract tags struct.
type TagExtracter struct {
	seg gse.Segmenter

	Idf      *Idf
	stopWord *StopWord
}

// WithGse register the gse segmenter
func (t *TagExtracter) WithGse(segs gse.Segmenter) {
	t.stopWord = NewStopWord()
	t.seg = segs
}

// LoadDict load and create a new dictionary from the file
func (t *TagExtracter) LoadDict(fileName ...string) error {
	t.stopWord = NewStopWord()
	return t.seg.LoadDict(fileName...)
}

// LoadIdf load and create a new Idf dictionary from the file.
func (t *TagExtracter) LoadIdf(fileName ...string) error {
	t.Idf = NewIdf()
	return t.Idf.LoadDict(fileName...)
}

// LoadIdfStr load and create a new Idf dictionary from the string.
func (t *TagExtracter) LoadIdfStr(str string) error {
	t.Idf = NewIdf()
	return t.Idf.seg.LoadDictStr(str)
}

// LoadStopWords load and create a new StopWord dictionary from the file.
func (t *TagExtracter) LoadStopWords(fileName ...string) error {
	t.stopWord = NewStopWord()
	return t.stopWord.LoadDict(fileName...)
}

// ExtractTags extract the topK key words from text.
func (t *TagExtracter) ExtractTags(text string, topK int) (tags segment.Segments) {
	freqMap := make(map[string]float64)

	for _, w := range t.seg.Cut(text, true) {
		w = strings.TrimSpace(w)
		if utf8.RuneCountInString(w) < 2 {
			continue
		}
		if t.stopWord.IsStopWord(w) {
			continue
		}

		if f, ok := freqMap[w]; ok {
			freqMap[w] = f + 1.0
		} else {
			freqMap[w] = 1.0
		}
	}

	total := 0.0
	for _, freq := range freqMap {
		total += freq
	}

	for k, v := range freqMap {
		freqMap[k] = v / total
	}

	ws := make(segment.Segments, 0)
	var s segment.Segment
	for k, v := range freqMap {
		if freq, _, ok := t.Idf.Freq(k); ok {
			s = segment.Segment{Text: k, Weight: freq * v}
		} else {
			s = segment.Segment{Text: k, Weight: t.Idf.median * v}
		}
		ws = append(ws, s)
	}

	sort.Sort(sort.Reverse(ws))

	if len(ws) > topK {
		tags = ws[:topK]
		return
	}

	tags = ws
	return
}
