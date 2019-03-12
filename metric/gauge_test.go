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
	"fmt"
	"go.opencensus.io/metric/metricdata"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestGauge(t *testing.T) {
	r := NewRegistry()
	f, _ := r.AddFloat64Gauge("TestGauge", "", "", "k1", "k2")
	e, _ := f.GetEntry(metricdata.LabelValue{}, metricdata.LabelValue{})
	e.Set(5)
	e, _ = f.GetEntry(metricdata.NewLabelValue("k1v1"), metricdata.LabelValue{})
	e.Add(1)
	e, _ = f.GetEntry(metricdata.NewLabelValue("k1v1"), metricdata.LabelValue{})
	e.Add(1)
	e, _ = f.GetEntry(metricdata.NewLabelValue("k1v2"), metricdata.NewLabelValue("k2v2"))
	e.Add(1)
	m := r.Read()
	want := []*metricdata.Metric{
		{
			Descriptor: metricdata.Descriptor{
				Name:      "TestGauge",
				LabelKeys: []string{"k1", "k2"},
			},
			TimeSeries: []*metricdata.TimeSeries{
				{
					LabelValues: []metricdata.LabelValue{
						{}, {},
					},
					Points: []metricdata.Point{
						metricdata.NewFloat64Point(time.Time{}, 5),
					},
				},
				{
					LabelValues: []metricdata.LabelValue{
						metricdata.NewLabelValue("k1v1"),
						{},
					},
					Points: []metricdata.Point{
						metricdata.NewFloat64Point(time.Time{}, 2),
					},
				},
				{
					LabelValues: []metricdata.LabelValue{
						metricdata.NewLabelValue("k1v2"),
						metricdata.NewLabelValue("k2v2"),
					},
					Points: []metricdata.Point{
						metricdata.NewFloat64Point(time.Time{}, 1),
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

func TestFloat64Entry_Add(t *testing.T) {
	r := NewRegistry()
	g, _ := r.AddFloat64Gauge("g", "", metricdata.UnitDimensionless)
	e, _ := g.GetEntry()
	e.Add(0)
	ms := r.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(float64), 0.0; got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
	e, _ = g.GetEntry()
	e.Add(1)
	ms = r.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(float64), 1.0; got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
	e, _ = g.GetEntry()
	e.Add(-1)
	ms = r.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(float64), 0.0; got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
}

func TestFloat64Gauge_Add_NegativeTotals(t *testing.T) {
	r := NewRegistry()
	g, _ := r.AddFloat64Gauge("g", "", metricdata.UnitDimensionless)
	e, _ := g.GetEntry()
	e.Add(-1.0)
	ms := r.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(float64), float64(0); got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
}

func TestInt64GaugeEntry_Add(t *testing.T) {
	r := NewRegistry()
	g, _ := r.AddInt64Gauge("g", "", metricdata.UnitDimensionless)
	e, _ := g.GetEntry()
	e.Add(0)
	ms := r.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(int64), int64(0); got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
	e, _ = g.GetEntry()
	e.Add(1)
	ms = r.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(int64), int64(1); got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
}

func TestInt64Gauge_Add_NegativeTotals(t *testing.T) {
	r := NewRegistry()
	g, _ := r.AddInt64Gauge("g", "", metricdata.UnitDimensionless)
	e, _ := g.GetEntry()
	e.Add(-1)
	ms := r.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(int64), int64(0); got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
}

func TestGaugeWithSameNameDiffType(t *testing.T) {
	r := NewRegistry()
	r.AddInt64Gauge("g", "", metricdata.UnitDimensionless)
	_, wantErr := r.AddFloat64Gauge("g", "", metricdata.UnitDimensionless)
	if wantErr == nil {
		t.Errorf("got: nil, want error: %s", errGaugeExistsWithDiffType)
	}
}

func TestGaugeWithLabelMismatch(t *testing.T) {
	r := NewRegistry()
	g, _ := r.AddInt64Gauge("g", "", metricdata.UnitDimensionless, "k1")
	_, wantErr := g.GetEntry(metricdata.NewLabelValue("k1v2"), metricdata.NewLabelValue("k2v2"))
	if wantErr == nil {
		t.Errorf("got: nil, want error: %s", errKeyValueMismatch)
	}
}

func TestMapKey(t *testing.T) {
	cases := [][]metricdata.LabelValue{
		{},
		{metricdata.LabelValue{}},
		{metricdata.NewLabelValue("")},
		{metricdata.NewLabelValue("-")},
		{metricdata.NewLabelValue(",")},
		{metricdata.NewLabelValue("v1"), metricdata.NewLabelValue("v2")},
		{metricdata.NewLabelValue("v1"), metricdata.LabelValue{}},
		{metricdata.NewLabelValue("v1"), metricdata.LabelValue{}, metricdata.NewLabelValue(string([]byte{0}))},
		{metricdata.LabelValue{}, metricdata.LabelValue{}},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			g := &gauge{
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

func TestRaceCondition(t *testing.T) {
	r := NewRegistry()

	// start reader before adding Gauge metric.
	var ms = []*metricdata.Metric{}
	for i := 0; i < 5; i++ {
		go func(k int) {
			for j := 0; j < 5; j++ {
				g, _ := r.AddInt64Gauge(fmt.Sprintf("g%d%d", k, j), "", metricdata.UnitDimensionless)
				e, _ := g.GetEntry()
				e.Add(1)
			}
		}(i)
	}
	time.Sleep(1 * time.Second)
	ms = r.Read()
	if got, want := ms[0].TimeSeries[0].Points[0].Value.(int64), int64(1); got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
}

func ignoreTimes(_, _ time.Time) bool {
	return true
}

func canonicalize(ms []*metricdata.Metric) {
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
