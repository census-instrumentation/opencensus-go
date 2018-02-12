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

package rest_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"go.opencensus.io/trace"
)

var (
	bPong = []byte("pong")
	bPing = []byte("ping")
)

func tracedPong(w http.ResponseWriter, r *http.Request) {
	ctx := trace.StartSpan(r.Context(), "/pong")
	pongInnards(w, r)
	trace.EndSpan(ctx)
}

func untracedPong(w http.ResponseWriter, r *http.Request) {
	pongInnards(w, r)
}

func pongInnards(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	_, _ = ioutil.ReadAll(r.Body)

	w.Write(bPong)
}

func BenchmarkTracedPong2QPS(b *testing.B) {
	benchmarkIt(b, tracedPong, 2)
}

func BenchmarkTracedPong10QPS(b *testing.B) {
	benchmarkIt(b, tracedPong, 10)
}

func BenchmarkTracedPong100QPS(b *testing.B) {
	benchmarkIt(b, tracedPong, 100)
}

func BenchmarkTracedPong1000QPS(b *testing.B) {
	benchmarkIt(b, tracedPong, 1000)
}

func BenchmarkTracedPong5000QPS(b *testing.B) {
	benchmarkIt(b, tracedPong, 5000)
}

func BenchmarkUnTracedPong2QPS(b *testing.B) {
	benchmarkIt(b, untracedPong, 2)
}

func BenchmarkUnTracedPong10QPS(b *testing.B) {
	benchmarkIt(b, untracedPong, 10)
}

func BenchmarkUnTracedPong100QPS(b *testing.B) {
	benchmarkIt(b, untracedPong, 100)
}

func BenchmarkUnTracedPong1000QPS(b *testing.B) {
	benchmarkIt(b, untracedPong, 1000)
}

func BenchmarkUnTracedPong5000QPS(b *testing.B) {
	benchmarkIt(b, untracedPong, 5000)
}

func benchmarkIt(b *testing.B, h http.HandlerFunc, qps int) {
	cst := httptest.NewServer(h)
	defer cst.Close()

	client := cst.Client()
	url := cst.URL

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qpsIt(qps, func() {
			req, err := http.NewRequest("POST", url, bytes.NewReader(bPing))
			if err != nil {
				b.Fatalf("#%d err: %v", i, err)
			}
			res, err := client.Do(req)
			if err != nil {
				b.Fatalf("#%d err: %v", i, err)
			}
			_, _ = io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		})
	}
	b.ReportAllocs()
}

func qpsIt(qps int, fn func()) {
	// 1000 QPS ==> (1s/1000Q) * (1e9ns/1s) ==> 0.001s or 1e6ns/Q
	timeSlice := time.Duration(int64(float64(1/float64(qps)) * 1e9))
	tick := time.NewTicker(timeSlice)
	defer tick.Stop()

	var wg sync.WaitGroup
	defer wg.Wait()

	for i := 0; i < qps; i++ {
		wg.Add(1)
		go func() {
			fn()
			wg.Done()
		}()
		<-tick.C
	}
}
