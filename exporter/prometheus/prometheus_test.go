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

package prometheus

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go.opencensus.io/exporter"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/prometheus/client_golang/prometheus"
)

func newViewData(measureName string, agg exporter.Aggregation, rows []*exporter.Row) *exporter.ViewData {
	m := stats.Int64(measureName, "bytes", stats.UnitBytes)
	return &exporter.ViewData{
		Name:        "foo",
		Description: "bar",
		Unit:        agg.Type.AggregatedUnit(m.Unit()),
		Aggregation: agg,
		TagKeys:     nil,
		Rows:        rows,
		Start:       time.Now(),
		End:         time.Now(),
	}
}

func TestOnlyCumulativeWindowSupported(t *testing.T) {
	// See Issue https://github.com/census-instrumentation/opencensus-go/issues/214.
	count1 := exporter.AggregationData{Count: 1}
	tests := []struct {
		vds  *exporter.ViewData
		want int
	}{
		0: {
			vds:  newViewData("TestOnlyCumulativeWindowSupported/m1", exporter.Aggregation{Type: exporter.AggTypeCount}, nil),
			want: 0, // no rows present
		},
		1: {
			vds: newViewData("TestOnlyCumulativeWindowSupported/m2", exporter.Aggregation{Type: exporter.AggTypeCount},
				[]*exporter.Row{
					{Data: count1},
				}),
			want: 1,
		},
	}

	for i, tt := range tests {
		reg := prometheus.NewRegistry()
		collector := newCollector(Options{}, reg)
		collector.addViewData(tt.vds)
		mm, err := reg.Gather()
		if err != nil {
			t.Errorf("#%d: Gather err: %v", i, err)
		}
		reg.Unregister(collector)
		if got, want := len(mm), tt.want; got != want {
			t.Errorf("#%d: got nil %v want nil %v", i, got, want)
		}
	}
}

func TestSingletonExporter(t *testing.T) {
	exp, err := NewExporter(Options{})
	if err != nil {
		t.Fatalf("NewExporter() = %v", err)
	}
	if exp == nil {
		t.Fatal("Nil exporter")
	}

	// Should all now fail
	exp, err = NewExporter(Options{})
	if err == nil {
		t.Fatal("NewExporter() = nil")
	}
	if exp != nil {
		t.Fatal("Non-nil exporter")
	}
}

func TestCollectNonRacy(t *testing.T) {
	// Despite enforcing the singleton, for this case we
	// need an exporter hence won't be using NewExporter.
	exp, err := newExporter(Options{})
	if err != nil {
		t.Fatalf("NewExporter: %v", err)
	}
	collector := exp.c

	// Synchronize and make sure every goroutine has terminated before we exit
	var waiter sync.WaitGroup
	waiter.Add(3)
	defer waiter.Wait()

	doneCh := make(chan bool)
	// 1. Viewdata write routine at 700ns
	go func() {
		defer waiter.Done()
		tick := time.NewTicker(700 * time.Nanosecond)
		defer tick.Stop()

		defer func() {
			close(doneCh)
		}()

		for i := 0; i < 1e3; i++ {
			count1 := exporter.AggregationData{Count: 1}
			vds := []*exporter.ViewData{
				newViewData(fmt.Sprintf("TestCollectNonRacy/m2-%d", i), exporter.Aggregation{Type: exporter.AggTypeCount}, []*exporter.Row{{Data: count1}}),
			}
			for _, v := range vds {
				exp.ExportView(v)
			}
			<-tick.C
		}
	}()

	inMetricsChan := make(chan prometheus.Metric, 1000)
	// 2. Simulating the Prometheus metrics consumption routine running at 900ns
	go func() {
		defer waiter.Done()
		tick := time.NewTicker(900 * time.Nanosecond)
		defer tick.Stop()

		for {
			select {
			case <-doneCh:
				return
			case <-inMetricsChan:
			}
		}
	}()

	// 3. Collect/Read routine at 800ns
	go func() {
		defer waiter.Done()
		tick := time.NewTicker(800 * time.Nanosecond)
		defer tick.Stop()

		for {
			select {
			case <-doneCh:
				return
			case <-tick.C:
				// Perform some collection here
				collector.Collect(inMetricsChan)
			}
		}
	}()
}

type mSlice []*stats.Int64Measure

func (measures *mSlice) createAndAppend(name, desc, unit string) {
	m := stats.Int64(name, desc, unit)
	*measures = append(*measures, m)
}

type vCreator []*view.View

func (vc *vCreator) createAndAppend(name, description string, keys []tag.Key, measure stats.Measure, agg *view.Aggregation) {
	v := &view.View{
		Name:        name,
		Description: description,
		TagKeys:     keys,
		Measure:     measure,
		Aggregation: agg,
	}
	*vc = append(*vc, v)
}

func TestMetricsEndpointOutput(t *testing.T) {
	e, err := newExporter(Options{})
	if err != nil {
		t.Fatalf("failed to create prometheus exporter: %v", err)
	}
	exporter.Register(e)

	names := []string{"foo", "bar", "baz"}

	var measures mSlice
	for _, name := range names {
		measures.createAndAppend("tests/"+name, name, "")
	}

	var vc vCreator
	for _, m := range measures {
		vc.createAndAppend(m.Name(), m.Description(), nil, m, view.Count())
	}

	if err := view.Subscribe(vc...); err != nil {
		t.Fatalf("failed to create views: %v", err)
	}
	view.SetReportingPeriod(time.Millisecond)

	for _, m := range measures {
		stats.Record(context.Background(), m.M(1))
	}

	srv := httptest.NewServer(e)
	defer srv.Close()

	time.Sleep(10 * time.Millisecond)

	var i int
	var output string
	for {
		if i == 10000 {
			t.Fatal("no output at /metrics (10s wait)")
		}
		i++

		resp, err := http.Get(srv.URL)
		if err != nil {
			t.Fatalf("failed to get /metrics: %v", err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		resp.Body.Close()

		output = string(body)
		if output != "" {
			break
		}
		time.Sleep(time.Millisecond)
	}

	if strings.Contains(output, "collected before with the same name and label values") {
		t.Errorf("metric name and labels being duplicated but must be unique")
	}

	if strings.Contains(output, "error(s) occurred") {
		t.Errorf("error reported by prometheus registry")
	}

	for _, name := range names {
		if !strings.Contains(output, "opencensus_tests_"+name+" 1") {
			t.Errorf("measurement missing in output: %v", name)
		}
	}
}
