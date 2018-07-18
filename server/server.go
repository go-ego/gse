/*

gse 分词服务器同时提供了两种模式：

	"/"	分词演示网页
	"/json"	JSON 格式的 RPC 服务
		输入：
			POST 或 GET 模式输入 text 参数
		输出 JSON 格式：
			{
				segments:[
					{"text":"服务器", "pos":"n"},
					{"text":"指令", "pos":"n"},
					...
				]
			}


测试服务器见 http://gse.weiboglass.com

*/

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"runtime"

	"encoding/json"
	"net/http"

	"github.com/go-ego/gse"
)

var (
	segmenter = gse.Segmenter{}

	host         = flag.String("host", "", "HTTP服务器主机名")
	port         = flag.Int("port", 8080, "HTTP服务器端口")
	dict         = flag.String("dict", "../data/dict/dictionary.txt", "词典文件")
	staticFolder = flag.String("static_folder", "static", "静态页面存放的目录")
)

// JsonResponse json response
type JsonResponse struct {
	Segments []*Segment `json:"segments"`
}

// Segment segment json struct
type Segment struct {
	Text string `json:"text"`
	Pos  string `json:"pos"`
}

// JsonRpcServer json rpc server
func JsonRpcServer(w http.ResponseWriter, req *http.Request) {
	// 得到要分词的文本
	text := req.URL.Query().Get("text")
	if text == "" {
		text = req.PostFormValue("text")
	}

	// 分词
	segments := segmenter.Segment([]byte(text))

	// 整理为输出格式
	ss := []*Segment{}
	for _, segment := range segments {
		ss = append(ss, &Segment{
			Text: segment.Token().Text(), Pos: segment.Token().Pos()})
	}
	response, _ := json.Marshal(&JsonResponse{Segments: ss})

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(response))
}

func main() {
	flag.Parse()

	// 将线程数设置为 CPU数
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 初始化分词器
	segmenter.LoadDict(*dict)

	http.HandleFunc("/json", JsonRpcServer)
	http.Handle("/", http.FileServer(http.Dir(*staticFolder)))

	log.Print("服务器启动")
	http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port), nil)
}
