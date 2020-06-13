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

package pos

import (
	"math"

	"github.com/go-ego/gse"
)

// Dict represents a thread-safe dictionary used for word segmentation.
type Dict struct {
	total, logTotal float64
	seg             gse.Segmenter
}

// func New(files ...string) gse.Segmenter {
// 	return gse.New(files...)
// }

// AddToken adds one token
func (d *Dict) AddToken(text string, frequency int, pos ...string) {
	d.seg.AddToken(text, frequency, pos...)
}

func (d *Dict) updateLogTotal() {
	d.logTotal = math.Log(d.total)
}

// Frequency returns the frequency and existence of give word
func (d *Dict) Frequency(key string) (int, bool) {
	return d.seg.Find(key)
}

// Pos returns the POS and existence of give word
func (d *Dict) Pos(key string) (string, bool) {
	value, _, _ := d.seg.Value(key)
	if value == 0 {
		return "", false
	}

	pos := d.seg.Dict.Tokens[value].Pos()
	return pos, true
}

func (d *Dict) loadDict(files ...string) error {
	return d.seg.LoadDict(files...)
}
