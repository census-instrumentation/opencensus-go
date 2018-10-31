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

// Package prometheus contains a Prometheus exporter that supports exporting
// OpenCensus views as Prometheus metrics.
package prometheus // import "go.opencensus.io/exporter/prometheus"

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opencensus.io/internal"
	"go.opencensus.io/metric"
	"go.opencensus.io/stats/view"
)

// Exporter exports stats to Prometheus, users need
// to register the exporter as an http.Handler to be
// able to export.
type Exporter struct {
	opts    Options
	g       prometheus.Gatherer
	c       prometheus.Collector
	handler http.Handler
}

// Options contains options for configuring the exporter.
type Options struct {
	Namespace   string
	Registry    *prometheus.Registry
	OnError     func(err error)
	ConstLabels prometheus.Labels // ConstLabels will be set as labels on all views.
}

// NewExporter returns an exporter that exports stats to Prometheus.
func NewExporter(o Options) (*Exporter, error) {
	if o.Registry == nil {
		o.Registry = prometheus.NewRegistry()
	}
	mc := NewMetricCollector()
	mc.OnError = o.OnError
	mc.GetName = func(m *metric.Metric) string {
		return viewName(o.Namespace, m.Descriptor.Name)
	}
	mc.ConstLabels = o.ConstLabels
	err := o.Registry.Register(mc)
	if err != nil {
		return nil, err
	}
	e := &Exporter{
		opts:    o,
		g:       o.Registry,
		c:       mc,
		handler: promhttp.HandlerFor(o.Registry, promhttp.HandlerOpts{}),
	}
	return e, nil
}

var _ http.Handler = (*Exporter)(nil)
var _ view.Exporter = (*Exporter)(nil)

func (o *Options) onError(err error) {
	if o.OnError != nil {
		o.OnError(err)
	} else {
		log.Printf("Failed to export to Prometheus: %v", err)
	}
}

// ExportView is a no-op. It exists for backwards compatibility.
// All view data will be exported on demand when requested by prometheus.
func (e *Exporter) ExportView(vd *view.Data) {
	// No-op. We always collect views via the metrics default registry on demand.
}

// ServeHTTP serves the Prometheus endpoint.
func (e *Exporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handler.ServeHTTP(w, r)
}

func viewName(namespace string, viewName string) string {
	var name string
	if namespace != "" {
		name = namespace + "_"
	}
	return name + internal.Sanitize(viewName)
}
