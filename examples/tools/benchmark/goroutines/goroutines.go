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
The gse concurrent segmentation speed, run from the repository root:

	go run ./examples/tools/benchmark/goroutines
	go run ./examples/tools/benchmark/goroutines -mode hmm -workers 8
*/
package main

import (
	"flag"
	"log"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/examples/tools/internal/segtool"
)

var (
	file    = flag.String("file", "testdata/zh/bailuyuan_zh.txt", "text file to segment")
	dict    = flag.String("dict", "", "dictionary files, the embedded zh dictionary by default")
	mode    = flag.String("mode", segtool.Default, "segment mode: "+strings.Join(segtool.Names(), ", "))
	runs    = flag.Int("runs", 50, "number of runs")
	workers = flag.Int("workers", runtime.NumCPU(), "number of goroutines")
)

// segmentParallel segments the lines runs times by the workers,
// returns the number of segments
func segmentParallel(seg *gse.Segmenter, fn segtool.Func, lines []string, runs, workers int) int64 {
	var (
		wg    sync.WaitGroup
		count atomic.Int64
		task  = make(chan string, workers*40)
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range task {
				count.Add(int64(len(fn(seg, line))))
			}
		}()
	}

	for i := 0; i < runs; i++ {
		for _, l := range lines {
			task <- l
		}
	}
	close(task)
	wg.Wait()

	return count.Load()
}

func main() {
	flag.Parse()
	if *workers < 1 {
		log.Fatal("workers must be greater than 0")
	}

	m, err := segtool.Get(*mode)
	if err != nil {
		log.Fatal(err)
	}

	var seg gse.Segmenter
	if err = segtool.LoadDict(&seg, *dict); err != nil {
		log.Fatal("Load dictionary error: ", err)
	}

	lines, size, err := segtool.ReadLines(*file)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Start segment, mode %s, %d goroutines", m.Name, *workers)

	t0 := time.Now()
	n := segmentParallel(&seg, m.Func, lines, *runs, *workers)
	cost := time.Since(t0)

	log.Printf("Segment time %v, %d segments", cost, n)
	log.Printf("Segment speed %f MB/s", float64(size**runs)/cost.Seconds()/(1024*1024))
}
