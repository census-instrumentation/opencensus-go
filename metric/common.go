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

package metric

import (
	"sync"
	"time"

	"go.opencensus.io/internal/tagencoding"
	"go.opencensus.io/metric/metricdata"
)

// baseMetric is common representation for gauge and cumulative metrics.
//
// baseMetric maintains a value for each combination of of label values passed to
// Set, Add, or Inc method.
//
// baseMetric should not be used directly, use metric specific type such as
// Float64Gauge or Int64Gauge.
type baseMetric struct {
	vals   sync.Map
	desc   metricdata.Descriptor
	start  time.Time
	keys   []string
	bmType baseMetricType
}

type baseMetricType int

const (
	gaugeInt64 baseMetricType = iota
	gaugeFloat64
	derivedGaugeInt64
	derivedGaugeFloat64
	cumulativeInt64
	cumulativeFloat64
	derivedCumulativeInt64
	derivedCumulativeFloat64
)

type baseEntry interface {
	read(t time.Time) metricdata.Point
}

// Read returns the current values of the baseMetric as a metric for export.
func (bm *baseMetric) read() *metricdata.Metric {
	now := time.Now()
	m := &metricdata.Metric{
		Descriptor: bm.desc,
	}
	bm.vals.Range(func(k, v interface{}) bool {
		entry := v.(baseEntry)
		key := k.(string)
		labelVals := bm.labelValues(key)
		m.TimeSeries = append(m.TimeSeries, &metricdata.TimeSeries{
			StartTime:   now, // Gauge value is instantaneous.
			LabelValues: labelVals,
			Points: []metricdata.Point{
				entry.read(now),
			},
		})
		return true
	})
	return m
}

func (bm *baseMetric) mapKey(labelVals []metricdata.LabelValue) string {
	vb := &tagencoding.Values{}
	for _, v := range labelVals {
		b := make([]byte, 1, len(v.Value)+1)
		if v.Present {
			b[0] = 1
			b = append(b, []byte(v.Value)...)
		}
		vb.WriteValue(b)
	}
	return string(vb.Bytes())
}

func (bm *baseMetric) labelValues(s string) []metricdata.LabelValue {
	vals := make([]metricdata.LabelValue, 0, len(bm.keys))
	vb := &tagencoding.Values{Buffer: []byte(s)}
	for range bm.keys {
		v := vb.ReadValue()
		if v[0] == 0 {
			vals = append(vals, metricdata.LabelValue{})
		} else {
			vals = append(vals, metricdata.NewLabelValue(string(v[1:])))
		}
	}
	return vals
}

func (bm *baseMetric) entryForValues(labelVals []metricdata.LabelValue, newEntry func() baseEntry) (interface{}, error) {
	if len(labelVals) != len(bm.keys) {
		return nil, errKeyValueMismatch
	}
	mapKey := bm.mapKey(labelVals)
	if entry, ok := bm.vals.Load(mapKey); ok {
		return entry, nil
	}
	entry, _ := bm.vals.LoadOrStore(mapKey, newEntry())
	return entry, nil
}

func (bm *baseMetric) upsertEntry(labelVals []metricdata.LabelValue, newEntry func() baseEntry) error {
	if len(labelVals) != len(bm.keys) {
		return errKeyValueMismatch
	}
	mapKey := bm.mapKey(labelVals)
	bm.vals.Delete(mapKey)
	bm.vals.Store(mapKey, newEntry())
	return nil
}
