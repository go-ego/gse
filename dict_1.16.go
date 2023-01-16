// Copyright 2016 The go-ego Project Developers. See the COPYRIGHT
// file at the top-level directory of this distribution and at
// https://github.com/go-ego/gse/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0> or the MIT license
// <LICENSE-MIT or http://opensource.org/licenses/MIT>, at your
// option. This file may not be copied, modified, or distributed
// except according to those terms.

package gse

import (
	"strings"
)

// //go:embed data/dict/dictionary.txt
// var dataDict string

// NewEmbed return new gse segmenter by embed dictionary
func NewEmbed(dict ...string) (seg Segmenter, err error) {
	if len(dict) > 1 && (dict[1] == "alpha" || dict[1] == "en") {
		seg.AlphaNum = true
	}

	err = seg.LoadDictEmbed(dict...)
	return
}

func (seg *Segmenter) loadZh() error {
	return seg.LoadDictStr(zhS + zhT)
}

func (seg *Segmenter) loadZhST(d string) (begin int, err error) {
	if strings.Contains(d, "zh,") {
		begin = 1
		// err = seg.LoadDictStr(dataDict)
		err = seg.loadZh()
	}

	if strings.Contains(d, "zh_s,") {
		begin = 1
		err = seg.LoadDictStr(zhS)
	}
	if strings.Contains(d, "zh_t,") {
		begin = 1
		err = seg.LoadDictStr(zhT)
	}

	return
}

// LoadDictEmbed load the dictionary by embed file
func (seg *Segmenter) LoadDictEmbed(dict ...string) (err error) {
	if len(dict) > 0 {
		d := dict[0]
		if d == "ja" {
			return seg.LoadDictStr(ja)
		}

		if d == "zh" {
			return seg.loadZh()
		}
		if d == "zh_s" {
			return seg.LoadDictStr(zhS)
		}
		if d == "zh_t" {
			return seg.LoadDictStr(zhT)
		}

		if strings.Contains(d, ", ") && seg.DictSep != "," {
			begin := 0
			s := strings.Split(d, ", ")
			begin, err = seg.loadZhST(d)

			for i := begin; i < len(s); i++ {
				err = seg.LoadDictStr(s[i])
			}
			return
		}

		err = seg.LoadDictStr(d)
		return
	}

	// return seg.LoadDictStr(dataDict)
	return seg.loadZh()
}

// LoadDictStr load the dictionary from string
func (seg *Segmenter) LoadDictStr(dict string) error {
	if seg.Dict == nil {
		seg.Dict = NewDict()
		seg.Init()
	}

	arr := strings.Split(dict, "\n")
	for i := 0; i < len(arr); i++ {
		s1 := strings.Split(arr[i], seg.DictSep+" ")
		size := len(s1)
		if size == 0 {
			continue
		}
		text := strings.TrimSpace(s1[0])

		freqText := ""
		if len(s1) > 1 {
			freqText = strings.TrimSpace(s1[1])
		}

		freq := seg.Size(size, text, freqText)
		if freq == 0.0 {
			continue
		}

		pos := ""
		if size > 2 {
			pos = strings.TrimSpace(strings.Trim(s1[2], "\n"))
		}

		// add the words to the token
		words := seg.SplitTextToWords([]byte(text))
		token := Token{text: words, freq: freq, pos: pos}
		seg.Dict.AddToken(token)
	}

	seg.CalcToken()
	return nil
}

// LoadStopEmbed load the stop dictionary from embed file
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
