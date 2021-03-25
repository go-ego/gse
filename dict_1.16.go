// +build go1.16

package gse

import (
	_ "embed"
	"strings"
)

//go:embed data/dict/dictionary.txt
var dataDict string

//go:embed data/dict/stop_tokens.txt
var stopDict string

// LoadDictEmbed load dictionary by embed file
func (seg *Segmenter) LoadDictEmbed() error {
	return seg.LoadDictStr(dataDict)
}

// LoadDictStr load dictionary from string
func (seg *Segmenter) LoadDictStr(dict string) error {
	if seg.Dict == nil {
		seg.Dict = NewDict()
		seg.Init()
	}

	arr := strings.Split(dict, "\n")
	for i := 0; i < len(arr); i++ {
		s1 := strings.Split(arr[i], " ")
		size := len(s1)
		if size == 0 {
			continue
		}
		text := strings.Trim(s1[0], " ")

		freqText := ""
		if len(s1) > 1 {
			freqText = s1[1]
		}

		frequency := seg.Size(size, text, freqText)
		if frequency == 0.0 {
			continue
		}

		pos := ""
		if size > 2 {
			pos = strings.Trim(s1[2], "\n")
		}

		// 将分词添加到字典中
		words := seg.SplitTextToWords([]byte(text))
		token := Token{text: words, frequency: frequency, pos: pos}
		seg.Dict.addToken(token)
	}

	seg.CalcToken()
	return nil
}

// LoadStopEmbed load stop dictionary from embed file
func (seg *Segmenter) LoadStopEmbed() error {
	if seg.StopWordMap == nil {
		seg.StopWordMap = make(map[string]bool)
	}

	arr := strings.Split(stopDict, "\n")
	for i := 0; i < len(arr); i++ {
		key := strings.TrimSpace(arr[i])
		if key != "" {
			seg.StopWordMap[key] = true
		}
	}

	return nil
}
