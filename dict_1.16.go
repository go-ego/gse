// Copyright 2016 The go-ego Project Developers.
//
// See the COPYRIGHT file at the top-level directory of this distribution and at
// https://github.com/go-ego/gse/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0>
//
// This file may not be copied, modified, or distributed
// except according to those terms.

package gse

import (
	"strings"

	"github.com/go-ego/gse/types"
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
	seg.loadDictStr(zhS, zhT)
	seg.CalcToken()
	return nil
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

// LoadDictStr load the dictionary from dict path
func (seg *Segmenter) LoadDictStr(dict string) error {
	seg.loadDictStr(dict)
	seg.CalcToken()
	return nil
}

// loadDictStr adds the dict tokens without calculating the token segments.
// The dicts are not concatenated: the token pos strings are substrings of
// the dict, so a concatenated copy would stay alive with the dictionary.
func (seg *Segmenter) loadDictStr(dicts ...string) {
	if seg.Dict == nil {
		seg.Dict = NewDict()
		seg.Init()
	}

	lines := 0
	for _, dict := range dicts {
		lines += strings.Count(dict, "\n") + 1
	}
	seg.Dict.grow(lines)

	sep := seg.DictSep + " "
	for _, dict := range dicts {
		for {
			line, rest, more := strings.Cut(dict, "\n")
			size, text, freqText, pos := splitDictLine(line, sep)
			text = strings.TrimSpace(text)
			freq := seg.Size(size, text, strings.TrimSpace(freqText))
			if freq != 0.0 {
				// add the words to the token
				seg.Dict.AddToken(seg.ToToken(text, freq, strings.TrimSpace(pos)))
			}

			if !more {
				break
			}
			dict = rest
		}
	}
}

// splitDictLine splits a dict line into its first three fields without
// allocating, size is the number of fields found (capped at 3)
func splitDictLine(line, sep string) (size int, text, freqText, pos string) {
	var ok bool
	text, line, ok = strings.Cut(line, sep)
	if !ok {
		return 1, text, "", ""
	}

	freqText, line, ok = strings.Cut(line, sep)
	if !ok {
		return 2, text, freqText, ""
	}

	pos, _, _ = strings.Cut(line, sep)
	return 3, text, freqText, pos
}

// LoadTFIDFDictStr load the TFIDF dictionary from dict path
func (seg *Segmenter) LoadTFIDFDictStr(dictFile *types.LoadDictFile) error {
	if seg.Dict == nil {
		seg.Dict = NewDict()
		seg.Init()
	}

	dict := dictFile.FilePath
	seg.Dict.grow(strings.Count(dict, "\n") + 1)
	sep := seg.DictSep + " "
	for {
		line, rest, more := strings.Cut(dict, "\n")
		size, text, freqText, inverseFreqText := splitDictLine(line, sep)
		text = strings.TrimSpace(text)
		// frequency
		freq := seg.Size(size, text, strings.TrimSpace(freqText))
		// invserse frequency
		inverseFreq := seg.Size(size, text, inverseFreqText)
		if freq != 0.0 && inverseFreq != 0.0 {
			// add the words to the token
			token := seg.ToToken(text, freq)
			token.inverseFreq = inverseFreq
			seg.Dict.AddToken(token)
		}

		if !more {
			break
		}
		dict = rest
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

// LoadStopStr load the stop dictionary from dict path
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
