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
var _ stats.ViewRegistrar = (*Exporter)(nil)

// RegisterView is the hook that's run either intentionally
// by the client or the stat's library will automatically invoke
// it everytime that a new view is added.
func (e *Exporter) RegisterView(view *stats.View) {
	e.c.registerViews(view)
}

func (e *Exporter) UnregisterView(view *stats.View) {
	// TODO: (@rakyll, @odeke-em) implement UnregisterView.
}

func (c *collector) namespace() string {
	c.mu.RLock()
	ns := c.opts.Namespace
	if ns == "" {
		ns = defaultNamespace
	}
	c.mu.RUnlock()
	return ns
}

func (c *collector) registerViews(views ...*stats.View) {
	if len(views) == 0 {
		return
	}

	namespace := c.namespace()

	newViewCount := 0
	for _, view := range views {
		c.mu.Lock()
		if _, registered := c.registeredViews[view]; !registered {
			desc := prometheus.NewDesc(
				internal.Sanitize(namespace+"_"+view.Name()),
				view.Description(),
				tagKeysToLabels(view.TagKeys()),
				nil,
			)
			c.registeredViews[view] = true
			c.descs = append(c.descs, desc)
			newViewCount += 1
		}
		c.mu.Unlock()
	}

	if newViewCount == 0 {
		return
	}

	c.mu.Lock()
	reg := c.reg
	c.mu.Unlock()

	if ok := reg.Unregister(c); !ok {
		log.Printf("unregister could not unregister: %v", c)
	}
	if err := reg.Register(c); err != nil {
		log.Printf("register err: %v", err)
	}
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

	e.c.addViewData(vd)
}

// ServeHTTP serves the Prometheus endpoint.
func (e *Exporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handler.ServeHTTP(w, r)
}

// collector implements prometheus.Collector
type collector struct {
	opts Options
	mu   sync.RWMutex

	// reg helps us keep a reference to our registrar
	// so that when views are dynamically added, we'll
	// able to register again without having to expose
	// an ugly API for the client in which they'll have
	// to invoke RegisterView for every single view that
	// they create.
	reg *prometheus.Registry

	// views are accumulated and atomically
	// appended to on every Export invocation, from
	// stats. These views are cleared out when
	// Collect is invoked and the cycle is repeated.
	views []*stats.ViewData

	// descs contains the one-time listing of all
	// descriptions that are retrieved after converting
	// each view into a prometheus.Metric.
	//
	// Note: we use slices here because trying to use channels
	// with Prometheus.Collector Describe and Collect methods
	// is quite hairy, moreover for methods that are run once,
	// yet the count of elements to be collected is unknown
	// trying to drain our input channel could potential block forever.
	descs []*prometheus.Desc

	registeredViews map[*stats.View]bool

	// seenMetrics maps from the metric's rawType to the actual Metric.
	// It is an interface to interface mapping
	// but the key is the zero value while the value is the instance.
	seenMetrics map[interface{}]prometheus.Metric
}

var _ prometheus.Collector = (*collector)(nil)

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, desc := range c.descs {
		ch <- desc
	}
}

func (c *collector) lookupMetric(key interface{}) (prometheus.Metric, bool) {
	c.mu.RLock()
	value, ok := c.seenMetrics[key]
	c.mu.RUnlock()
	return value, ok
}

func (c *collector) memoizeMetric(key interface{}, value prometheus.Metric) {
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
	// Get the last views
	views := c.views
	// Now clear them out for the next accumulation
	c.views = c.views[:0]
	c.mu.Unlock()

	if len(views) == 0 {
		return
	}

	seen := make(map[prometheus.Metric]bool)
	for _, vd := range views {
		for _, row := range vd.Rows {
			metric := c.toMetric(vd.View, row)
			if _, ok := seen[metric]; !ok && metric != nil {
				ch <- metric
				seen[metric] = true
			}
		}
	}
}

func (c *collector) toMetric(view *stats.View, row *stats.Row) prometheus.Metric {
	switch aggregation := view.Aggregation().(type) {
	case stats.CountAggregation:
		data := row.Data.(*stats.CountData)
		var key *stats.CountData
		sc, ok := c.lookupMetric(key)
		firstTime := !ok
		if firstTime {
			sc = prometheus.NewCounter(prometheus.CounterOpts{
				Name:      internal.Sanitize(view.Name()),
				Help:      view.Description(),
				Namespace: c.namespace(),
			})
			c.memoizeMetric(key, sc)
		}
		counter := sc.(prometheus.Counter)
		counter.Add(float64(*data))
		return counter

	case stats.DistributionAggregation:
		var key *stats.DistributionData
		hm, ok := c.lookupMetric(key)
		firstTime := !ok
		if firstTime {
			hOpts := prometheus.HistogramOpts{
				Name:        internal.Sanitize(view.Name()),
				Help:        view.Description(),
				Namespace:   c.namespace(),
				ConstLabels: tagsToLabels(row.Tags),
			}
			hm = prometheus.NewHistogram(hOpts)
			c.memoizeMetric(key, hm)
		}
		histogram := hm.(prometheus.Histogram)
		for _, point := range aggregation {
			histogram.Observe(float64(point))
		}
		return histogram

	case stats.SumAggregation:
		panic("stats.SumData:: unimplemented")

	case *stats.MeanAggregation:
		panic("stats.MeanData:: unimplemented")

	default:
		panic("default: Unknown")
	}
}

func tagKeysToLabels(keys []tag.Key) (labels []string) {
	for _, key := range keys {
		labels = append(labels, key.Name())
	}
	return labels
}

func tagsToLabels(tags []tag.Tag) map[string]string {
	m := make(map[string]string)
	for _, tag := range tags {
		m[tag.Key.Name()] = tag.Value
	}
	return m
}

func newCollector(opts Options, registrar *prometheus.Registry) *collector {
	return &collector{
		reg:  registrar,
		opts: opts,

		registeredViews: make(map[*stats.View]bool),
		seenMetrics:     make(map[interface{}]prometheus.Metric),
	}
}

func (c *collector) addViewData(vd *stats.ViewData) {
	c.mu.Lock()
	c.views = append(c.views, vd)
	c.mu.Unlock()
}
