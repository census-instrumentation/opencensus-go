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
//

package view

import (
	"context"
	"testing"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/exporter"
	"go.opencensus.io/tag"
)

func Test_View_MeasureFloat64_AggregationDistribution_Simple(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")
	k3, _ := tag.NewKey("k3")
	m := stats.Int64("Test_View_MeasureFloat64_AggregationDistribution/m1", "", stats.UnitNone)
	view, err := newViewInternal(&View{
		TagKeys:     []tag.Key{k1, k2},
		Measure:     m,
		Aggregation: Distribution(2),
	})
	if err != nil {
		t.Fatal(err)
	}

	type tagString struct {
		k tag.Key
		v string
	}
	type record struct {
		f    float64
		tags []tagString
	}

	records := []record{
		{f: 1, tags: []tagString{{k1, "v1 is a very long value key"}}},
		{f: 5, tags: []tagString{{k1, "v1 is a very long value key"}, {k3, "v3"}}},
		{f: 1, tags: []tagString{{k1, "v1 is another very long value key"}}},
		{f: 1, tags: []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
		{f: 5, tags: []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
		{f: 3, tags: []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
		{f: 3, tags: []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
	}
	wantRows := []*exporter.Row{
		{
			[]tag.Tag{{Key: k1, Value: "v1 is a very long value key"}},
			exporter.AggregationData{
				Count: 2, Min: 1, Max: 5, Mean: 3, SumOfSquaredDev: 8, CountPerBucket: []int64{1, 1}, Bounds: []float64{2},
			},
		},
		{
			[]tag.Tag{{Key: k1, Value: "v1 is another very long value key"}},
			exporter.AggregationData{
				Count: 1, Min: 1, Max: 1, Mean: 1, CountPerBucket: []int64{1, 0}, Bounds: []float64{2},
			},
		},
		{
			[]tag.Tag{{Key: k1, Value: "v1 is a very long value key"}, {Key: k2, Value: "v2 is a very long value key"}},
			exporter.AggregationData{
				Count: 4, Min: 1, Max: 5, Mean: 3, SumOfSquaredDev: 2.66666666666667 * 3, CountPerBucket: []int64{1, 3}, Bounds: []float64{2},
			},
		},
	}

	view.clearRows()
	view.subscribe()
	for _, r := range records {
		var mods []tag.Mutator
		for _, t := range r.tags {
			mods = append(mods, tag.Insert(t.k, t.v))
		}
		ctx, err := tag.New(context.Background(), mods...)
		if err != nil {
			t.Fatalf("NewMap = %+v", err)
		}
		view.addSample(tag.FromContext(ctx), r.f)
	}

	gotRows := view.collectedRows()
	for i, got := range gotRows {
		if !containsRow(wantRows, got) {
			t.Errorf("%d: got unexpected row %#v", i, got)
			break
		}
	}

	for i, want := range wantRows {
		if !containsRow(gotRows, want) {
			t.Errorf("%d: got none; want row %#v", i, want)
			break
		}
	}
}

func Test_View_MeasureFloat64_AggregationDistribution(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")
	k3, _ := tag.NewKey("k3")
	agg1 := Distribution(2)
	m := stats.Int64("Test_View_MeasureFloat64_AggregationDistribution/m1", "", stats.UnitNone)
	view1 := &View{
		TagKeys:     []tag.Key{k1, k2},
		Measure:     m,
		Aggregation: agg1,
	}
	view, err := newViewInternal(view1)
	if err != nil {
		t.Fatal(err)
	}

	type tagString struct {
		k tag.Key
		v string
	}
	type record struct {
		f    float64
		tags []tagString
	}

	type testCase struct {
		label    string
		records  []record
		wantRows []*exporter.Row
	}

	tcs := []testCase{
		{
			"1",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}}},
			},
			[]*exporter.Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					exporter.AggregationData{
						Count: 2, Min: 1, Max: 5, Mean: 3, SumOfSquaredDev: 8, CountPerBucket: []int64{1, 1}, Bounds: []float64{2},
					},
				},
			},
		},
		{
			"2",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k2, "v2"}}},
			},
			[]*exporter.Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					exporter.AggregationData{
						Count: 1, Min: 1, Max: 1, Mean: 1, CountPerBucket: []int64{1, 0}, Bounds: []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					exporter.AggregationData{
						Count: 1, Min: 5, Max: 5, Mean: 5, CountPerBucket: []int64{0, 1}, Bounds: []float64{2},
					},
				},
			},
		},
		{
			"3",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}, {k3, "v3"}}},
				{1, []tagString{{k1, "v1 other"}}},
				{5, []tagString{{k2, "v2"}}},
				{5, []tagString{{k1, "v1"}, {k2, "v2"}}},
			},
			[]*exporter.Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					exporter.AggregationData{
						Count: 2, Min: 1, Max: 5, Mean: 3, SumOfSquaredDev: 8, CountPerBucket: []int64{1, 1}, Bounds: []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1 other"}},
					exporter.AggregationData{
						Count: 1, Min: 1, Max: 1, Mean: 1, CountPerBucket: []int64{1, 0}, Bounds: []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					exporter.AggregationData{
						Count: 1, Min: 5, Max: 5, Mean: 5, CountPerBucket: []int64{0, 1}, Bounds: []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1"}, {Key: k2, Value: "v2"}},
					exporter.AggregationData{
						Count: 1, Min: 5, Max: 5, Mean: 5, CountPerBucket: []int64{0, 1}, Bounds: []float64{2},
					},
				},
			},
		},
		{
			"4",
			[]record{
				{1, []tagString{{k1, "v1 is a very long value key"}}},
				{5, []tagString{{k1, "v1 is a very long value key"}, {k3, "v3"}}},
				{1, []tagString{{k1, "v1 is another very long value key"}}},
				{1, []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
				{5, []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
				{3, []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
				{3, []tagString{{k1, "v1 is a very long value key"}, {k2, "v2 is a very long value key"}}},
			},
			[]*exporter.Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1 is a very long value key"}},
					exporter.AggregationData{
						Count: 2, Min: 1, Max: 5, Mean: 3, SumOfSquaredDev: 8, CountPerBucket: []int64{1, 1}, Bounds: []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1 is another very long value key"}},
					exporter.AggregationData{
						Count: 1, Min: 1, Max: 1, Mean: 1, CountPerBucket: []int64{1, 0}, Bounds: []float64{2},
					},
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1 is a very long value key"}, {Key: k2, Value: "v2 is a very long value key"}},
					exporter.AggregationData{
						Count: 4, Min: 1, Max: 5, Mean: 3, SumOfSquaredDev: 2.66666666666667 * 3, CountPerBucket: []int64{1, 3}, Bounds: []float64{2},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		view.clearRows()
		view.subscribe()
		for _, r := range tc.records {
			var mods []tag.Mutator
			for _, t := range r.tags {
				mods = append(mods, tag.Insert(t.k, t.v))
			}
			ctx, err := tag.New(context.Background(), mods...)
			if err != nil {
				t.Errorf("%v: NewMap = %v", tc.label, err)
			}
			view.addSample(tag.FromContext(ctx), r.f)
		}

		gotRows := view.collectedRows()
		for i, got := range gotRows {
			if !containsRow(tc.wantRows, got) {
				t.Errorf("%v-%d: got unexpected row %v", tc.label, i, got)
				break
			}
		}

		for i, want := range tc.wantRows {
			if !containsRow(gotRows, want) {
				t.Errorf("%v-%d: got none; want row %v", tc.label, i, want)
				break
			}
		}
	}
}

func Test_View_MeasureFloat64_AggregationSum(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")
	k3, _ := tag.NewKey("k3")
	m := stats.Int64("Test_View_MeasureFloat64_AggregationSum/m1", "", stats.UnitNone)
	view, err := newViewInternal(&View{TagKeys: []tag.Key{k1, k2}, Measure: m, Aggregation: Sum()})
	if err != nil {
		t.Fatal(err)
	}

	type tagString struct {
		k tag.Key
		v string
	}
	type record struct {
		f    float64
		tags []tagString
	}

	tcs := []struct {
		label    string
		records  []record
		wantRows []*exporter.Row
	}{
		{
			label: "1",
			records: []record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}}},
			},
			wantRows: []*exporter.Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					newSumDist(6),
				},
			},
		},
		{
			"2",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k2, "v2"}}},
			},
			[]*exporter.Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					newSumDist(1),
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					newSumDist(5),
				},
			},
		},
		{
			"3",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}, {k3, "v3"}}},
				{1, []tagString{{k1, "v1 other"}}},
				{5, []tagString{{k2, "v2"}}},
				{5, []tagString{{k1, "v1"}, {k2, "v2"}}},
			},
			[]*exporter.Row{
				{
					[]tag.Tag{{Key: k1, Value: "v1"}},
					newSumDist(6),
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1 other"}},
					newSumDist(1),
				},
				{
					[]tag.Tag{{Key: k2, Value: "v2"}},
					newSumDist(5),
				},
				{
					[]tag.Tag{{Key: k1, Value: "v1"}, {Key: k2, Value: "v2"}},
					newSumDist(5),
				},
			},
		},
	}

	for _, tt := range tcs {
		view.clearRows()
		view.subscribe()
		for _, r := range tt.records {
			var mods []tag.Mutator
			for _, t := range r.tags {
				mods = append(mods, tag.Insert(t.k, t.v))
			}
			ctx, err := tag.New(context.Background(), mods...)
			if err != nil {
				t.Errorf("%v: New = %v", tt.label, err)
			}
			view.addSample(tag.FromContext(ctx), r.f)
		}

		gotRows := view.collectedRows()
		for i, got := range gotRows {
			if !containsRow(tt.wantRows, got) {
				t.Errorf("%v-%d: got row %v; want none", tt.label, i, got)
				break
			}
		}

		for i, want := range tt.wantRows {
			if !containsRow(gotRows, want) {
				t.Errorf("%v-%d: got none; want row %v", tt.label, i, want)
				break
			}
		}
	}
}

