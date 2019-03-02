// Copyright 2019, OpenCensus Authors
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
//

package reader

import (
	"sync"
	"testing"
	"time"

	"go.opencensus.io/metric"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/producer"
)

var reader *Reader
var exporter = &metricExporter{}

func TestNewReader(t *testing.T) {
	restart(t)

	r := metric.NewRegistry()

	producer.GlobalManager().AddProducer(r)

	g := r.AddInt64Gauge("active_request", "Number of active requests, per method.", metricdata.UnitDimensionless, "method")

	e := g.GetEntry(metricdata.NewLabelValue("foo"))
	e.Add(1)

	time.Sleep(2 * time.Second)

	exporter.Lock()
	if len(exporter.metrics) == 0 {
		t.Fatal("Got no view data; want at least one")
	}
	want := "active_request"
	for _, metric := range exporter.metrics {
		got := metric.Descriptor.Name
		if got != want {
			t.Errorf("got %s, want %s\n", got, want)
		}
	}
	exporter.metrics = nil
	exporter.Unlock()
}

func TestNewReaderWithNilExporter(t *testing.T) {

	_, err := NewReader(nil, Options{})
	if err == nil {
		t.Fatalf("expected error but got nil\n")
	}
}

func TestNewReaderWithInvalidOption(t *testing.T) {

	_, err := NewReader(nil, Options{500 * time.Millisecond, ""})
	if err == nil {
		t.Fatalf("expected error but got nil\n")
	}
}

type metricExporter struct {
	sync.Mutex
	metrics []*metricdata.Metric
}

func (e *metricExporter) ExportMetric(metrics []*metricdata.Metric) {
	e.Lock()
	defer e.Unlock()

	e.metrics = append(e.metrics, metrics...)
}

// restart stops the current processors and creates a new one.
func restart(t *testing.T) {
	if reader != nil {
		reader.Stop()
	}
	r, err := NewReader(exporter, Options{1 * time.Second, ""})
	if err != nil {
		t.Fatalf("error creating reader %v\n", err)
	}
	reader = r
}
