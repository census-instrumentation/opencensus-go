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
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"

	"go.opencensus.io/internal"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultNamespace = "opencensus"
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
	Namespace string
	OnError   func(err error)
}

// NewExporter returns an exporter that exports stats to Prometheus.
func NewExporter(o Options) (*Exporter, error) {
	if o.Namespace == "" {
		o.Namespace = defaultNamespace
	}
	reg := prometheus.NewRegistry()
	collector := newCollector(o, reg)
	e := &Exporter{
		opts:    o,
		g:       reg,
		c:       collector,
		handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
	}
	return e, nil
}

var _ http.Handler = (*Exporter)(nil)
var _ stats.Exporter = (*Exporter)(nil)

func (c *collector) registerViews(views ...*stats.View) {
	count := 0
	for _, view := range views {
		sig := viewSignature(c.opts.Namespace, view)
		c.registeredViewsMu.Lock()
		_, ok := c.registeredViews[sig]
		c.registeredViewsMu.Unlock()

		if !ok {
			desc := prometheus.NewDesc(
				viewName(c.opts.Namespace, view),
				view.Description(),
				tagKeysToLabels(view.TagKeys()),
				nil,
			)
			c.registeredViewsMu.Lock()
			c.registeredViews[sig] = desc
			c.registeredViewsMu.Unlock()
			count++
		}
	}
	if count == 0 {
		return
	}

	c.reg.Unregister(c)
	if err := c.reg.Register(c); err != nil {
		c.opts.onError(fmt.Errorf("cannot register the collector: %v", err))
	}
}

func (o *Options) onError(err error) {
	if o.OnError != nil {
		o.OnError(err)
	} else {
		log.Printf("Failed to export to Prometheus: %v", err)
	}
}

// Export exports to the Prometheus if view data has one or more rows.
func (e *Exporter) Export(vd *stats.ViewData) {
	if len(vd.Rows) == 0 {
		return
	}
	e.c.addViewData(vd)
}

// ServeHTTP serves the Prometheus endpoint.
func (e *Exporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handler.ServeHTTP(w, r)
}

// collector implements prometheus.Collector
type collector struct {
	opts Options
	mu   sync.Mutex // mu guards all the fields.

	// reg helps collector register views dyanmically.
	reg *prometheus.Registry

	// viewData are accumulated and atomically
	// appended to on every Export invocation, from
	// stats. These views are cleared out when
	// Collect is invoked and the cycle is repeated.
	viewData []*stats.ViewData

	registeredViewsMu sync.Mutex
	// registeredViews maps a view to a prometheus desc.
	registeredViews map[string]*prometheus.Desc

	// seenMetrics maps from the metric's rawType to the actual Metric.
	// It is an interface to interface mapping
	// but the key is the zero value while the value is the instance.
	seenMetrics map[stats.AggregationData]prometheus.Metric
}

func (c *collector) addViewData(vd *stats.ViewData) {
	c.registerViews(vd.View)

	c.mu.Lock()
	c.viewData = append(c.viewData, vd)
	c.mu.Unlock()
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	c.registeredViewsMu.Lock()
	registered := make(map[string]*prometheus.Desc)
	for k, desc := range c.registeredViews {
		registered[k] = desc
	}
	c.registeredViewsMu.Unlock()

	for _, desc := range registered {
		ch <- desc
	}
}

func (c *collector) lookupMetric(key stats.AggregationData) (prometheus.Metric, bool) {
	c.mu.Lock()
	value, ok := c.seenMetrics[key]
	c.mu.Unlock()
	return value, ok
}

func (c *collector) memoizeMetric(key stats.AggregationData, value prometheus.Metric) {
	c.mu.Lock()
	c.seenMetrics[key] = value
	c.mu.Unlock()
}

// Collect fetches the statistics from OpenCensus
// and delivers them as Prometheus Metrics.
// Collect is invoked everytime a prometheus.Gatherer is run
// for example when the HTTP endpoint is invoked by Prometheus.
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, vd := range c.viewData {
		for _, row := range vd.Rows {
			metric, err := c.toMetric(vd.View, row)
			if err != nil {
				c.opts.onError(err)
			} else {
				ch <- metric
			}
		}
	}
}

func (c *collector) toMetric(view *stats.View, row *stats.Row) (prometheus.Metric, error) {
	switch agg := view.Aggregation().(type) {
	case stats.CountAggregation:
		data := row.Data.(*stats.CountData)
		var key *stats.CountData
		sc, ok := c.lookupMetric(key)
		if !ok {
			sc = prometheus.NewCounter(prometheus.CounterOpts{
				Name:      internal.Sanitize(view.Name()),
				Help:      view.Description(),
				Namespace: c.opts.Namespace,
			})
			c.memoizeMetric(key, sc)
		}
		counter := sc.(prometheus.Counter)
		counter.Add(float64(*data))
		return counter, nil

	case stats.DistributionAggregation:
		data := row.Data.(*stats.DistributionData)
		sig := viewSignature(c.opts.Namespace, view)

		c.registeredViewsMu.Lock()
		desc := c.registeredViews[sig]
		c.registeredViewsMu.Unlock()

		var tagValues []string
		for _, t := range row.Tags {
			tagValues = append(tagValues, t.Value)
		}

		points := make(map[float64]uint64)
		for i, b := range agg {
			points[b] = uint64(data.CountPerBucket[i])
		}
		hist, err := prometheus.NewConstHistogram(desc, uint64(data.Count), data.Sum(), points, tagValues...)
		if err != nil {
			return nil, err
		}
		return hist, nil

	case stats.SumAggregation:
		panic("stats.SumData not supported yet")

	case *stats.MeanAggregation:
		panic("stats.MeanData ont supported yet")

	default:
		return nil, fmt.Errorf("aggregation %T is not yet supported", view.Aggregation())
	}
}

func tagKeysToLabels(keys []tag.Key) (labels []string) {
	for _, key := range keys {
		labels = append(labels, internal.Sanitize(key.Name()))
	}
	return labels
}

func tagsToLabels(tags []tag.Tag) []string {
	var names []string
	for _, tag := range tags {
		names = append(names, internal.Sanitize(tag.Key.Name()))
	}
	return names
}

func newCollector(opts Options, registrar *prometheus.Registry) *collector {
	return &collector{
		reg:             registrar,
		opts:            opts,
		registeredViews: make(map[string]*prometheus.Desc),
		seenMetrics:     make(map[stats.AggregationData]prometheus.Metric),
	}
}

func viewName(namespace string, v *stats.View) string {
	return namespace + "_" + internal.Sanitize(v.Name())
}

func viewSignature(namespace string, v *stats.View) string {
	var buf bytes.Buffer
	buf.WriteString(viewName(namespace, v))
	for _, k := range v.TagKeys() {
		buf.WriteString("-" + k.Name())
	}
	return buf.String()
}
