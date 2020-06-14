package idf

import (
	"github.com/go-ego/gse"
)

// StopWordMap default contains some stop words.
var StopWordMap = map[string]int{
	"the":   1,
	"of":    1,
	"is":    1,
	"and":   1,
	"to":    1,
	"in":    1,
	"that":  1,
	"we":    1,
	"for":   1,
	"an":    1,
	"are":   1,
	"by":    1,
	"be":    1,
	"as":    1,
	"on":    1,
	"with":  1,
	"can":   1,
	"if":    1,
	"from":  1,
	"which": 1,
	"you":   1,
	"it":    1,
	"this":  1,
	"then":  1,
	"at":    1,
	"have":  1,
	"all":   1,
	"not":   1,
	"one":   1,
	"has":   1,
	"or":    1,
}

// StopWord is a dictionary for all stop words.
type StopWord struct {
	stopWordMap map[string]int

	seg gse.Segmenter
}

// AddToken adds a token into StopWord dictionary.
func (s *StopWord) AddToken(text string) {
	s.stopWordMap[text] = 1
}

// NewStopWord create a new StopWord with default stop words.
func NewStopWord() *StopWord {
	s := new(StopWord)
	s.stopWordMap = StopWordMap
	return s
}

// IsStopWord checks if a given word is stop word.
func (s *StopWord) IsStopWord(word string) bool {
	_, ok := s.stopWordMap[word]
	return ok
}

func (s *StopWord) loadDict(fileName ...string) error {
	return s.seg.LoadDict(fileName...)
}
