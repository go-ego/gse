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
The gse segmentation server, run from the repository root:

	go run ./examples/tools/server
	go run ./examples/tools/server -port 8080 -mode hmm -dict "zh, your_dict.txt"

Routes:

	"/"       the demo web page
	"/json"   segment the text, GET or POST
		input:  text, mode (optional, the -mode flag by default)
		output: {"mode":"segment","segments":[{"text":"服务器","pos":"n"}, ...]}
	"/add"    add a word to the dictionary, GET or POST
		input:  text, freq (optional, suggested by default), pos (optional)
		output: {"code":200,"text":"ok"}
	"/modes"  list the segmentation modes
		output: [{"name":"segment","desc":"..."}, ...]

Errors return a non 200 status with {"code":400,"text":"error message"}.
*/
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/examples/tools/internal/segtool"
)

// maxBodySize limits the request body size
const maxBodySize = 4 << 20

//go:embed static
var staticFS embed.FS

var (
	host = flag.String("host", "", "HTTP server host")
	port = flag.Int("port", 8080, "HTTP server port")

	mode = flag.String("mode", segtool.Default, "default segment mode: "+strings.Join(segtool.Names(), ", "))
	hmm  = flag.Bool("hmm", false, "use the hmm mode by default, same as -mode hmm")
	dict = flag.String("dict", "", "dictionary files, the embedded zh dictionary by default")

	staticFolder = flag.String("static_folder", "", "static web page directory, the embedded page by default")
)

// Segment segment json struct
type Segment struct {
	Text string `json:"text"`
	Pos  string `json:"pos"`
}

// JsonResponse the segment json response
type JsonResponse struct {
	Mode     string     `json:"mode"`
	Segments []*Segment `json:"segments"`
}

// Resp http response
type Resp struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type server struct {
	// mu guards seg, adding a word rewrites the dictionary
	mu   sync.RWMutex
	seg  *gse.Segmenter
	mode string
}

func (s *server) handler(static fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/json", s.segment)
	mux.HandleFunc("/add", s.add)
	mux.HandleFunc("/modes", s.modes)
	mux.Handle("/", http.FileServerFS(static))
	return mux
}

func (s *server) segment(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, maxBodySize)
	name := req.FormValue("mode")
	if name == "" {
		name = s.mode
	}

	m, err := segtool.Get(name)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	text := req.FormValue("text")
	s.mu.RLock()
	segs := m.Func(s.seg, text)
	ss := make([]*Segment, len(segs))
	for i, sp := range segs {
		pos := sp.Pos
		if pos == "" {
			_, pos, _ = s.seg.Find(sp.Text)
		}
		ss[i] = &Segment{Text: sp.Text, Pos: pos}
	}
	s.mu.RUnlock()

	writeJSON(w, http.StatusOK, &JsonResponse{Mode: name, Segments: ss})
}

func (s *server) add(w http.ResponseWriter, req *http.Request) {
	req.Body = http.MaxBytesReader(w, req.Body, maxBodySize)
	text := strings.TrimSpace(req.FormValue("text"))
	if text == "" {
		writeErr(w, http.StatusBadRequest, "text is empty")
		return
	}

	var freq float64
	if f := req.FormValue("freq"); f != "" {
		var err error
		freq, err = strconv.ParseFloat(f, 64)
		if err != nil || freq <= 0 {
			writeErr(w, http.StatusBadRequest, "freq must be a positive number")
			return
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if freq == 0 {
		freq = s.seg.SuggestFreq(text)
	}

	if err := s.seg.AddTokenForce(text, freq, req.FormValue("pos")); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &Resp{Code: http.StatusOK, Text: "ok"})
}

func (s *server) modes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, segtool.Modes)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, &Resp{Code: code, Text: msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("Write the response error: ", err)
	}
}

func main() {
	flag.Parse()
	if *hmm {
		*mode = "hmm"
	}
	if _, err := segtool.Get(*mode); err != nil {
		log.Fatal(err)
	}

	var seg gse.Segmenter
	if err := segtool.LoadDict(&seg, *dict); err != nil {
		log.Fatal("Load dictionary error: ", err)
	}

	static, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal(err)
	}
	if *staticFolder != "" {
		static = os.DirFS(*staticFolder)
	}

	s := &server{seg: &seg, mode: *mode}
	srv := &http.Server{
		Addr:              net.JoinHostPort(*host, strconv.Itoa(*port)),
		Handler:           s.handler(static),
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Println("Server listen: ", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
