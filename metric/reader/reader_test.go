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

var (
	reader1    *Reader
	reader2    *Reader
	exporter1  = &metricExporter{}
	exporter2  = &metricExporter{}
	gaugeEntry *metric.Int64GaugeEntry
	options1   = Options{1000 * time.Millisecond, ""}
	options2   = Options{2000 * time.Millisecond, ""}
)

type metricExporter struct {
	sync.Mutex
	metrics []*metricdata.Metric
}

func (e *metricExporter) ExportMetric(metrics []*metricdata.Metric) {
	e.Lock()
	defer e.Unlock()

	e.metrics = append(e.metrics, metrics...)
}

func init() {
	r := metric.NewRegistry()
	producer.GlobalManager().AddProducer(r)
	g := r.AddInt64Gauge("active_request", "Number of active requests, per method.", metricdata.UnitDimensionless, "method")
	gaugeEntry = g.GetEntry(metricdata.NewLabelValue("foo"))
}

func TestNewReader(t *testing.T) {
	reader1 = restartReader(reader1, exporter1, options1, t)

	gaugeEntry.Add(1)

	time.Sleep(1500 * time.Millisecond)
	checkExportedCount(exporter1, 1, t)
	checkExportedMetricDesc(exporter1, "active_request", t)
	resetExporter(exporter1)
}

func TestProducerWithReaderStop(t *testing.T) {
	reader1 = restartReader(reader1, exporter1, options1, t)
	reader1.Stop()

	gaugeEntry.Add(1)

	time.Sleep(1500 * time.Millisecond)

	checkExportedCount(exporter1, 0, t)
	checkExportedMetricDesc(exporter1, "active_request", t)
	resetExporter(exporter1)
}

func TestProducerWithMultipleReaders(t *testing.T) {
	reader1 = restartReader(reader1, exporter1, options1, t)
	reader2 = restartReader(reader2, exporter2, options2, t)

	gaugeEntry.Add(1)

	time.Sleep(2500 * time.Millisecond)

	checkExportedCount(exporter1, 2, t)
	checkExportedMetricDesc(exporter1, "active_request", t)
	checkExportedCount(exporter2, 1, t)
	checkExportedMetricDesc(exporter2, "active_request", t)
	resetExporter(exporter1)
	resetExporter(exporter1)
}

func TestReaderMultipleStop(t *testing.T) {
	reader1 = restartReader(reader1, exporter1, options1, t)
	stop := make(chan bool, 1)
	go func() {
		reader1.Stop()
		reader1.Stop()
		stop <- true
	}()

	select {
	case _ = <-stop:
	case <-time.After(1 * time.Second):
		t.Fatalf("reader1 stop got blocked")
	}
}

func TestNewReaderWithNilExporter(t *testing.T) {
	_, err := NewReader(nil, Options{})
	if err == nil {
		t.Fatalf("expected error but got nil\n")
	}
}

func TestNewReaderWithInvalidOption(t *testing.T) {
	_, err := NewReader(exporter1, Options{500 * time.Millisecond, ""})
	if err == nil {
		t.Fatalf("expected error but got nil\n")
	}
}

func checkExportedCount(exporter *metricExporter, wantCount int, t *testing.T) {
	exporter.Lock()
	defer exporter.Unlock()
	gotCount := len(exporter.metrics)
	if gotCount != wantCount {
		t.Fatalf("exported metric count: got %d, want %d\n", gotCount, wantCount)
	}
}

func checkExportedMetricDesc(exporter *metricExporter, wantMdName string, t *testing.T) {
	exporter.Lock()
	defer exporter.Unlock()
	for _, metric := range exporter.metrics {
		gotMdName := metric.Descriptor.Name
		if gotMdName != wantMdName {
			t.Errorf("got %s, want %s\n", gotMdName, wantMdName)
		}
	}
	exporter.metrics = nil
}

func resetExporter(exporter *metricExporter) {
	exporter.Lock()
	defer exporter.Unlock()
	exporter.metrics = nil
}

// restartReader stops the current processors and creates a new one.
func restartReader(reader *Reader, exporter *metricExporter, options Options, t *testing.T) *Reader {
	if reader != nil {
		reader.Stop()
	}
	r, err := NewReader(exporter, options)
	if err != nil {
		t.Fatalf("error creating reader %v\n", err)
	}
	return r
}
