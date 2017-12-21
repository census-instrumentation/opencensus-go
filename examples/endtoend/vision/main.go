// Copyright 2017, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/net/context"

	"go.opencensus.io/trace"
)

var mux = http.NewServeMux()

var (
	addr string
)

func main() {
	log.Printf("Serving on: %q\n", addr)
	mux.HandleFunc("/upload", byRawUploads)
	mux.HandleFunc("/url", byURL)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("listenAndServe: %v", err)
	}
}

func byRawUploads(rw http.ResponseWriter, req *http.Request) {
	ctx := trace.StartSpan(req.Context(), "/upload")
	defer trace.EndSpan(ctx)

	if err := req.ParseMultipartForm(1 << 40); err != nil {
		// TODO: break these down to record specifically parse errors
		recordStatsErrorCount(ctx, 1)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		recordStatsErrorCount(ctx, 1)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	detectAndReply(file, ctx, rw)
}

func detectAndReply(r io.Reader, ctx context.Context, rw http.ResponseWriter) {
	res, err := detectFacesAndLogos(r, ctx)
	if err != nil {
		recordStatsErrorCount(ctx, 1)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	enc := json.NewEncoder(rw)
	if err := enc.Encode(res); err != nil {
		recordStatsErrorCount(ctx, 1)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

type urlIn struct {
	URL string `json:"url"`
}

func byURL(rw http.ResponseWriter, req *http.Request) {
	ctx := trace.StartSpan(req.Context(), "/url")
	defer trace.EndSpan(ctx)

	defer req.Body.Close()

	blob, err := ioutil.ReadAll(req.Body)
	if err != nil {
		recordStatsErrorCount(ctx, 1)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	ui := new(urlIn)
	if err := json.Unmarshal(blob, ui); err != nil {
		recordStatsErrorCount(ctx, 1)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	body, err := fetchIt(ui.URL, ctx)
	if err := json.Unmarshal(blob, ui); err != nil {
		recordStatsErrorCount(ctx, 1)
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
	detectAndReply(body, ctx, rw)
}

func fetchIt(url string, ctx context.Context) (io.ReadCloser, error) {
	dlCtx := trace.StartSpan(ctx, "/url-get")
	defer trace.EndSpan(dlCtx)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if code := res.StatusCode; code < 200 || code > 299 {
		return res.Body, fmt.Errorf("%s %d", res.Status, code)
	}
	return res.Body, nil
}
