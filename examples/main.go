package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/hmm/extracker"
	"github.com/go-ego/gse/hmm/pos"
)

var (
	seg    gse.Segmenter
	posSeg pos.Segmenter

	text  = "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄制作的的科幻片."
	text1 = flag.String("text", text, "要分词的文本")

	text2 = "西雅图地标建筑, Seattle Space Needle, 西雅图太空针. Sky tree."
)

func main() {
	flag.Parse()

	// Loading the default dictionary
	seg.LoadDict()
	// Loading the default dictionary with embed
	// seg.LoadDictEmbed()
	//
	// Loading the simple chinese dictionary
	// seg.LoadDict("zh_s")
	// seg.LoadDictEmbed("zh_s")
	//
	// Loading the traditional chinese dictionary
	// seg.LoadDict("zh_t")
	//
	// Loading the japanese dictionary
	// seg.LoadDict("jp")
	//
	// seg.LoadDict("../data/dict/dictionary.txt")
	//
	// Loading the custom dictionary
	// seg.LoadDict("zh,../../testdata/zh/test_dict.txt,../../testdata/zh/test_dict1.txt")

	addToken()

	cut()
	//
	cutPos()
	segCut()

	extAndRank(seg)
}

func addToken() {
	err := seg.AddToken("《复仇者联盟3：无限战争》", 100, "n")
	fmt.Println("add token: ", err)
	seg.AddToken("西雅图中心", 100)
	seg.AddToken("西雅图太空针", 100, "n")
	seg.AddToken("Space Needle", 100, "n")
	// seg.AddTokenForce("上海东方明珠广播电视塔", 100, "n")
	//
	seg.AddToken("太空针", 100)
	seg.ReAddToken("太空针", 100, "n")
	freq, pos, ok := seg.Find("太空针")
	fmt.Println("seg.Find: ", freq, pos, ok)

	// seg.CalcToken()
	err = seg.RemoveToken("太空针")
	fmt.Println("remove token: ", err)
}

// 使用 DAG 或 HMM 模式分词
func cut() {
	// "《复仇者联盟3：无限战争》是全片使用IMAX摄影机拍摄制作的的科幻片."

	// use DAG and HMM
	hmm := seg.Cut(text, true)
	fmt.Println("cut use hmm: ", hmm)
	// cut use hmm:  [《复仇者联盟3：无限战争》 是 全片 使用 imax 摄影机 拍摄 制作 的 的 科幻片 .]

	cut := seg.Cut(text)
	fmt.Println("cut: ", cut)
	// cut:  [《 复仇者 联盟 3 ： 无限 战争 》 是 全片 使用 imax 摄影机 拍摄 制作 的 的 科幻片 .]

	hmm = seg.CutSearch(text, true)
	fmt.Println("cut search use hmm: ", hmm)
	// cut search use hmm:  [复仇 仇者 联盟 无限 战争 复仇者 《复仇者联盟3：无限战争》 是 全片 使用 imax 摄影 摄影机 拍摄 制作 的 的 科幻 科幻片 .]
	fmt.Println("analyze: ", seg.Analyze(hmm, text))

	cut = seg.CutSearch(text)
	fmt.Println("cut search: ", cut)
	// cut search:  [《 复仇 者 复仇者 联盟 3 ： 无限 战争 》 是 全片 使用 imax 摄影 机 摄影机 拍摄 制作 的 的 科幻 片 科幻片 .]

	cut = seg.CutAll(text)
	fmt.Println("cut all: ", cut)
	// cut all:  [《复仇者联盟3：无限战争》 复仇 复仇者 仇者 联盟 3 ： 无限 战争 》 是 全片 使用 i m a x 摄影 摄影机 拍摄 摄制 制作 的 的 科幻 科幻片 .]

	s := seg.CutStr(cut, ", ")
	fmt.Println("cut all to string: ", s)
	// cut all to string:  《复仇者联盟3：无限战争》, 复仇, 复仇者, 仇者, 联盟, 3, ：, 无限, 战争, 》, 是, 全片, 使用, i, m, a, x, 摄影, 摄影机, 拍摄, 摄制, 制作, 的, 的, 科幻, 科幻片, .

	analyzeAndTrim(cut)

	reg := regexp.MustCompile(`(\d+年|\d+月|\d+日|[\p{Latin}]+|[\p{Hangul}]+|\d+\.\d+|[a-zA-Z0-9]+)`)
	text1 := `헬로월드 헬로 서울, 2021年09月10日, 3.14`
	hmm = seg.CutDAG(text1, reg)
	fmt.Println("Cut with hmm and regexp: ", hmm, hmm[0], hmm[6])
}

