//go:build go1.16
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

// NewEmbed return new gse segmenter by embed dictionary
func NewEmbed(dict ...string) (seg Segmenter, err error) {
	if len(dict) > 1 && (dict[1] == "alpha" || dict[1] == "en") {
		seg.AlphaNum = true
	}

	err = seg.LoadDictEmbed(dict...)
	return
}

// LoadDictEmbed load dictionary by embed file
func (seg *Segmenter) LoadDictEmbed(dict ...string) (err error) {
	if len(dict) > 0 {
		d := dict[0]
		if strings.Contains(d, ", ") {
			begin := 0
			s := strings.Split(d, ", ")
			if strings.Contains(d, "zh,") {
				begin = 1
				err = seg.LoadDictStr(dataDict)
			}

			for i := begin; i < len(s); i++ {
				err = seg.LoadDictStr(s[i])
			}
			return
		}

		err = seg.LoadDictStr(d)
		return
	}

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
		text := strings.TrimSpace(s1[0])

		freqText := ""
		if len(s1) > 1 {
			freqText = strings.TrimSpace(s1[1])
		}

		frequency := seg.Size(size, text, freqText)
		if frequency == 0.0 {
			continue
		}

		pos := ""
		if size > 2 {
			pos = strings.TrimSpace(strings.Trim(s1[2], "\n"))
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
func (seg *Segmenter) LoadStopEmbed(dict ...string) (err error) {
	if len(dict) > 0 {
		d := dict[0]
		if strings.Contains(d, ", ") {
			begin := 0
			s := strings.Split(d, ", ")
			if strings.Contains(d, "zh,") {
				begin = 1
				err = seg.LoadStopStr(stopDict)
			}

			for i := begin; i < len(s); i++ {
				err = seg.LoadStopStr(s[i])
			}
			return
		}

		err = seg.LoadStopStr(d)
		return
	}

	return seg.LoadStopStr(stopDict)
}

// LoadDictStr load the stop dictionary from string
func (seg *Segmenter) LoadStopStr(dict string) error {
	if seg.StopWordMap == nil {
		seg.StopWordMap = make(map[string]bool)
	}

	arr := strings.Split(dict, "\n")
	for i := 0; i < len(arr); i++ {
		key := strings.TrimSpace(arr[i])
		if key != "" {
			seg.StopWordMap[key] = true
		}
	}

	return nil
}
