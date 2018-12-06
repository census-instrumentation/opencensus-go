// Copyright 2018, OpenCensus Authors
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

package view

import (
	"go.opencensus.io/metric/metricexport"
	"time"

	"go.opencensus.io/metric"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

func (w *worker) Read() (ms []*metricexport.Metric) {
	w.runSync(func() {
		now := time.Now()
		ms = make([]*metricexport.Metric, 0, len(w.views))
		for _, v := range w.views {
			if !v.isSubscribed() {
				continue
			}
			rows := v.collectedRows()
			_, ok := w.startTimes[v]
			if !ok {
				w.startTimes[v] = now
			}
			m := &metricexport.Metric{
				Descriptor: metricDesc(v.view),
			}
			for _, row := range rows {
				var labelVals []metric.LabelValue
				for _, k := range m.Descriptor.LabelKeys {
					lv := metric.NewLabelValue(lookupTagByKeyName(row.Tags, k))
					labelVals = append(labelVals, lv)
				}
				_, isFloat := v.view.Measure.(*stats.Float64Measure)
				ts := &metricexport.TimeSeries{
					StartTime:   w.startTimes[v],
					LabelValues: labelVals,
					Points: []metricexport.Point{
						row.Data.exportAsPoint(now, isFloat),
					},
				}
				m.TimeSeries = append(m.TimeSeries, ts)
			}
			ms = append(ms, m)
		}
	})
	return ms
}

func metricDesc(v *View) metricexport.Descriptor {
	var labelKeys []string
	for _, tk := range v.TagKeys {
		labelKeys = append(labelKeys, tk.Name())
	}
	return metricexport.Descriptor{
		Name:        v.Name,
		Description: v.Description,
		Unit:        metric.Unit(v.Measure.Unit()),
		Type:        metricType(v.Aggregation.Type, v.Measure),
		LabelKeys:   labelKeys,
	}
}

func metricType(aggType AggType, measure stats.Measure) metricexport.Type {
	switch aggType {
	case AggTypeCount:
		return metricexport.TypeCumulativeInt64
	case AggTypeDistribution:
		return metricexport.TypeCumulativeDistribution
	case AggTypeLastValue, AggTypeSum:
		if _, ok := measure.(*stats.Float64Measure); ok {
			return metricexport.TypeCumulativeFloat64
		} else {
			return metricexport.TypeCumulativeInt64
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
