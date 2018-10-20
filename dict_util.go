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

// getCurrentFilePath get current file path
func getCurrentFilePath() string {
	_, filePath, _, _ := runtime.Caller(1)
	return filePath
}

// Read read the dict flie
func (seg *Segmenter) Read(file string) error {
	log.Printf("Load the gse dictionary: \"%s\" ", file)
	dictFile, err := os.Open(file)
	if err != nil {
		log.Printf("Could not load dictionaries: \"%s\", %v \n", file, err)
		return err
	}
	defer dictFile.Close()

	reader := bufio.NewReader(dictFile)
	var (
		text      string
		freqText  string
		frequency int
		pos       string
	)

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
				log.Printf("File '%v' line \"%v\" read error: %v, skip",
					file, line, fsErr.Error())
			} else {
				log.Printf("File '%v' line \"%v\" is empty, read error: %v, skip",
					file, line, fsErr.Error())
			}
		}

		if size == 0 {
			// 文件结束或错误行
			// break
			continue
		} else if size < 2 {
			// 无效行
			continue
		} else if size == 2 {
			// 没有词性标注时设为空字符串
			pos = ""
		}

		// 解析词频
		var err error
		frequency, err = strconv.Atoi(freqText)
		if err != nil {
			continue
		}

		// 过滤频率太小的词
		if frequency < minTokenFrequency {
			continue
		}
		// 过滤, 降低词频
		if len([]rune(text)) < 2 {
			// continue
			frequency = 2
		}

		// 将分词添加到字典中
		words := splitTextToWords([]byte(text))
		token := Token{text: words, frequency: frequency, pos: pos}
		seg.dict.addToken(token)
	}

	return nil
}

// DictPaths get the dict's paths
func DictPaths(dictDir, filePath string) (files []string) {
	var dictPath string

	if filePath == "en" {
		return
	}

	if filePath == "zh" {
		dictPath = path.Join(dictDir, "dict/dictionary.txt")
		files = []string{dictPath}

		return
	}

	if filePath == "jp" {
		dictPath = path.Join(dictDir, "dict/jp/dict.txt")
		files = []string{dictPath}

		return
	}

	// if strings.Contains(filePath, ",") {
	fileName := strings.Split(filePath, ",")
	for i := 0; i < len(fileName); i++ {
		if fileName[i] == "jp" {
			dictPath = path.Join(dictDir, "dict/jp/dict.txt")
		}

		if fileName[i] == "zh" {
			dictPath = path.Join(dictDir, "dict/dictionary.txt")
		}

		// if str[i] == "ti" {
		// }

		dictName := fileName[i] != "en" && fileName[i] != "zh" &&
			fileName[i] != "jp" && fileName[i] != "ti"

		if dictName {
			dictPath = fileName[i]
		}

		if dictPath != "" {
			files = append(files, dictPath)
		}
	}
	// }
	log.Println("Dict files path: ", files)

	return
}

// IsJp is jp char return true
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

// SegToken add segmenter token
func (seg *Segmenter) SegToken() {
	// 计算每个分词的路径值，路径值含义见 Token 结构体的注释
	logTotalFrequency := float32(math.Log2(float64(seg.dict.totalFrequency)))
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		token.distance = logTotalFrequency - float32(math.Log2(float64(token.frequency)))
	}

	// 对每个分词进行细致划分，用于搜索引擎模式，
	// 该模式用法见 Token 结构体的注释。
	for i := range seg.dict.tokens {
		token := &seg.dict.tokens[i]
		segments := seg.segmentWords(token.text, true)

		// 计算需要添加的子分词数目
		numTokensToAdd := 0
		for iToken := 0; iToken < len(segments); iToken++ {
			// if len(segments[iToken].token.text) > 1 {
			// 略去字元长度为一的分词
			// TODO: 这值得进一步推敲，特别是当字典中有英文复合词的时候
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
			// if len(segments[iToken].token.text) > 1 {
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
