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

/*
The gse single thread segmentation speed, run from the repository root:

	go run ./examples/tools/benchmark
	go run ./examples/tools/benchmark -mode hmm -runs 10

Write the segments to a file:

	go run ./examples/tools/benchmark -output output.txt

Profile the cpu and memory:

	go build -o benchmark ./examples/tools/benchmark
	./benchmark -cpuprofile cpu.prof -memprofile mem.prof
	go tool pprof benchmark cpu.prof
	go tool pprof benchmark mem.prof
*/
package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/examples/tools/internal/segtool"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write the cpu profile to the file")
	memprofile = flag.String("memprofile", "", "write the memory profile to the file")
	output     = flag.String("output", "", "write the segments to the file")

	file = flag.String("file", "testdata/zh/bailuyuan_zh.txt", "text file to segment")
	dict = flag.String("dict", "", "dictionary files, the embedded zh dictionary by default")
	mode = flag.String("mode", segtool.Default, "segment mode: "+strings.Join(segtool.Names(), ", "))
	runs = flag.Int("runs", 20, "number of runs")
)

// segmentLines segments the lines runs times, writes the segments to w when not nil
func segmentLines(seg *gse.Segmenter, fn segtool.Func, lines []string, runs int, w io.Writer) error {
	for i := 0; i < runs; i++ {
		for _, l := range lines {
			segs := fn(seg, l)
			if w == nil {
				continue
			}

			if _, err := io.WriteString(w, segtool.Format(segs)+"\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeMemProfile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	runtime.GC()
	return pprof.WriteHeapProfile(f)
}

func main() {
	flag.Parse()
	// single thread
	runtime.GOMAXPROCS(1)

	m, err := segtool.Get(*mode)
	if err != nil {
		log.Fatal(err)
	}

	t0 := time.Now()
	var seg gse.Segmenter
	if err = segtool.LoadDict(&seg, *dict); err != nil {
		log.Fatal("Load dictionary error: ", err)
	}
	log.Println("Load dictionary time: ", time.Since(t0))

	lines, size, err := segtool.ReadLines(*file)
	if err != nil {
		log.Fatal(err)
	}

	var (
		out io.Writer
		bw  *bufio.Writer
	)
	if *output != "" {
		of, err := os.Create(*output)
		if err != nil {
			log.Fatal(err)
		}
		defer of.Close()
		bw = bufio.NewWriter(of)
		out = bw
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if err = pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
	}

	t1 := time.Now()
	if err = segmentLines(&seg, m.Func, lines, *runs, out); err != nil {
		log.Fatal(err)
	}
	cost := time.Since(t1)

	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}
	if bw != nil {
		if err = bw.Flush(); err != nil {
			log.Fatal(err)
		}
	}
	if *memprofile != "" {
		if err = writeMemProfile(*memprofile); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Mode %s, segment time %v", m.Name, cost)
	log.Printf("Segment speed %f MB/s", float64(size**runs)/cost.Seconds()/(1024*1024))
}
