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
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"

	"go.opencensus.io/internal"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opencensus.io/metric/metricexport"
	"go.opencensus.io/metric/metricdata"
	"context"
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
	collector := newCollector(o, o.Registry)
	e := &Exporter{
		opts:    o,
		g:       o.Registry,
		c:       collector,
		handler: promhttp.HandlerFor(o.Registry, promhttp.HandlerOpts{}),
	}
	collector.ensureRegisteredOnce()

	return e, nil
}

var _ http.Handler = (*Exporter)(nil)
var _ view.Exporter = (*Exporter)(nil)

func (c *collector) registerViews(views ...*view.View) {
	count := 0
	for _, view := range views {
		sig := viewSignature(c.opts.Namespace, view)
		c.registeredViewsMu.Lock()
		_, ok := c.registeredViews[sig]
		c.registeredViewsMu.Unlock()

		if !ok {
			desc := prometheus.NewDesc(
				viewName(c.opts.Namespace, view),
				view.Description,
				tagKeysToLabels(view.TagKeys),
				c.opts.ConstLabels,
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

	c.ensureRegisteredOnce()
}

// ensureRegisteredOnce invokes reg.Register on the collector itself
// exactly once to ensure that we don't get errors such as
//  cannot register the collector: descriptor Desc{fqName: *}
//  already exists with the same fully-qualified name and const label values
// which is documented by Prometheus at
//  https://github.com/prometheus/client_golang/blob/fcc130e101e76c5d303513d0e28f4b6d732845c7/prometheus/registry.go#L89-L101
func (c *collector) ensureRegisteredOnce() {
	c.registerOnce.Do(func() {
		if err := c.reg.Register(c); err != nil {
			c.opts.onError(fmt.Errorf("cannot register the collector: %v", err))
		}
	})

}

func (o *Options) onError(err error) {
	if o.OnError != nil {
		o.OnError(err)
	} else {
		log.Printf("Failed to export to Prometheus: %v", err)
	}
}

// ExportView exports to the Prometheus if view data has one or more rows.
// Each OpenCensus AggregationData will be converted to
// corresponding Prometheus Metric: SumData will be converted
// to Untyped Metric, CountData will be a Counter Metric,
// DistributionData will be a Histogram Metric.
func (e *Exporter) ExportView(vd *view.Data) {
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

	registerOnce sync.Once

	// reg helps collector register views dynamically.
	reg *prometheus.Registry

	// viewData are accumulated and atomically
	// appended to on every Export invocation, from
	// stats. These views are cleared out when
	// Collect is invoked and the cycle is repeated.
	viewData map[string]*view.Data

	registeredViewsMu sync.Mutex
	// registeredViews maps a view to a prometheus desc.
	registeredViews map[string]*prometheus.Desc

	// reader reads metrics from all registered producers.
	reader  *metricexport.Reader
}

func (c *collector) addViewData(vd *view.Data) {
	c.registerViews(vd.View)
	sig := viewSignature(c.opts.Namespace, vd.View)

	c.mu.Lock()
	c.viewData[sig] = vd
	c.mu.Unlock()
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	c.readDesc(ch)
}

// Collect fetches the statistics from OpenCensus
// and delivers them as Prometheus Metrics.
// Collect is invoked everytime a prometheus.Gatherer is run
// for example when the HTTP endpoint is invoked by Prometheus.
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.readMetrics(ch)
}

func (c *collector) toMetric(desc *prometheus.Desc, v *view.View, row *view.Row) (prometheus.Metric, error) {
	switch data := row.Data.(type) {
	case *view.CountData:
		return prometheus.NewConstMetric(desc, prometheus.CounterValue, float64(data.Value), tagValues(row.Tags, v.TagKeys)...)

	case *view.DistributionData:
		points := make(map[float64]uint64)
		// Histograms are cumulative in Prometheus.
		// Get cumulative bucket counts.
		cumCount := uint64(0)
		for i, b := range v.Aggregation.Buckets {
			cumCount += uint64(data.CountPerBucket[i])
			points[b] = cumCount
		}
		return prometheus.NewConstHistogram(desc, uint64(data.Count), data.Sum(), points, tagValues(row.Tags, v.TagKeys)...)

	case *view.SumData:
		return prometheus.NewConstMetric(desc, prometheus.UntypedValue, data.Value, tagValues(row.Tags, v.TagKeys)...)

	case *view.LastValueData:
		return prometheus.NewConstMetric(desc, prometheus.GaugeValue, data.Value, tagValues(row.Tags, v.TagKeys)...)

	default:
		return nil, fmt.Errorf("aggregation %T is not yet supported", v.Aggregation)
	}
}

func tagKeysToLabels(keys []tag.Key) (labels []string) {
	for _, key := range keys {
		labels = append(labels, internal.Sanitize(key.Name()))
	}
	return labels
}

func newCollector(opts Options, registrar *prometheus.Registry) *collector {
	return &collector{
		reg:             registrar,
		opts:            opts,
		registeredViews: make(map[string]*prometheus.Desc),
		viewData:        make(map[string]*view.Data),
		reader: metricexport.NewReader()}
}

func tagValues(t []tag.Tag, expectedKeys []tag.Key) []string {
	var values []string
	// Add empty string for all missing keys in the tags map.
	idx := 0
	for _, t := range t {
		for t.Key != expectedKeys[idx] {
			idx++
			values = append(values, "")
		}
		values = append(values, t.Value)
		idx++
	}
	for idx < len(expectedKeys) {
		idx++
		values = append(values, "")
	}
	return values
}

func viewName(namespace string, v *view.View) string {
	var name string
	if namespace != "" {
		name = namespace + "_"
	}
	return name + internal.Sanitize(v.Name)
}

func viewSignature(namespace string, v *view.View) string {
	var buf bytes.Buffer
	buf.WriteString(viewName(namespace, v))
	for _, k := range v.TagKeys {
		buf.WriteString("-" + k.Name())
	}
	return buf.String()
}

func (c *collector) cloneViewData() map[string]*view.Data {
	c.mu.Lock()
	defer c.mu.Unlock()

	viewDataCopy := make(map[string]*view.Data)
	for sig, viewData := range c.viewData {
		viewDataCopy[sig] = viewData
	}
	return viewDataCopy
}

func metricLabelsToPromLabels(ls []string) (labels []string) {
	for _, l := range ls {
		labels = append(labels, internal.Sanitize(l))
	}
	return labels
}


func (c *collector) metricToDesc(metric *metricdata.Metric) *prometheus.Desc {
	return prometheus.NewDesc(
		metricName(c.opts.Namespace, metric),
		metric.Descriptor.Description,
		metricLabelsToPromLabels(metric.Descriptor.LabelKeys),
		c.opts.ConstLabels)
}

func (c *collector)readMetrics(ch chan<- prometheus.Metric) {
	me := &metricExporter{c: c, metricCh: ch }
	c.reader.ReadAndExport(me)
}

func (c *collector)readDesc(ch chan<- *prometheus.Desc) {
	de := &descExporter{c: c, descCh: ch }
	c.reader.ReadAndExport(de)
}

//func (c *collector) registerMetricDesc(metric *metricdata.Metric) *prometheus.Desc {
//	sig := metricSignature(c.opts.Namespace, metric)
//	c.registeredViewsMu.Lock()
//	desc, ok := c.registeredViews[sig]
//	c.registeredViewsMu.Unlock()
//
//	if !ok {
//		desc = c.metricToDesc(metric)
//		c.registeredViewsMu.Lock()
//		c.registeredViews[sig] = desc
//		c.registeredViewsMu.Unlock()
//	}
//	return desc
//}
//

type metricExporter struct {
	c *collector
	metricCh chan<- prometheus.Metric
}

// ExportMetrics exports to the Prometheus.
// Each OpenCensus Metric will be converted to
// corresponding Prometheus Metric:
// SumData will be converted to Untyped Metric,
// CountData will be a Counter Metric,
// DistributionData will be a Histogram Metric.
func (me *metricExporter) ExportMetrics(ctx context.Context, metrics []*metricdata.Metric) error {
	for _, metric := range metrics {
		desc := me.c.metricToDesc(metric)
		for _, ts := range metric.TimeSeries {
			tvs := metricTagValues(ts.LabelValues)
			for _, point := range ts.Points {
				metric, err := me.fromOcMetricToPromMetric(desc, metric, point, tvs)
				if err != nil {
					me.c.opts.onError(err)
				} else {
					me.metricCh <- metric
				}
			}
		}
	}
	return nil
}

//func metricSignature(namespace string, m *metricdata.Metric) string {
//	var buf bytes.Buffer
//	buf.WriteString(metricName(namespace, m))
//	for _, k := range m.Descriptor.LabelKeys {
//		buf.WriteString("-" + k)
//	}
//	return buf.String()
//}
//
func metricName(namespace string, m *metricdata.Metric) string {
	var name string
	if namespace != "" {
		name = namespace + "_"
	}
	return name + internal.Sanitize(m.Descriptor.Name)
}

func (me *metricExporter) fromOcMetricToPromMetric(
	desc *prometheus.Desc,
	metric *metricdata.Metric,
	point metricdata.Point,
	tvs []string) (prometheus.Metric, error) {
	switch (metric.Descriptor.Type) {
	case metricdata.TypeCumulativeFloat64, metricdata.TypeCumulativeInt64:
		pv, err := pointToPromValue(point)
		if err != nil {
			return nil, err
		}
		return prometheus.NewConstMetric(desc, prometheus.CounterValue, pv, tvs...)

	case metricdata.TypeGaugeFloat64, metricdata.TypeGaugeInt64:
		pv, err := pointToPromValue(point)
		if err != nil {
			return nil, err
		}
		return prometheus.NewConstMetric(desc, prometheus.GaugeValue, pv, tvs...)

	case metricdata.TypeCumulativeDistribution:
		switch v := point.Value.(type) {
		case *metricdata.Distribution:
			points := make(map[float64]uint64)
			// Histograms are cumulative in Prometheus.
			// Get cumulative bucket counts.
			cumCount := uint64(0)
			for i, b := range v.BucketOptions.Bounds {
				cumCount += uint64(v.Buckets[i].Count)
				points[b] = cumCount
			}
			return prometheus.NewConstHistogram(desc, uint64(v.Count), v.Sum, points, tvs...)
		default:
			return nil, pointTypeError(point)
		}

	default:
		return nil, fmt.Errorf("aggregation %T is not yet supported", metric.Descriptor.Type)
	}
}

func metricTagValues(lvs []metricdata.LabelValue) []string {
	var values []string
	for _, lv := range lvs {
		if lv.Present {
			values = append(values, lv.Value)
		} else {
			values = append(values, "")
		}
	}
	return values
}

func pointTypeError(point metricdata.Point) error {
	return fmt.Errorf("point type %T is not yet supported", point)

}

func pointToPromValue(point metricdata.Point) (float64, error) {
	switch v := point.Value.(type) {
	case float64:
		return v, nil
	case int64:
		return float64(v), nil
	default:
		return 0.0, pointTypeError(point)
	}
}

type descExporter struct {
	c *collector
	descCh chan<- *prometheus.Desc
}

// ExportMetrics exports descriptor to the Prometheus.
// It is invoked when request to scrape descriptors is received.
func (me *descExporter) ExportMetrics(ctx context.Context, metrics []*metricdata.Metric) error {
	for _, metric := range metrics {
		desc := me.c.metricToDesc(metric)
		me.descCh <- desc
	}
	return nil
}
