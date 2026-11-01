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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-ego/gse/types"
)

var (
	// ToLower set alpha to lowercase
	ToLower = true
)

const (
	zhS1 = "dict/zh/s_1.txt"
	zhT1 = "dict/zh/t_1.txt"
)

// Init initializes the segmenter config
func (seg *Segmenter) Init() {
	if seg.MinTokenFreq == 0 {
		seg.MinTokenFreq = 2.0
	}

	if seg.TextFreq == "" {
		seg.TextFreq = "2.0"
	}

	// init the model of hmm cut
	if !seg.NotLoadHMM {
		seg.LoadModel()
	}
}

// Dictionary returns the dictionary used by the tokenizer
func (seg *Segmenter) Dictionary() *Dictionary {
	return seg.Dict
}

// ToToken make the text, freq and pos to token structure
func (seg *Segmenter) ToToken(text string, freq float64, pos ...string) Token {
	var po string
	if len(pos) > 0 {
		po = pos[0]
	}

	b := []byte(text)
	var buf [16]Text
	words := seg.appendWords(buf[:0], b)
	// the words are b split, only the letters may be lowered
	if bytes.ContainsFunc(b, isUpper) {
		b = textSliceToBytes(words)
	}

	return Token{text: b, size: int32(len(words)), freq: freq, pos: po}
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// AddToken add a new text to the token
func (seg *Segmenter) AddToken(text string, freq float64, pos ...string) error {
	token := seg.ToToken(text, freq, pos...)
	return seg.Dict.AddToken(token)
}

// AddTokenForce add new text to token and force
// time-consuming
func (seg *Segmenter) AddTokenForce(text string, freq float64, pos ...string) (err error) {
	err = seg.AddToken(text, freq, pos...)
	seg.CalcToken()
	return
}

// ReAddToken remove and add token again
func (seg *Segmenter) ReAddToken(text string, freq float64, pos ...string) error {
	token := seg.ToToken(text, freq, pos...)
	err := seg.Dict.RemoveToken(token)
	if err != nil {
		return err
	}
	return seg.Dict.AddToken(token)
}

// RemoveToken remove token in dictionary
func (seg *Segmenter) RemoveToken(text string) error {
	return seg.Dict.RemoveToken(seg.ToToken(text, 0))
}

// Empty empty the seg dictionary
func (seg *Segmenter) Empty() error {
	seg.Dict = nil
	return nil
}

// LoadDictMap load dictionary from []map[string]string
func (seg *Segmenter) LoadDictMap(dict []map[string]string) error {
	if seg.Dict == nil {
		seg.Dict = NewDict()
		seg.Init()
	}

	for _, d := range dict {
		// Parse the word frequency
		freq := seg.Size(len(d), d["text"], d["freq"])
		if freq == 0.0 {
			continue
		}

		seg.Dict.AddToken(seg.ToToken(d["text"], freq, d["pos"]))
	}

	seg.CalcToken()
	return nil
}

// LoadDict load the dictionary from the file
//
// The format of the dictionary is (one for each participle):
//
//	participle text, frequency, part of speech
//
// # And you can option the dictionary separator by seg.DictSep = ","
//
// Can load multiple dictionary files, the file name separated by "," or ", "
// the front of the dictionary preferentially load the participle,
//
//	such as: "user_dictionary.txt,common_dictionary.txt"
//
// When a participle appears both in the user dictionary and
// in the `common dictionary`, the `user dictionary` is given priority.
func (seg *Segmenter) LoadDict(files ...string) error {
	if seg.Dict == nil {
		seg.Dict = NewDict()
		// seg.Load = true
		seg.Init()
	}

	var (
		dictDir  = filepath.Join(filepath.Dir(seg.GetCurrentFilePath()), "data")
		dictPath string
		// load     bool
	)

	if len(files) > 0 {
		dictFiles := DictPaths(dictDir, files[0])
		if !seg.SkipLog {
			log.Println("Dict files path: ", dictFiles)
		}

		if len(dictFiles) == 0 {
			log.Println("Warning: dict files is nil.")
			// return errors.New("Dict files is nil.")
		}

		if len(dictFiles) > 0 {
			// load = true
			// files = dictFiles
			for i := 0; i < len(dictFiles); i++ {
				err := seg.Read(dictFiles[i])
				if err != nil {
					return err
				}
			}
		}
	}

	if len(files) == 0 {
		dictPath = filepath.Join(dictDir, zhS1)
		path1 := filepath.Join(dictDir, zhT1)
		// files = []string{dictPath}
		err := seg.Read(dictPath)
		if err != nil {
			return err
		}

		err = seg.Read(path1)
		if err != nil {
			return err
		}
	}

	// if files[0] != "" && files[0] != "en" && !load {
	// 	for _, file := range strings.Split(files[0], ",") {
	// 		// for _, file := range files {
	// 		err := seg.Read(file)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	seg.CalcToken()
	if !seg.SkipLog {
		log.Println("Gse dictionary loaded finished.")
	}

	return nil
}

// LoadTFIDFDict load tfidf dict for cal tfidf & bm25
func (seg *Segmenter) LoadTFIDFDict(files []*types.LoadDictFile) error {
	if seg.Dict == nil {
		seg.Dict = NewDict()
		// seg.Load = true
		seg.Init()
	}

	var (
		dictDir = filepath.Join(filepath.Dir(seg.GetCurrentFilePath()), "data")
	)

	for _, file := range files {
		dictFiles := DictPaths(dictDir, file.FilePath)
		if !seg.SkipLog {
			log.Println("Dict files path: ", dictFiles)
		}

		if len(dictFiles) == 0 {
			log.Println("Warning: dict files is nil.")
			// return errors.New("Dict files is nil.")
		}

		if len(dictFiles) > 0 {
			for i := 0; i < len(dictFiles); i++ {
				err := seg.ReadTFIDF(dictFiles[i])
				if err != nil {
					return err
				}
			}
		}
	}

	seg.CalcToken()
	if !seg.SkipLog {
		log.Println("Gse dictionary loaded finished.")
	}
	return nil
}

// GetCurrentFilePath get the current file path
func (seg *Segmenter) GetCurrentFilePath() string {
	if seg.DictPath != "" {
		return seg.DictPath
	}

	_, filePath, _, _ := runtime.Caller(1)
	return filePath
}

// GetIdfPath get the idf path
func (seg *Segmenter) GetIdfPath(files ...string) []string {
	var (
		dictDir  = filepath.Join(filepath.Dir(seg.GetCurrentFilePath()), "data")
		dictPath = filepath.Join(dictDir, "dict/zh/idf.txt")
	)

	files = append(files, dictPath)
	return files
}

// LoadCorpusAverLen load the average length of corpus
func (seg *Segmenter) LoadCorpusAverLen(files ...string) (corpusTotal float64, err error) {
	filePaths := seg.GetCorpusPath(files...)
	corpusTotal = 0
	for _, v := range filePaths {
		var number float64
		number, err = seg.ReadCorpus(v)
		if err != nil {
			log.Printf("Could not read corpus from file path: \"%s\", %v \n", v, err)
			return
		}
		corpusTotal += number
	}

	corpusTotal = corpusTotal / float64(len(filePaths))
	return
}

// GetCorpusPath get the corpus path
func (seg *Segmenter) GetCorpusPath(files ...string) []string {
	var (
		dictDir  = filepath.Join(filepath.Dir(seg.GetCurrentFilePath()), "data")
		dictPath = filepath.Join(dictDir, "dict/zh/tf_idf_origin.txt")
	)

	files = append(files, dictPath)
	return files
}

func (seg *Segmenter) ReadCorpus(file string) (corpusAverLen float64, err error) {
	if !seg.SkipLog {
		log.Printf("Load the gse dictionary: \"%s\" ", file)
	}
	var corpusNumber float64 = 0
	var corpusLength float64 = 0
	dictFile, err := os.Open(file)
	if err != nil {
		log.Printf("Could not load dictionaries: \"%s\", %v \n", file, err)
		return
	}
	defer dictFile.Close()

	// new the Scanner to read file content
	scanner := bufio.NewScanner(dictFile)
	// read file content by line
	for scanner.Scan() {
		corpusNumber++
		line := scanner.Text()
		corpusLength += float64(utf8.RuneCountInString(line))
	}
	corpusAverLen = corpusLength / corpusNumber
	return
}

// GetTfIdfPath get the tfidf path
func (seg *Segmenter) GetTfIdfPath(files ...string) []string {
	var (
		dictDir  = filepath.Join(filepath.Dir(seg.GetCurrentFilePath()), "data")
		dictPath = filepath.Join(dictDir, "dict/zh/tf_idf.txt")
	)

	files = append(files, dictPath)
	return files
}

// Read read the dict file
func (seg *Segmenter) Read(file string) error {
	if !seg.SkipLog {
		log.Printf("Load the gse dictionary: \"%s\" ", file)
	}

	dictFile, err := os.Open(file)
	if err != nil {
		log.Printf("Could not load dictionaries: \"%s\", %v \n", file, err)
		return err
	}
	defer dictFile.Close()

	reader := bufio.NewReader(dictFile)
	return seg.Reader(reader, file)
}

// ReadTFIDF read the dict file
func (seg *Segmenter) ReadTFIDF(file string) error {
	if !seg.SkipLog {
		log.Printf("Load the gse dictionary: \"%s\" ", file)
	}

	dictFile, err := os.Open(file)
	if err != nil {
		log.Printf("Could not load dictionaries: \"%s\", %v \n", file, err)
		return err
	}
	defer dictFile.Close()

	reader := bufio.NewReader(dictFile)
	return seg.ReaderTFIDF(reader, file)
}

// Size frequency is calculated based on the size of the text
func (seg *Segmenter) Size(size int, text, freqText string) (freq float64) {
	if size == 0 {
		// End of file or error line
		// continue
		return
	}

	if size < 2 {
		if !seg.LoadNoFreq {
			// invalid row line
			return
		}

		freqText = seg.TextFreq
	}

	// Analyze the word frequency
	var err error
	freq, err = strconv.ParseFloat(freqText, 64)
	if err != nil {
		// continue
		return
	}

	// Filter out the words that are too infrequent
	if freq < seg.MinTokenFreq {
		return 0.0
	}

	// Filter words with a length of 1 to reduce the word frequency
	if len([]rune(text)) < 2 {
		freq = 2
	}

	return
}

// ReadN read the tokens by '\n'
func (seg *Segmenter) ReadN(reader *bufio.Reader) (size int, text, freqText, pos string, fsErr error) {
	var txt string
	txt, fsErr = reader.ReadString('\n')

	parts := strings.Split(txt, seg.DictSep+" ")
	size = len(parts)

	text = parts[0]
	if size > 1 {
		freqText = strings.TrimSpace(parts[1])
	}
	if size > 2 {
		pos = strings.TrimSpace(strings.Trim(parts[2], "\n"))
	}

	return
}

// ReadNTFIDF read the tokens with tfidf by '\n'
func (seg *Segmenter) ReadNTFIDF(reader *bufio.Reader) (size int, text, freqText, idfText string, fsErr error) {
	var txt string
	txt, fsErr = reader.ReadString('\n')

	parts := strings.Split(txt, seg.DictSep+" ")
	size = len(parts)

	text = parts[0]
	if size > 2 {
		freqText = strings.TrimSpace(parts[1])
		idfText = strings.TrimSpace(strings.Trim(parts[2], "\n"))
	}

	return
}

// Reader load dictionary from io.Reader
func (seg *Segmenter) Reader(reader *bufio.Reader, files ...string) error {
	var (
		file           string
		text, freqText string
		freq           float64
		pos            string
	)

	if len(files) > 0 {
		file = files[0]
	}

	// Read the word segmentation line by line
	line := 0
	for {
		line++
		var (
			size  int
			fsErr error
		)
		if seg.DictSep == "" {
			size, fsErr = fmt.Fscanln(reader, &text, &freqText, &pos)
		} else {
			size, text, freqText, pos, fsErr = seg.ReadN(reader)
		}

		if fsErr != nil {
			if fsErr == io.EOF {
				// End of file
				if seg.DictSep == "" {
					break
				}

				if seg.DictSep != "" && text == "" {
					break
				}
			}

			if size > 0 {
				if seg.MoreLog {
					log.Printf("File '%v' line \"%v\" read error: %v, skip",
						file, line, fsErr.Error())
				}
			} else {
				log.Printf("File '%v' line \"%v\" is empty, read error: %v, skip",
					file, line, fsErr.Error())
			}
		}

		freq = seg.Size(size, text, freqText)
		if freq == 0.0 {
			continue
		}

		if size == 2 {
			// No part of speech, marked as an empty string
			pos = ""
		}

		// Add participle tokens to the dictionary
		seg.Dict.AddToken(seg.ToToken(text, freq, pos))
	}

	return nil
}

// ReaderTFIDF load tfidf dictionary from io.Reader
func (seg *Segmenter) ReaderTFIDF(reader *bufio.Reader, files ...string) error {
	var (
		file                    string
		text, freqText, idfText string
		freq                    float64
		inverseFreq             float64
	)

	if len(files) > 0 {
		file = files[0]
	}

	// Read the word segmentation line by line
	line := 0
	for {
		line++
		var (
			size  int
			fsErr error
		)
		if seg.DictSep == "" {
			size, fsErr = fmt.Fscanln(reader, &text, &freqText, &idfText)
		} else {
			size, text, freqText, idfText, fsErr = seg.ReadNTFIDF(reader)
		}

		if fsErr != nil {
			if fsErr == io.EOF {
				// End of file
				if seg.DictSep == "" {
					break
				}

				if seg.DictSep != "" && text == "" {
					break
				}
			}

			if size > 0 {
				if seg.MoreLog {
					log.Printf("File '%v' line \"%v\" read error: %v, skip",
						file, line, fsErr.Error())
				}
			} else {
				log.Printf("File '%v' line \"%v\" is empty, read error: %v, skip",
					file, line, fsErr.Error())
			}
		}

		freq = seg.Size(size, text, freqText)
		inverseFreq = seg.Size(size, text, idfText)
		if freq == 0.0 || inverseFreq == 0.0 {
			continue
		}

		// Add participle tokens to the dictionary
		token := seg.ToToken(text, freq)
		token.inverseFreq = inverseFreq
		seg.Dict.AddToken(token)
	}

	return nil
}

// DictPaths get the dict's paths
func DictPaths(dictDir, filePath string) (files []string) {
	var dictPath string

	if filePath == "en" {
		return
	}

	var fileName []string
	if strings.Contains(filePath, ", ") {
		fileName = strings.Split(filePath, ", ")
	} else {
		fileName = strings.Split(filePath, ",")
	}

	for i := 0; i < len(fileName); i++ {
		if fileName[i] == "ja" || fileName[i] == "jp" {
			dictPath = filepath.Join(dictDir, "dict/jp/dict.txt")
		}

		if fileName[i] == "zh" {
			dictPath = filepath.Join(dictDir, zhS1)
			path1 := filepath.Join(dictDir, zhT1)
			files = append(files, path1)
		}

		if fileName[i] == "zh_s" {
			dictPath = filepath.Join(dictDir, zhS1)
		}

		if fileName[i] == "zh_t" {
			dictPath = filepath.Join(dictDir, zhT1)
		}

		// if str[i] == "ti" {
		// }

		dictName := fileName[i] != "en" &&
			fileName[i] != "zh" &&
			fileName[i] != "zh_s" && fileName[i] != "zh_t" &&
			fileName[i] != "ja" && fileName[i] != "jp" &&
			fileName[i] != "ko" && fileName[i] != "ti"

		if dictName {
			dictPath = fileName[i]
		}

		if dictPath != "" && dictPath != " " {
			files = append(files, dictPath)
		}
	}
	// }

	return
}

// IsJp is Japan char return true
func IsJp(segText string) bool {
	for _, r := range segText {
		jp := unicode.Is(unicode.Scripts["Hiragana"], r) ||
			unicode.Is(unicode.Scripts["Katakana"], r)
		if jp {
			return true
		}
	}
	return false
}

// firstIsJp reports whether the first Text unit of text is Japanese,
// a kana is never alphanumeric, so it is always a one rune unit
func firstIsJp(text Text) bool {
	r, _ := utf8.DecodeRune(text)
	return unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r)
}