func analyzeAndTrim(cut []string) {
	a := seg.Analyze(cut, "")
	fmt.Println("analyze the segment: ", a)
	// analyze the segment:

	cut = seg.Trim(cut)
	fmt.Println("cut all: ", cut)
	// cut all:  [复仇者联盟3无限战争 复仇 复仇者 仇者 联盟 3 无限 战争 是 全片 使用 i m a x 摄影 摄影机 拍摄 摄制 制作 的 的 科幻 科幻片]

	fmt.Println(seg.String(text2, true))
	// 西雅图/nr 地标/n 建筑/n ,/x  /x seattle/x  /x space needle/n ,/x  /x 西雅图太空针/n ./x  /x sky/x  /x tree/x ./x
	fmt.Println(seg.Slice(text2, true))
	// [西雅图 地标 建筑 ,   seattle   space needle ,   西雅图太空针 .   sky   tree .]
}

func cutPos() {
	// "西雅图地标建筑, Seattle Space Needle, 西雅图太空针. Sky tree."

	po := seg.Pos(text2, true)
	fmt.Println("pos: ", po)
	// pos:  [{西雅图 nr} {地标 n} {建筑 n} {, x} {  x} {seattle x} {  x} {space needle n} {, x} {  x} {西雅图太空针 n} {. x} {  x} {sky x} {  x} {tree x} {. x}]

	po = seg.TrimWithPos(po, "zg")
	fmt.Println("trim pos: ", po)
	// trim pos:  [{西雅图 nr} {地标 n} {建筑 n} {, x} {  x} {seattle x} {  x} {space needle n} {, x} {  x} {西雅图太空针 n} {. x} {  x} {sky x} {  x} {tree x} {. x}]

	posSeg.WithGse(seg)
	po = posSeg.Cut(text, true)
	fmt.Println("pos: ", po)
	// pos:  [{《 x} {复仇 v} {者 k} {联盟 j} {3 x} {： x} {无限 v} {战争 n} {》 x} {是 v} {全片 n} {使用 v} {imax eng} {摄影 n} {机 n} {拍摄 v} {制作 vn} {的的 u} {科幻 n} {片 q} {. m}]

	po = posSeg.TrimWithPos(po, "zg")
	fmt.Println("trim pos: ", po)
	// trim pos:  [{《 x} {复仇 v} {者 k} {联盟 j} {3 x} {： x} {无限 v} {战争 n} {》 x} {是 v} {全片 n} {使用 v} {imax eng} {摄影 n} {机 n} {拍摄 v} {制作 vn} {的的 u} {科幻 n} {片 q} {. m}]
}

// 使用最短路径和动态规划分词
func segCut() {
	segments := seg.Segment([]byte(*text1))
	fmt.Println(gse.ToString(segments, true))
	// 《/x 复仇/v 者/k 复仇者/n 联盟/j 3/x ：/x 无限/v 战争/n 》/x 是/v 全片/n 使用/v imax/x 摄影/n 机/n 摄影机/n 拍摄/v 制作/vn 的/uj 的/uj 科幻/n 片/q 科幻片/n ./x

	segs := seg.Segment([]byte(text2))
	// log.Println(gse.ToString(segs, false))
	log.Println(gse.ToString(segs))
	// 西雅图/nr 地标/n 建筑/n ,/x  /x seattle/x  /x space needle/n ,/x  /x 西雅图太空针/n ./x  /x sky/x  /x tree/x ./x

	// 搜索模式主要用于给搜索引擎提供尽可能多的关键字
	// segs := seg.ModeSegment(text2, true)
	log.Println("搜索模式: ", gse.ToString(segs, true))
	// 搜索模式:  西雅图/nr 地标/n 建筑/n ,/x  /x seattle/x  /x space needle/n ,/x  /x 西雅图太空针/n ./x  /x sky/x  /x tree/x ./x

	log.Println("to slice", gse.ToSlice(segs, true))
	// to slice [西雅图 地标 建筑 ,   seattle   space needle ,   西雅图太空针 .   sky   tree .]
}

func extAndRank(segs gse.Segmenter) {
	var te extracker.TagExtracter
	te.WithGse(segs)
	err := te.LoadIdf()
	fmt.Println("load idf: ", err)

	segments := te.ExtractTags(text, 5)
	fmt.Println("segments: ", len(segments), segments)
	// segments:  5 [{科幻片 1.6002581704125} {全片 1.449761569875} {摄影机 1.2764747747375} {拍摄 0.9690261695075} {制作 0.8246043033375}]

	var tr extracker.TextRanker
	tr.WithGse(segs)

	results := tr.TextRank(text, 5)
	fmt.Println("results: ", results)
	// results:  [{机 1} {全片 0.9931964427972227} {摄影 0.984870660504368} {使用 0.9769826633059524} {是 0.8489363954683677}]
}
