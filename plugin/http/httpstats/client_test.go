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
package httpstats

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"strings"
	"go.opencensus.io/stats"
	"time"
	"io/ioutil"
)

type mockExporter map[string]stats.AggregationData

func (e mockExporter) Export(viewData *stats.ViewData) {
	// keep the last value, since all stats are cumulative
	e[viewData.View.Name()] = viewData.Rows[0].Data
}

func TestClientStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Write([]byte("Hello, world!"))
	}))
	defer server.Close()

	ClientLatencyDistribution.Subscribe()
	ClientRequestCount.Subscribe()
	ClientConnectionsOpenedCount.Subscribe()

	e := make(mockExporter)
	stats.RegisterExporter(&e)
	defer stats.UnregisterExporter(&e)

	tr := Transport{}

	for i := 0; i < 10; i++ {
		req, err := http.NewRequest("POST", server.URL, strings.NewReader("req-body"))
		if err != nil {
			t.Fatalf("error creating request: %#v", err)
		}
		resp, err := tr.RoundTrip(req)
		if err != nil {
			t.Fatalf("response error: %#v", err)
		}
		if got, want := resp.StatusCode, 200; got != want {
			t.Fatalf("resp.StatusCode=%d; want %d", got, want)
		}
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}

	stats.SetReportingPeriod(time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	stats.SetReportingPeriod(time.Second)
	stats.UnregisterExporter(&e)

	if len(e) == 0 {
		t.Fatalf("no viewdata received")
	}

	expect := []struct {
		name string
		want int64
	}{
		{name: "opencensus.io/http/client/started", want: int64(10)},
		{name: "opencensus.io/http/client/connections_opened", want: int64(1)},
		{name: "opencensus.io/http/client/latency", want: int64(10)},
	}
	for _, exp := range expect {
		switch data := e[exp.name].(type) {
		case *stats.CountData:
			if got := *(*int64)(data); got != exp.want {
				t.Fatalf("%q = %d; want %d", exp.name, got, exp.want)
			}
		case *stats.DistributionData:
			if got := data.Count; got != exp.want {
				t.Fatalf("%q = %d; want %d", exp.name, got, exp.want)
			}
		}
	}
}
