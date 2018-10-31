package view

import (
	"time"

	"go.opencensus.io/metric"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

func (w *worker) Read() (ms []*metric.Metric) {
	w.runSync(func() {
		now := time.Now()
		ms = make([]*metric.Metric, 0, len(w.views))
		for _, v := range w.views {
			if !v.isSubscribed() {
				continue
			}
			rows := v.collectedRows()
			_, ok := w.startTimes[v]
			if !ok {
				w.startTimes[v] = now
			}
			m := &metric.Metric{
				Descriptor: metricDesc(v.view),
			}
			for _, row := range rows {
				var labelVals []metric.LabelValue
				for _, k := range m.Descriptor.LabelKeys {
					lv := metric.NewLabelValue(lookupTagByKeyName(row.Tags, k))
					labelVals = append(labelVals, lv)
				}
				ts := &metric.TimeSeries{
					StartTime:   w.startTimes[v],
					LabelValues: labelVals,
					Points: []metric.Point{
						row.Data.toMetricPoint(now, m.Descriptor.Type.ValueType()),
					},
				}
				m.TimeSeries = append(m.TimeSeries, ts)
			}
			ms = append(ms, m)
		}
	})
	return ms
}

func metricDesc(v *View) *metric.Descriptor {
	var labelKeys []string
	for _, tk := range v.TagKeys {
		labelKeys = append(labelKeys, tk.Name())
	}
	return &metric.Descriptor{
		Name:        v.Name,
		Description: v.Description,
		Unit:        metric.Unit(v.Measure.Unit()),
		Type:        metricType(v.Aggregation.Type, v.Measure),
		LabelKeys:   labelKeys,
	}
}

func metricType(aggType AggType, measure stats.Measure) metric.Type {
	switch aggType {
	case AggTypeCount:
		return metric.TypeCumulativeInt64
	case AggTypeDistribution:
		return metric.TypeCumulativeDistribution
	case AggTypeLastValue, AggTypeSum:
		if _, ok := measure.(*stats.Float64Measure); ok {
			return metric.TypeCumulativeDouble
		} else {
			return metric.TypeCumulativeInt64
		}
	default:
		panic("unable to map to metric type")
	}
}

func lookupTagByKeyName(ts []tag.Tag, name string) string {
	for _, t := range ts {
		if t.Key.Name() == name {
			return t.Value
		}
	}
	return ""
}