func TestCanonicalize(t *testing.T) {
	k1, _ := tag.NewKey("k1")
	k2, _ := tag.NewKey("k2")
	m := stats.Int64("TestCanonicalize/m1", "desc desc", stats.UnitNone)
	v := &View{TagKeys: []tag.Key{k2, k1}, Measure: m, Aggregation: Sum()}
	err := v.canonicalize()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := v.Name, "TestCanonicalize/m1"; got != want {
		t.Errorf("vc.Name = %q; want %q", got, want)
	}
	if got, want := v.Description, "desc desc"; got != want {
		t.Errorf("vc.Description = %q; want %q", got, want)
	}
	if got, want := len(v.TagKeys), 2; got != want {
		t.Errorf("len(vc.TagKeys) = %d; want %d", got, want)
	}
	if got, want := v.TagKeys[0].Name(), "k1"; got != want {
		t.Errorf("vc.TagKeys[0].Name() = %q; want %q", got, want)
	}
}

func TestViewSortedKeys(t *testing.T) {
	k1, _ := tag.NewKey("a")
	k2, _ := tag.NewKey("b")
	k3, _ := tag.NewKey("c")
	ks := []tag.Key{k1, k3, k2}

	m := stats.Int64("TestViewSortedKeys/m1", "", stats.UnitNone)
	Register(&View{
		Name:        "sort_keys",
		Description: "desc sort_keys",
		TagKeys:     ks,
		Measure:     m,
		Aggregation: Sum(),
	})
	// Subscribe normalizes the view by sorting the tag keys, retrieve the normalized view
	v := Find("sort_keys")

	want := []string{"a", "b", "c"}
	vks := v.TagKeys
	if len(vks) != len(want) {
		t.Errorf("Keys = %+v; want %+v", vks, want)
	}

	for i, v := range want {
		if got, want := v, vks[i].Name(); got != want {
			t.Errorf("View name = %q; want %q", got, want)
		}
	}
}

// containsRow returns true if rows contain r.
func containsRow(rows []*exporter.Row, r *exporter.Row) bool {
	for _, x := range rows {
		if r.Equal(x) {
			return true
		}
	}
	return false
}
