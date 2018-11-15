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

package gauge

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"go.opencensus.io/metric/metricexport"

	"github.com/google/go-cmp/cmp"
	"go.opencensus.io/metric"
)

func TestGauge(t *testing.T) {
	g := NewFloat64("TestGauge", "", "", "k1", "k2")
	g.Set(5, metric.LabelValue{}, metric.LabelValue{})
	g.Add(1, metric.NewLabelValue("k1v1"), metric.LabelValue{})
	g.Add(1, metric.NewLabelValue("k1v1"), metric.LabelValue{})
	g.Add(1, metric.NewLabelValue("k1v2"), metric.NewLabelValue("k2v2"))
	m := g.Read()
	want := []*metricexport.Metric{
		{
			Descriptor: metricexport.Descriptor{
				Name:      "TestGauge",
				LabelKeys: []string{"k1", "k2"},
			},
			TimeSeries: []*metricexport.TimeSeries{
				{
					LabelValues: []metric.LabelValue{
						{}, {},
					},
					Points: []metricexport.Point{
						metricexport.NewFloat64Point(time.Time{}, 5),
					},
				},
				{
					LabelValues: []metric.LabelValue{
						metric.NewLabelValue("k1v1"),
						{},
					},
					Points: []metricexport.Point{
						metricexport.NewFloat64Point(time.Time{}, 2),
					},
				},
				{
					LabelValues: []metric.LabelValue{
						metric.NewLabelValue("k1v2"),
						metric.NewLabelValue("k2v2"),
					},
					Points: []metricexport.Point{
						metricexport.NewFloat64Point(time.Time{}, 1),
					},
				},
			},
		},
	}
	canonicalize(m)
	canonicalize(want)
	if diff := cmp.Diff(m, want, cmp.Comparer(ignoreTimes)); diff != "" {
		t.Errorf("-got +want: %s", diff)
	}
}

func TestFloat64_Add(t *testing.T) {
	g := NewFloat64("g", "", metric.UnitDimensionless)
	g.Add(0)
	ms := g.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(float64), 0.0; got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
	g.Add(1)
	ms = g.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(float64), 1.0; got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
	g.Add(-2)
	ms = g.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(float64), -1.0; got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
}

func TestInt64_Add(t *testing.T) {
	g := NewInt64("g", "", metric.UnitDimensionless)
	g.Add(0)
	ms := g.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(int64), int64(0); got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
	g.Add(1)
	ms = g.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(int64), int64(1); got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
}

func TestMapKey(t *testing.T) {
	cases := [][]metric.LabelValue{
		{},
		{metric.LabelValue{}},
		{metric.NewLabelValue("")},
		{metric.NewLabelValue("-")},
		{metric.NewLabelValue(",")},
		{metric.NewLabelValue("v1"), metric.NewLabelValue("v2")},
		{metric.NewLabelValue("v1"), metric.LabelValue{}},
		{metric.NewLabelValue("v1"), metric.LabelValue{}, metric.NewLabelValue(string([]byte{0}))},
		{metric.LabelValue{}, metric.LabelValue{}},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			g := &Gauge{
				keys: make([]string, len(tc)),
			}
			mk := g.mapKey(tc)
			vals := g.labelValues(mk)
			if diff := cmp.Diff(vals, tc); diff != "" {
				t.Errorf("values differ after serialization -got +want: %s", diff)
			}
		})
	}
}

func ignoreTimes(_, _ time.Time) bool {
	return true
}

func canonicalize(ms []*metricexport.Metric) {
	for _, m := range ms {
		sort.Slice(m.TimeSeries, func(i, j int) bool {
			// sort time series by their label values
			iLabels := m.TimeSeries[i].LabelValues
			jLabels := m.TimeSeries[j].LabelValues
			for k := 0; k < len(iLabels); k++ {
				if !iLabels[k].Present {
					if jLabels[k].Present {
						return true
					}
				} else if !jLabels[k].Present {
					return false
				} else {
					return iLabels[k].Value < jLabels[k].Value
				}
			}
			panic("should have returned")
		})
	}
}