// CalcToken calc the segmenter token
func (seg *Segmenter) CalcToken() {
	// Calculate the path value of each word segment.
	// For the meaning of the path value, see the notes of the Token structure
	logTotalFreq := float32(math.Log2(seg.Dict.totalFreq))
	for i := range seg.Dict.Tokens {
		token := &seg.Dict.Tokens[i]
		token.distance = logTotalFreq - float32(math.Log2(token.freq))
	}

	// Each word segmentation is carefully divided for search engine mode,
	// For the usage of this mode, see the comments of the Token structure.
	if seg.SkipSubSeg {
		return
	}

	var words []Text
	for i := range seg.Dict.Tokens {
		token := &seg.Dict.Tokens[i]
		words = seg.appendWords(words[:0], token.text)
		segments := seg.segmentWords(words, true)

		// Calculate the number of sub-segments that need to be added
		numTokensToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if segments[iToken].token.size > 0 {
				hasJp := false
				if segments[iToken].token.size == 1 {
					segText := string(segments[iToken].token.text)
					hasJp = IsJp(segText)
				}

				if !hasJp {
					numTokensToAdd++
				}
			}
		}
		token.segments = make([]*Segment, numTokensToAdd)

		// add sub-segments subparticiple
		iSegmentsToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if segments[iToken].token.size > 0 {
				hasJp := false
				if segments[iToken].token.size == 1 {
					segText := string(segments[iToken].token.text)
					hasJp = IsJp(segText)
				}

				if !hasJp {
					token.segments[iSegmentsToAdd] = &segments[iToken]
					iSegmentsToAdd++
				}
			}
		}
	}
}
