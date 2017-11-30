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
	"go.opencensus.io/stats"
)

// Exporter exports stats to Prometheus, users need
// to register the exporter as an http.Handler to be
// able to export.
type Exporter struct {
	opts    Options
	g       prometheus.Gatherer
	c       *collector
	handler http.Handler
}

// Options contains options for configuring the exporter.
type Options struct {
	OnError func(err error)
}

type collector struct {
	descs   []*prometheus.Desc
	metrics []prometheus.Metric
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	panic("not implemented")
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	panic("not implemented")
}

// NewExporter returns an exporter that exports stats to Prometheus.
func NewExporter(o Options) (*Exporter, error) {
	r := prometheus.NewRegistry()
	collector := &collector{}
	e := &Exporter{
		opts:    o,
		g:       r,
		c:       collector,
		handler: promhttp.HandlerFor(r, promhttp.HandlerOpts{}),
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
	// TODO(jbd,odeke-em): Make sure Export is not blocking
	// for a long period of time.
	panic("not implemented")
}

// ServeHTTP serves the Prometheus endpoint.
func (e *Exporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handler.ServeHTTP(w, r)
}
