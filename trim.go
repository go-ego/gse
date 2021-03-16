// Copyright 2016 ego authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gse

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

func notPunct(ru []rune) bool {
	for i := 0; i < len(ru); i++ {
		if !unicode.IsSpace(ru[i]) && !unicode.IsPunct(ru[i]) {
			return true
		}
	}

	return false
}

// TrimPunct trim []string exclude space and punct
func (seg *Segmenter) TrimPunct(s []string) (r []string) {
	for i := 0; i < len(s); i++ {
		ru := []rune(s[i])
		if len(ru) > 0 {
			r0 := ru[0]
			if !unicode.IsSpace(r0) && !unicode.IsPunct(r0) {
				r = append(r, s[i])
			} else if len(ru) > 1 && notPunct(ru) {
				r = append(r, s[i])
			}
		}
	}

	return
}

// TrimPosPunct trim SegPos not space and punct
func (seg *Segmenter) TrimPosPunct(se []SegPos) (re []SegPos) {
	for i := 0; i < len(se); i++ {
		if !seg.NotStop && seg.IsStop(se[i].Text) {
			se[i].Text = ""
		}

		if se[i].Text != "" && len(se[i].Text) > 0 {
			ru := []rune(se[i].Text)[0]
			if !unicode.IsSpace(ru) && !unicode.IsPunct(ru) {
				re = append(re, se[i])
			}
		}
	}

	return
}

// TrimWithPos trim some seg with pos
func (seg *Segmenter) TrimWithPos(se []SegPos, pos ...string) (re []SegPos) {
	for h := 0; h < len(pos); h++ {
		if h > 0 {
			se = re
			re = nil
		}

		for i := 0; i < len(se); i++ {
			if se[i].Pos != pos[h] {
				re = append(re, se[i])
			}
		}
	}

	return
}

// Trim trim []string exclude symbol, space and punct
func (seg *Segmenter) Trim(s []string) (r []string) {
	for i := 0; i < len(s); i++ {
		si := FilterSymbol(s[i])
		if !seg.NotStop && seg.IsStop(si) {
			si = ""
		}

		if si != "" {
			r = append(r, si)
		}
	}

	return
}

// TrimSymbol trim []string exclude symbol, space and punct
func (seg *Segmenter) TrimSymbol(s []string) (r []string) {
	for i := 0; i < len(s); i++ {
		si := FilterSymbol(s[i])
		if si != "" {
			r = append(r, si)
		}
	}

	return
}

// TrimPos trim SegPos not symbol, space and punct
func (seg *Segmenter) TrimPos(s []SegPos) (r []SegPos) {
	for i := 0; i < len(s); i++ {
		si := FilterSymbol(s[i].Text)
		if !seg.NotStop && seg.IsStop(si) {
			si = ""
		}

		if si != "" {
			r = append(r, s[i])
		}
	}

	return
}

// CutTrim cut string and tirm
func (seg *Segmenter) CutTrim(str string, hmm ...bool) []string {
	s := seg.Cut(str, hmm...)
	return seg.Trim(s)
}

// PosTrim cut string pos and trim
func (seg *Segmenter) PosTrim(str string, search bool, pos ...string) []SegPos {
	p := seg.Pos(str, search)
	p = seg.TrimWithPos(p, pos...)
	return seg.TrimPos(p)
}

// PosTrimArr cut string return pos.Text []string
func (seg *Segmenter) PosTrimArr(str string, search bool, pos ...string) (re []string) {
	p1 := seg.PosTrim(str, search, pos...)
	for i := 0; i < len(p1); i++ {
		re = append(re, p1[i].Text)
	}

	return
}

// PosTrimStr cut string return pos.Text string
func (seg *Segmenter) PosTrimStr(str string, search bool, pos ...string) string {
	pa := seg.PosTrimArr(str, search, pos...)
	return seg.CutStr(pa)
}

// CutTrimHtml cut string trim html and symbol return []string
func (seg *Segmenter) CutTrimHtml(str string, hmm ...bool) []string {
	str = FilterHtml(str)
	s := seg.Cut(str, hmm...)
	return seg.TrimSymbol(s)
}

// CutTrimHtmls cut string trim html and symbol return string
func (seg *Segmenter) CutTrimHtmls(str string, hmm ...bool) string {
	s := seg.CutTrimHtml(str, hmm...)
	return seg.CutStr(s, " ")
}

// CutUrl cut url string trim symbol return []string
func (seg *Segmenter) CutUrl(str string, num ...bool) []string {
	if len(num) <= 0 {
		// seg.Num = true
		str = SplitNums(str)
	}
	s := seg.Cut(str)
	return seg.TrimSymbol(s)
}

// CutUrls cut url string trim symbol return string
func (seg *Segmenter) CutUrls(str string, num ...bool) string {
	return seg.CutStr(seg.CutUrl(str, num...), " ")
}

// SplitNum cut string by num to []string
func SplitNum(text string) []string {
	r := regexp.MustCompile(`\d+|\D+`)
	return r.FindAllString(text, -1)
}

// SplitNums cut string by num to string
func SplitNums(text string) string {
	return strings.Join(SplitNum(text), " ")
}

// FilterEmoji filter the emoji
func FilterEmoji(text string) (new string) {
	for _, value := range text {
		_, size := utf8.DecodeRuneInString(string(value))
		if size <= 3 {
			new += string(value)
		}
	}

	return
}

// FilterSymbol filter the symbol
func FilterSymbol(text string) (new string) {
	for _, value := range text {
		if !unicode.IsSymbol(value) &&
			!unicode.IsSpace(value) && !unicode.IsPunct(value) {
			new += string(value)
		}
	}

	return
}

// FilterHtml filter the html tag
func FilterHtml(text string) string {
	regHtml := regexp.MustCompile(`(?U)\<[^>]*[\w|=|"]+\>`)
	text = regHtml.ReplaceAllString(text, "")
	return text
}

// FilterLang filter the language
func FilterLang(text, lang string) (new string) {
	for _, value := range text {
		if unicode.IsLetter(value) || unicode.Is(unicode.Scripts[lang], value) {
			new += string(value)
		}
	}

	return
}

// Range range text to []string
func Range(text string) (new []string) {
	for _, value := range text {
		new = append(new, string(value))
	}

	return
}

// RangeText range text to string
func RangeText(text string) (new string) {
	for _, value := range text {
		new += string(value) + " "
	}

	return
}
