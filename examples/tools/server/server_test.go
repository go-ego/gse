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

package main

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/go-ego/gse"
	"github.com/go-ego/gse/examples/tools/internal/segtool"
	"github.com/vcaesar/tt"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	seg := &gse.Segmenter{SkipLog: true}
	tt.Nil(t, seg.LoadDictStr("服务器 100 n\n指令 100 n\n你好 2000 l\n世界 2000 n\n"))

	static, err := fs.Sub(staticFS, "static")
	tt.Nil(t, err)
	s := &server{seg: seg, mode: segtool.Default}
	return s.handler(static)
}

func do(t *testing.T, h http.Handler, method, target string, form url.Values, v any) int {
	t.Helper()
	var req *http.Request
	if method == http.MethodPost {
		req = httptest.NewRequest(method, target, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req = httptest.NewRequest(method, target+"?"+form.Encode(), nil)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if v != nil {
		tt.Equal(t, "application/json", w.Header().Get("Content-Type"))
		tt.Nil(t, json.Unmarshal(w.Body.Bytes(), v))
	}
	return w.Code
}

func segText(r JsonResponse) string {
	var ss []string
	for _, s := range r.Segments {
		ss = append(ss, s.Text+"/"+s.Pos)
	}
	return strings.Join(ss, " ")
}

func TestSegment(t *testing.T) {
	h := newTestServer(t)

	var r JsonResponse
	code := do(t, h, http.MethodPost, "/json", url.Values{"text": {"服务器指令"}}, &r)
	tt.Equal(t, http.StatusOK, code)
	tt.Equal(t, "segment", r.Mode)
	tt.Equal(t, "服务器/n 指令/n", segText(r))

	// the cut modes fill the pos from the dictionary
	r = JsonResponse{}
	code = do(t, h, http.MethodGet, "/json", url.Values{"text": {"你好世界"}, "mode": {"hmm"}}, &r)
	tt.Equal(t, http.StatusOK, code)
	tt.Equal(t, "hmm", r.Mode)
	tt.Equal(t, "你好/l 世界/n", segText(r))

	for _, m := range segtool.Names() {
		r = JsonResponse{}
		code = do(t, h, http.MethodGet, "/json", url.Values{"text": {"你好世界"}, "mode": {m}}, &r)
		tt.Equal(t, http.StatusOK, code, m)
		tt.Bool(t, len(r.Segments) > 0, m)
	}

	r = JsonResponse{}
	do(t, h, http.MethodGet, "/json", nil, &r)
	tt.Equal(t, 0, len(r.Segments))

	var resp Resp
	code = do(t, h, http.MethodGet, "/json", url.Values{"text": {"你好"}, "mode": {"nope"}}, &resp)
	tt.Equal(t, http.StatusBadRequest, code)
	tt.Equal(t, http.StatusBadRequest, resp.Code)
	tt.Bool(t, strings.Contains(resp.Text, "unknown mode"))
}

func TestAdd(t *testing.T) {
	h := newTestServer(t)

	var resp Resp
	code := do(t, h, http.MethodGet, "/add", url.Values{"text": {"你好世界"}, "freq": {"10000"}, "pos": {"l"}}, &resp)
	tt.Equal(t, http.StatusOK, code)
	tt.Equal(t, Resp{Code: http.StatusOK, Text: "ok"}, resp)

	var r JsonResponse
	do(t, h, http.MethodGet, "/json", url.Values{"text": {"你好世界"}}, &r)
	tt.Equal(t, "你好世界/l", segText(r))

	// without freq, the suggested freq keeps the word
	code = do(t, h, http.MethodPost, "/add", url.Values{"text": {"服务器指令"}, "pos": {"nz"}}, &resp)
	tt.Equal(t, http.StatusOK, code)
	r = JsonResponse{}
	do(t, h, http.MethodGet, "/json", url.Values{"text": {"服务器指令"}}, &r)
	tt.Equal(t, "服务器指令/nz", segText(r))

	code = do(t, h, http.MethodGet, "/add", url.Values{"text": {" "}}, &resp)
	tt.Equal(t, http.StatusBadRequest, code)
	tt.Equal(t, "text is empty", resp.Text)

	for _, f := range []string{"abc", "0", "-1"} {
		code = do(t, h, http.MethodGet, "/add", url.Values{"text": {"指令"}, "freq": {f}}, &resp)
		tt.Equal(t, http.StatusBadRequest, code, f)
	}
}

func TestModes(t *testing.T) {
	h := newTestServer(t)

	var modes []segtool.Mode
	code := do(t, h, http.MethodGet, "/modes", nil, &modes)
	tt.Equal(t, http.StatusOK, code)
	tt.Equal(t, len(segtool.Modes), len(modes))
	tt.Equal(t, segtool.Default, modes[0].Name)
	tt.Bool(t, modes[0].Desc != "")
}

func TestStatic(t *testing.T) {
	h := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	tt.Equal(t, http.StatusOK, w.Code)
	tt.Bool(t, strings.Contains(w.Body.String(), "<title>gse segment 演示</title>"))
}

func TestConcurrent(t *testing.T) {
	h := newTestServer(t)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			var r JsonResponse
			tt.Equal(t, http.StatusOK, do(t, h, http.MethodGet, "/json",
				url.Values{"text": {"你好世界, 服务器指令"}, "mode": {"search_hmm"}}, &r))
		}()
		go func() {
			defer wg.Done()
			var resp Resp
			tt.Equal(t, http.StatusOK, do(t, h, http.MethodGet, "/add",
				url.Values{"text": {"指令集"}, "freq": {"10"}}, &resp))
		}()
	}
	wg.Wait()
}
