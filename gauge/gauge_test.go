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
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.opencensus.io/metric"
	"go.opencensus.io/resource"
	"go.opencensus.io/tag"
)

func TestGauge(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")
	g := NewDouble("TestGauge", "", "", k1, k2)
	ctx := context.Background()
	g.Set(ctx, 5)
	ctx, _ = tag.New(ctx, tag.Upsert(k1, "k1v1"))
	g.Add(ctx, 1)
	g.Add(ctx, 1)
	ctx, _ = tag.New(ctx, tag.Upsert(k1, "k1v2"))
	ctx, _ = tag.New(ctx, tag.Upsert(k2, "k2v2"))
	g.Add(ctx, 1)
	m := g.Read()
	want := []*metric.Metric{
		{
			Descriptor: &metric.Descriptor{
				Name:      "TestGauge",
				LabelKeys: []string{"k1", "k2"},
			},
			Resource: &resource.Resource{},
			TimeSeries: []*metric.TimeSeries{
				{
					LabelValues: []metric.LabelValue{
						nil, nil,
					},
					Points: []metric.Point{
						metric.NewDoublePoint(time.Time{}, 5),
					},
				},
				{
					LabelValues: []metric.LabelValue{
						metric.NewLabelValue("k1v1"),
						nil,
					},
					Points: []metric.Point{
						metric.NewDoublePoint(time.Time{}, 2),
					},
				},
				{
					LabelValues: []metric.LabelValue{
						metric.NewLabelValue("k1v2"),
						metric.NewLabelValue("k2v2"),
					},
					Points: []metric.Point{
						metric.NewDoublePoint(time.Time{}, 1),
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

func ignoreTimes(_, _ time.Time) bool {
	return true
}

func canonicalize(ms []*metric.Metric) {
	for _, m := range ms {
		sort.Slice(m.TimeSeries, func(i, j int) bool {
			// sort time series by their label values
			iLabels := m.TimeSeries[i].LabelValues
			jLabels := m.TimeSeries[j].LabelValues
			for k := 0; k < len(iLabels); k++ {
				if iLabels[k] == nil {
					if jLabels[k] != nil {
						return true
					}
				} else if jLabels[k] == nil {
					return false
				} else if *iLabels[k] != *jLabels[k] {
					return *iLabels[k] < *jLabels[k]
				}
			}
			panic("should have returned")
		})
	}
}
