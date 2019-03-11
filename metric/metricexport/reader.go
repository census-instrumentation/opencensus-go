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

// Reader periodically reads metrics from all producers registered
// with producer manager and exports those metrics using provided
// exporter. Call Reader.Stop() to stop the reader.
type Reader struct {
	exporter   metric.Exporter
	timer      *time.Ticker
	quit, done chan bool
	sampler    trace.Sampler
	options    Options
	mu         sync.RWMutex
}

// Options to configure optional parameters for Reader.
type Options struct {
	// ReportingInterval sets the interval between reporting metrics.
	// If it is set to zero then default value is used.
	ReportingInterval time.Duration

	// SpanName is the name used for span created to export metrics.
	SpanName string
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

// NewReader creates a reader and starts a go routine
// that periodically reads metrics from all producers
// and exports them using provided exporter.
// Use options to specify periodicity.
func NewReader(exporter metric.Exporter, options Options) (*Reader, error) {
	if exporter == nil {
		return nil, fmt.Errorf("exporter is nil")
	}
	if options.ReportingInterval == 0 {
		options.ReportingInterval = DefaultReportingDuration
	} else {
		if options.ReportingInterval.Seconds() < MinimumReportingDuration.Seconds() {
			return nil, fmt.Errorf("invalid reporting duration %f, minimum should be %f",
				options.ReportingInterval.Seconds(), MinimumReportingDuration.Seconds())
		}
	}
	if options.SpanName == "" {
		options.SpanName = DefaultSpanName
	}

	r := &Reader{
		exporter: exporter,
		timer:    time.NewTicker(options.ReportingInterval),
		quit:     make(chan bool),
		done:     make(chan bool),
		sampler:  trace.ProbabilitySampler(0.0001),
		options:  options,
	}
	go r.start()
	return r, nil
}

func (r *Reader) start() {
	for {
		select {
		case <-r.timer.C:
			r.readAndExport(time.Now())
		case <-r.quit:
			r.timer.Stop()
			r.done <- true
			return
		}
	}
}

// Stop stops the reader from reading and exporting metrics.
// Additional call to Stop are no-ops.
func (r *Reader) Stop() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.quit == nil {
		return
	}
	r.quit <- true
	<-r.done
	r.quit = nil
}

func (r *Reader) readAndExport(now time.Time) {
	ctx := context.Background()
	_, span := trace.StartSpan(
		ctx,
		r.options.SpanName,
		trace.WithSampler(r.sampler),
	)
	defer span.End()
	producers := metricproducer.GlobalManager().GetAll()
	data := []*metricdata.Metric{}
	for _, producer := range producers {
		data = append(data, producer.Read()...)
	}
	r.exporter.ExportMetric(data)
}
