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

package metricexport

import (
	"fmt"
	"time"

	"context"
	"go.opencensus.io/metric"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricproducer"
	"go.opencensus.io/trace"
	"sync"
)

var (
	defaultSampler = trace.ProbabilitySampler(0.0001)
)

// IntervalReader periodically reads metrics from all producers registered
// with producer manager and exports those metrics using provided
// exporter. Call Reader.Stop() to stop the reader.
type IntervalReader struct {
	exporter   metric.Exporter
	timer      *time.Ticker
	quit, done chan bool
	mu         sync.RWMutex
	reader     *Reader
	options    Options
}

// Reader reads metrics from all producers registered
// with producer manager and exports those metrics using provided
// exporter.
type Reader struct {
	Sampler trace.Sampler

	// SpanName is the name used for span created to export metrics.
	SpanName string
}

// Options to configure optional parameters for Reader.
type Options struct {
	// ReportingInterval sets the interval between reporting metrics.
	// If it is set to zero then default value is used.
	ReportingInterval time.Duration
}

const (
	// DefaultReportingDuration is default reporting duration.
	DefaultReportingDuration = 60 * time.Second

	// MinimumReportingDuration represents minimum value of reporting duration
	MinimumReportingDuration = 1 * time.Second

	// DefaultSpanName is the default name of the span generated
	// for reading and exporting metrics.
	DefaultSpanName = "ExportMetrics"
)

// NewIntervalReader creates a reader and starts a go routine
// that periodically reads metrics from all producers
// and exports them using provided exporter.
// Use options to specify periodicity.
func NewIntervalReader(reader *Reader, exporter metric.Exporter, options Options) (*IntervalReader, error) {
	if exporter == nil {
		return nil, fmt.Errorf("exporter is nil")
	}
	if reader == nil {
		return nil, fmt.Errorf("reader is nil")
	}

	if options.ReportingInterval == 0 {
		options.ReportingInterval = DefaultReportingDuration
	} else {
		if options.ReportingInterval.Seconds() < MinimumReportingDuration.Seconds() {
			return nil, fmt.Errorf("invalid reporting duration %f, minimum should be %f",
				options.ReportingInterval.Seconds(), MinimumReportingDuration.Seconds())
		}
	}

	r := &IntervalReader{
		exporter: exporter,
		timer:    time.NewTicker(options.ReportingInterval),
		quit:     make(chan bool),
		done:     make(chan bool),
		options:  options,
		reader:   reader,
	}
	go r.start()
	return r, nil
}

func (ir *IntervalReader) start() {
	for {
		select {
		case <-ir.timer.C:
			ir.reader.ReadAndExport(ir.exporter)
		case <-ir.quit:
			ir.timer.Stop()
			ir.done <- true
			return
		}
	}
}

// Stop stops the reader from reading and exporting metrics.
// Additional call to Stop are no-ops.
func (ir *IntervalReader) Stop() {
	if ir == nil {
		return
	}
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if ir.quit == nil {
		return
	}
	ir.quit <- true
	<-ir.done
	ir.quit = nil
}

// ReadAndExport reads metrics from all producer registered with
// producer manager and then exports them using provided exporter.
func (r *Reader) ReadAndExport(exporter metric.Exporter) {
	spanName := DefaultSpanName
	sampler := defaultSampler
	if r.SpanName == "" {
		spanName = r.SpanName
	}
	if r.Sampler != nil {
		sampler = r.Sampler
	}

	ctx := context.Background()
	_, span := trace.StartSpan(
		ctx,
		spanName,
		trace.WithSampler(sampler),
	)
	defer span.End()
	producers := metricproducer.GlobalManager().GetAll()
	data := []*metricdata.Metric{}
	for _, producer := range producers {
		data = append(data, producer.Read()...)
	}
	exporter.ExportMetric(ctx, data)
}
