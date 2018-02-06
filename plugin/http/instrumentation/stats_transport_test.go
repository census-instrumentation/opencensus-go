// Copyright 2018, OpenCensus Authors
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

package instrumentation

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"go.opencensus.io/stats"
)

const reqCount = 5

func TestClientStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("Hello, world!"))
	}))
	defer server.Close()

	views := []string {
		"opencensus.io/http/client/requests",
		"opencensus.io/http/client/latency",
		"opencensus.io/http/client/request_size",
		"opencensus.io/http/client/response_size",
	}
	for _, name := range views {
		v := stats.FindView(name)
		if v == nil {
			t.Errorf("view not found %q", name)
			continue
		}
		v.Subscribe()
	}

	var (
		w  sync.WaitGroup
		tr statsTransport
	)
	w.Add(reqCount)
	for i := 0; i < reqCount; i++ {
		go func() {
			req, err := http.NewRequest("POST", server.URL, strings.NewReader("req-body"))
			if err != nil {
				t.Fatalf("error creating request: %v", err)
			}
			resp, err := tr.RoundTrip(req)
			if err != nil {
				t.Fatalf("response error: %v", err)
			}
			if got, want := resp.StatusCode, 200; got != want {
				t.Fatalf("resp.StatusCode=%d; wantCount %d", got, want)
			}
			ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			w.Done()
		}()
	}
	w.Wait()

	for _, viewName := range views {
		v := stats.FindView(viewName)
		if v == nil {
			t.Errorf("view not found %q", viewName)
			continue
		}
		rows, err := v.RetrieveData()
		if err != nil {
			t.Error(err)
			continue
		}
		if got, want := len(rows), 1; got != want {
			t.Errorf("len(%q) = %d; want %d", viewName, got, want)
			continue
		}
		data := rows[0].Data
		var count int64
		switch data := data.(type) {
		case *stats.CountData:
			count = *(*int64)(data)
		case *stats.DistributionData:
			count = data.Count
		default:
			t.Errorf("don't know how to handle data type: %v", data)
		}
		if got := count; got != reqCount {
			t.Fatalf("%s = %d; want %d", viewName, got, reqCount)
		}
	}
}
