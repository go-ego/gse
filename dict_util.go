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
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

var (
	// ToLower set alpha tolower
	ToLower = true
)

const (
	zhS1 = "dict/zh/s_1.txt"
	zhT1 = "dict/zh/t_1.txt"
)

// Init init seg config
func (seg *Segmenter) Init() {
	if seg.MinTokenFreq == 0 {
		seg.MinTokenFreq = 2.0
	}

	if seg.TextFreq == "" {
		seg.TextFreq = "2.0"
	}
}

// Dictionary 返回分词器使用的词典
func (seg *Segmenter) Dictionary() *Dictionary {
	return seg.Dict
}

// ToToken make the text, freq and pos to token structure
func (seg *Segmenter) ToToken(text string, freq float64, pos ...string) Token {
	var po string
	if len(pos) > 0 {
		po = pos[0]
	}

	words := seg.SplitTextToWords([]byte(text))
	token := Token{text: words, freq: freq, pos: po}
	return token
}

// AddToken add new text to token
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
	words := seg.SplitTextToWords([]byte(text))
	token := Token{text: words}

	return seg.Dict.RemoveToken(token)
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

		words := seg.SplitTextToWords([]byte(d["text"]))
		token := Token{text: words, freq: freq, pos: d["pos"]}
		seg.Dict.AddToken(token)
	}

	seg.CalcToken()
	return nil
}

// LoadDict load the dictionary from the file
//
// The format of the dictionary is (one for each participle):
//	participle text, frequency, part of speech
//
// Can load multiple dictionary files, the file name separated by "," or ", "
// the front of the dictionary preferentially load the participle,
//	such as: "user_dictionary.txt,common_dictionary.txt"
//
// When a participle appears both in the user dictionary and
// in the `common dictionary`, the `user dictionary` is given priority.
//
// 从文件中载入词典
//
// 可以载入多个词典文件，文件名用 "," 或 ", " 分隔，排在前面的词典优先载入分词，比如:
// 	"用户词典.txt,通用词典.txt"
//
// 当一个分词既出现在用户词典也出现在 `通用词典` 中，则优先使用 `用户词典`。
//
// 词典的格式为（每个分词一行）：
//	分词文本 频率 词性
func (seg *Segmenter) LoadDict(files ...string) error {
	if !seg.Load {
		seg.Dict = NewDict()
		seg.Load = true
		seg.Init()
	}

	var (
		dictDir  = path.Join(path.Dir(GetCurrentFilePath()), "data")
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
		dictPath = path.Join(dictDir, zhS1)
		path1 := path.Join(dictDir, zhT1)
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

// GetCurrentFilePath get the current file path
func GetCurrentFilePath() string {
	_, filePath, _, _ := runtime.Caller(1)
	return filePath
}

// GetIdfPath get the idf path
func GetIdfPath(files ...string) []string {
	var (
		dictDir  = path.Join(path.Dir(GetCurrentFilePath()), "data")
		dictPath = path.Join(dictDir, "dict/zh/idf.txt")
	)

	files = append(files, dictPath)

	return files
}

// Read read the dict flie
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

// Size frequency is calculated based on the size of the text
func (seg *Segmenter) Size(size int, text, freqText string) (freq float64) {
	if size == 0 {
		// 文件结束或错误行
		// continue
		return
	}

	if size < 2 {
		if !seg.LoadNoFreq {
			// 无效行
			return
		}

		freqText = seg.TextFreq
	}

	// 解析词频
	var err error
	freq, err = strconv.ParseFloat(freqText, 64)
	if err != nil {
		// continue
		return
	}

	// 过滤频率太小的词
	if freq < seg.MinTokenFreq {
		return 0.0
	}

	// 过滤长度为1的词, 降低词频
	if len([]rune(text)) < 2 {
		freq = 2
	}

	return
}

// Reader load dictionary from io.Reader
func (seg *Segmenter) Reader(reader io.Reader, files ...string) error {
	var (
		file           string
		text, freqText string
		freq           float64
		pos            string
	)

	if len(files) > 0 {
		file = files[0]
	}

	// 逐行读入分词
	line := 0
	for {
		line++
		size, fsErr := fmt.Fscanln(reader, &text, &freqText, &pos)
		if fsErr != nil {
			if fsErr == io.EOF {
				// End of file
				break
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
			// 没有词性, 标注为空字符串
			pos = ""
		}

		// 将分词添加到字典中
		words := seg.SplitTextToWords([]byte(text))
		token := Token{text: words, freq: freq, pos: pos}
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
			dictPath = path.Join(dictDir, "dict/jp/dict.txt")
		}

		if fileName[i] == "zh" {
			dictPath = path.Join(dictDir, zhS1)
			path1 := path.Join(dictDir, zhT1)
			files = append(files, path1)
		}

		if fileName[i] == "zh_s" {
			dictPath = path.Join(dictDir, zhS1)
		}

		if fileName[i] == "zh_t" {
			dictPath = path.Join(dictDir, zhT1)
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

// CalcToken calc the segmenter token
func (seg *Segmenter) CalcToken() {
	// 计算每个分词的路径值，路径值含义见 Token 结构体的注释
	logTotalFreq := float32(math.Log2(seg.Dict.totalFreq))
	for i := range seg.Dict.Tokens {
		token := &seg.Dict.Tokens[i]
		token.distance = logTotalFreq - float32(math.Log2(token.freq))
	}

	// 对每个分词进行细致划分，用于搜索引擎模式，
	// 该模式用法见 Token 结构体的注释。
	for i := range seg.Dict.Tokens {
		token := &seg.Dict.Tokens[i]
		segments := seg.segmentWords(token.text, true)

		// 计算需要添加的子分词数目
		numTokensToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if len(segments[iToken].token.text) > 0 {
				hasJp := false
				if len(segments[iToken].token.text) == 1 {
					segText := string(segments[iToken].token.text[0])
					hasJp = IsJp(segText)
				}

				if !hasJp {
					numTokensToAdd++
				}
			}
		}
		token.segments = make([]*Segment, numTokensToAdd)

		// 添加子分词
		iSegmentsToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			if len(segments[iToken].token.text) > 0 {
				hasJp := false
				if len(segments[iToken].token.text) == 1 {
					segText := string(segments[iToken].token.text[0])
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
