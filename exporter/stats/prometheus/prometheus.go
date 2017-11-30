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

// Package prometheus contains the Prometheus exporters for
// Stackdriver Monitoring.
//
// Please note that this exporter is currently work in progress and not complete.
package prometheus

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opencensus.io/internal"
	"go.opencensus.io/stats"
)

// Exporter exports stats to Prometheus, users need
// to register the exporter as an http.Handler to be
// able to export.
type Exporter struct {
	opts Options
	g    prometheus.Gatherer
	c    *collector
}

// Options contains options for configuring the exporter.
type Options struct {
	OnError func(err error)
}

type collector struct {
	descCh   chan *prometheus.Desc
	metricCh chan prometheus.Metric
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- <-c.descCh
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	ch <- <-c.metricCh
}

// NewExporter returns an exporter that exports stats to Prometheus.
func NewExporter(o Options) (*Exporter, error) {
	r := prometheus.NewRegistry()
	collector := &collector{
		descCh:   make(chan *prometheus.Desc),
		metricCh: make(chan prometheus.Metric),
	}
	e := &Exporter{
		opts: o,
		g:    r,
		c:    collector,
	}
	go func() {
		if err := r.Register(collector); err != nil {
			e.onError(err)
		}
	}()
	// TODO(jbd): Implement a Close function to unregister.
	return e, nil
}

func (e *Exporter) onError(err error) {
	if e.opts.OnError != nil {
		e.opts.OnError(err)
		return
	}
	log.Printf("Failed to export to Prometheus: %v", err)
}

// Export exports to the Prometheus if view data has one or more rows.
func (e *Exporter) Export(vd *stats.ViewData) {
	if len(vd.Rows) == 0 {
		return
	}
	go e.export(vd)
}

func (e *Exporter) export(vd *stats.ViewData) {
	view := vd.View
	for _, r := range vd.Rows {
		var labels []string
		for _, t := range r.Tags {
			labels = append(labels, internal.Sanitize(t.Key.Name()))
		}
		// Sanitize the view name.
		desc := prometheus.NewDesc(view.Name(), view.Description(), labels, nil)
		e.c.descCh <- desc

		// TODO(jbd): Support other metric types.
		switch v := r.Data.(type) {
		case *stats.CountData:
			e.c.metricCh <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, float64(*v))
		}
	}
}

// ServeHTTP serves the Prometheus endpoint.
func (e *Exporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	promhttp.HandlerFor(e.g, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}
