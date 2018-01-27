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

package prometheus

import (
	"testing"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"

	"github.com/prometheus/client_golang/prometheus"
)

func newView(agg view.Aggregation, window view.Window) *view.View {
	m, _ := stats.NewMeasureInt64("tests/foo1", "bytes", "byte")
	view, _ := view.New("foo", "bar", nil, m, agg, window)
	return view
}

func TestOnlyCumulativeWindowSupported(t *testing.T) {
	// See Issue https://github.com/census-instrumentation/opencensus-go/issues/214.
	count1 := view.CountData(1)
	mean1 := view.MeanData{
		Mean:  4.5,
		Count: 5,
	}
	tests := []struct {
		vds  *view.Data
		want int
	}{
		0: {
			vds: &view.Data{
				View: newView(view.CountAggregation{}, view.Cumulative{}),
			},
			want: 0, // no rows present
		},
		1: {
			vds: &view.Data{
				View: newView(view.CountAggregation{}, view.Cumulative{}),
				Rows: []*view.Row{
					{nil, &count1},
				},
			},
			want: 1,
		},
		2: {
			vds: &view.Data{
				View: newView(view.CountAggregation{}, view.Interval{}),
				Rows: []*view.Row{
					{nil, &count1},
				},
			},
			want: 0,
		},
		3: {
			vds: &view.Data{
				View: newView(view.MeanAggregation{}, view.Cumulative{}),
				Rows: []*view.Row{
					{nil, &mean1},
				},
			},
			want: 1,
		},
	}

	for i, tt := range tests {
		reg := prometheus.NewRegistry()
		collector := newCollector(Options{}, reg)
		collector.addViewData(tt.vds)
		mm, err := reg.Gather()
		if err != nil {
			t.Errorf("#%d: Gather err: %v", i, err)
		}
		reg.Unregister(collector)
		if got, want := len(mm), tt.want; got != want {
			t.Errorf("#%d: got nil %v want nil %v", i, got, want)
		}
	}
}

func TestSingletonExporter(t *testing.T) {
	exp, err := NewExporter(Options{})
	if err != nil {
		t.Fatalf("NewExporter() = %v", err)
	}
	if exp == nil {
		t.Fatal("Nil exporter")
	}

	// Should all now fail
	exp, err = NewExporter(Options{})
	if err == nil {
		t.Fatal("NewExporter() = nil")
	}
	if exp != nil {
		t.Fatal("Non-nil exporter")
	}
}
