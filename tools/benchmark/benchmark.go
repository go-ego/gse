/*

测试 gse 分词速度

go run benchmark.go

输出分词结果到文件：

go run benchmark.go -output=output.txt

分析性能瓶颈：

go build benchmark.go
./benchmark -cpuprofile=cpu.prof
go tool pprof benchmark cpu.prof

分析内存占用：

go build benchmark.go
./benchmark -memprofile=mem.prof
go tool pprof benchmark mem.prof

*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/go-ego/gse"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "处理器profile文件")
	memprofile = flag.String("memprofile", "", "内存profile文件")
	output     = flag.String("output", "", "输出分词结果到此文件")
	numRuns    = 20
)

func mem() {
	// 写入内存 profile 文件
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		defer f.Close()
	}
}

func openFile() ([][]byte, int) {
	// 打开将要分词的文件
	file, err := os.Open("../testdata/bailuyuan.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 逐行读入
	scanner := bufio.NewScanner(file)
	size := 0
	lines := [][]byte{}
	for scanner.Scan() {
		var text string
		fmt.Sscanf(scanner.Text(), "%s", &text)
		content := []byte(text)
		size += len(content)
		lines = append(lines, content)
	}

	return lines, size
}

func outputInit() *os.File {
	// 当指定输出文件时打开输出文件
	var (
		of  *os.File
		err error
	)

	if *output != "" {
		of, err = os.Create(*output)
		if err != nil {
			log.Fatal(err)
		}
		// defer of.Close()
	}

	return of
}

func cpu() {
	// 打开处理器 profile 文件
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
}

func main() {
	// 确保单线程，因为 Go 从 1.5 开始默认多线程
	runtime.GOMAXPROCS(1)

	// 解析命令行参数
	flag.Parse()

	// 记录时间
	t0 := time.Now()

	var segmenter gse.Segmenter
	segmenter.LoadDict("../data/dict/dictionary.txt")

	// 记录时间
	t1 := time.Now()
	log.Printf("载入词典花费时间 %v", t1.Sub(t0))

	mem()

	lines, size := openFile()
	of := outputInit()
	defer of.Close()

	// 记录时间
	t2 := time.Now()

	cpu()

	// 分词
	for i := 0; i < numRuns; i++ {
		for _, l := range lines {
			segments := segmenter.Segment(l)
			if *output != "" {
				of.WriteString(gse.ToString(segments))
				of.WriteString("\n")
			}
		}
	}

	// 停止处理器 profile
	if *cpuprofile != "" {
		defer pprof.StopCPUProfile()
	}

	// 记录时间并计算分词速度
	t3 := time.Now()
	log.Printf("分词花费时间 %v", t3.Sub(t2))

	ts := float64(size*numRuns) / t3.Sub(t2).Seconds() / (1024 * 1024)
	log.Printf("分词速度 %f MB/s", ts)
}
