package prometheus

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/metric"
	"go.opencensus.io/resource"
)

// MetricCollector is a Prometheus Collector that collects from an OpenCensus
// metric provider.
type MetricCollector struct {
	// GetName may be provided to customize how the metric string is converted
	// to a (namespaced) metric name for Prometheus.
	GetName func(m *metric.Metric) string
	// GetResourceLabels may be provided to customize the Prometheus labels
	// for the resource associated with the metric.
	// By default, just the labels of the resource are added.
	GetResourceLabels func(*resource.Resource, map[string]string)
	// OnError may be provided as a callback when an error occurs converting
	// OpenCensus metrics to Prometheus metrics.
	OnError func(err error)
	// MetricProvider is the provider from which metrics will be read.
	// By default, it is the default registry.
	MetricProvider metric.Producer
	// ConstLabels are constant labels added to each metric.
	ConstLabels prometheus.Labels
	// NilLabelValue is a special string that will be used to represent nil
	// label values. By default, it is just the empty string.
	NilLabelValue string
}

var _ prometheus.Collector = (*MetricCollector)(nil)

// NewMetricCollector creates a new MetricCollector initialized to collect
// from the default registry.
func NewMetricCollector() *MetricCollector {
	mc := &MetricCollector{
		MetricProvider: metric.DefaultRegistry(),
		GetResourceLabels: func(r *resource.Resource, labels map[string]string) {
			if r == nil {
				return
			}
			for k, v := range r.Labels {
				labels[k] = v
			}
		},
		ConstLabels: nil,
		GetName: func(m *metric.Metric) string {
			return viewName("", m.Descriptor.Name)
		},
		OnError: func(err error) {
			log.Printf("Failed to export to Prometheus: %v", err)
		},
	}
	return mc
}

func (mc *MetricCollector) Describe(chan<- *prometheus.Desc) {
	// Not supported.
}

func (mc *MetricCollector) Collect(out chan<- prometheus.Metric) {
	for _, m := range mc.MetricProvider.Read() {
		for _, ts := range m.TimeSeries {
			labels := make(map[string]string)
			mc.GetResourceLabels(m.Resource, labels)
			for i, lk := range m.Descriptor.LabelKeys {
				lv := ts.LabelValues[i]
				if lv == nil {
					labels[lk] = mc.NilLabelValue
				} else {
					labels[lk] = *lv
				}
			}
			labelKeys := make([]string, 0, len(labels))
			labelVals := make([]string, 0, len(labels))
			for k, v := range labels {
				labelKeys = append(labelKeys, k)
				labelVals = append(labelVals, v)
			}
			desc := prometheus.NewDesc(
				mc.GetName(m), m.Descriptor.Description, labelKeys, mc.ConstLabels)
			for _, p := range ts.Points {
				converter := &metricConverter{
					metric:    m,
					desc:      desc,
					labelVals: labelVals,
				}
				p.ReadValue(converter)
				if converter.outErr != nil {
					mc.OnError(converter.outErr)
				} else {
					out <- converter.outMetric
				}
			}
		}
	}
}

type metricConverter struct {
	metric    *metric.Metric
	desc      *prometheus.Desc
	labelVals []string
	outMetric prometheus.Metric
	outErr    error
}

func (mc *metricConverter) VisitDoubleValue(v float64) {
	var valType prometheus.ValueType
	if mc.metric.Descriptor.Type.IsGuage() {
		valType = prometheus.GaugeValue
	} else {
		valType = prometheus.UntypedValue
	}
	mc.outMetric, mc.outErr = prometheus.NewConstMetric(mc.desc, valType, v, mc.labelVals...)
}

func (mc *metricConverter) VisitDistributionValue(d *metric.Distribution) {
	buckets := map[float64]uint64{}
	cumulative := uint64(0)
	for i, b := range d.BucketOptions.ExplicitBoundaries {
		cumulative += uint64(d.Buckets[i].Count)
		buckets[b] = cumulative
	}
	mc.outMetric, mc.outErr = prometheus.NewConstHistogram(mc.desc, uint64(d.Count), d.Sum, buckets, mc.labelVals...)
}

func (mc *metricConverter) VisitInt64Value(v int64) {
	var valType prometheus.ValueType
	if mc.metric.Descriptor.Type.IsGuage() {
		valType = prometheus.GaugeValue
	} else {
		valType = prometheus.CounterValue
	}
	mc.outMetric, mc.outErr = prometheus.NewConstMetric(mc.desc, valType, float64(v), mc.labelVals...)
}

func (mc *metricConverter) VisitSummaryValue(v *metric.Summary) {
	quantiles := map[float64]float64{}
	for p, v := range v.Snapshot.Percentiles {
		quantiles[p/100.0] = v
	}
	prometheus.NewConstSummary(mc.desc, uint64(v.Snapshot.Count), v.Snapshot.Sum, quantiles, mc.labelVals...)
}
