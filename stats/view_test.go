// Copyright 2017 Google Inc.
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

package stats

import (
	"reflect"
	"testing"
	"time"

	"github.com/census-instrumentation/opencensus-go/tags"
)

func rowFoundInRows(row *Row, rows []*Row) bool {
	for _, x := range rows {
		if rowsAreEqual(row, x) {
			return true
		}
	}
	return false
}

func rowsAreEqual(r1, r2 *Row) bool {
	return reflect.DeepEqual(r1.Tags, r2.Tags) && aggregationValueAreEqual(r1.AggregationValue, r2.AggregationValue)
}

func Test_View_MeasureFloat64_AggregationDistribution_WindowCumulative(t *testing.T) {
	k1, _ := tags.CreateKeyString("k1")
	k2, _ := tags.CreateKeyString("k2")
	k3, _ := tags.CreateKeyString("k3")
	agg1 := NewAggregationDistribution([]float64{2})
	vw1 := NewViewFloat64("VF1", "desc VF1", []tags.Key{k1, k2}, nil, agg1, NewWindowCumulative())

	type tagString struct {
		k *tags.KeyString
		v string
	}
	type record struct {
		f    float64
		tags []tagString
	}

	type testCase struct {
		label    string
		records  []record
		wantRows []*Row
	}

	tcs := []testCase{
		{
			"1",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}}},
			},
			[]*Row{
				{
					[]tags.Tag{{k1, []byte("v1")}},
					&AggregationDistributionValue{
						2, 1, 5, 6, []int64{1, 1}, agg1.bounds,
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
			[]*Row{
				{
					[]tags.Tag{{k1, []byte("v1")}},
					&AggregationDistributionValue{
						1, 1, 1, 1, []int64{1, 0}, agg1.bounds,
					},
				},
				{
					[]tags.Tag{{k2, []byte("v2")}},
					&AggregationDistributionValue{
						1, 5, 5, 5, []int64{0, 1}, agg1.bounds,
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
			[]*Row{
				{
					[]tags.Tag{{k1, []byte("v1")}},
					&AggregationDistributionValue{
						2, 1, 5, 6, []int64{1, 1}, agg1.bounds,
					},
				},
				{
					[]tags.Tag{{k1, []byte("v1 other")}},
					&AggregationDistributionValue{
						1, 1, 1, 1, []int64{1, 0}, agg1.bounds,
					},
				},
				{
					[]tags.Tag{{k2, []byte("v2")}},
					&AggregationDistributionValue{
						1, 5, 5, 5, []int64{0, 1}, agg1.bounds,
					},
				},
				{
					[]tags.Tag{{k1, []byte("v1")}, {k2, []byte("v2")}},
					&AggregationDistributionValue{
						1, 5, 5, 5, []int64{0, 1}, agg1.bounds,
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
			[]*Row{
				{
					[]tags.Tag{{k1, []byte("v1 is a very long value key")}},
					&AggregationDistributionValue{
						2, 1, 5, 6, []int64{1, 1}, agg1.bounds,
					},
				},
				{
					[]tags.Tag{{k1, []byte("v1 is another very long value key")}},
					&AggregationDistributionValue{
						1, 1, 1, 1, []int64{1, 0}, agg1.bounds,
					},
				},
				{
					[]tags.Tag{{k1, []byte("v1 is a very long value key")}, {k2, []byte("v2 is a very long value key")}},
					&AggregationDistributionValue{
						4, 1, 5, 12, []int64{1, 3}, agg1.bounds,
					},
				},
			},
		},
	}

	tsb := &tags.TagSetBuilder{}
	for _, tc := range tcs {
		vw1.clearRows()
		for _, r := range tc.records {
			tsb.StartFromEmpty()
			for _, t := range r.tags {
				tsb.InsertString(t.k, t.v)
			}
			vw1.addSample(tsb.Build(), r.f, time.Now())
		}

		gotRows := vw1.collectedRows(time.Now())

		for _, gotRow := range gotRows {
			if !rowFoundInRows(gotRow, tc.wantRows) {
				t.Errorf("got unexpected row '%v' for test case: '%v'", gotRow, tc.label)
				break
			}
		}

		for _, wantRow := range tc.wantRows {
			if !rowFoundInRows(wantRow, gotRows) {
				t.Errorf("want row '%v' for test case: '%v'. Not received", wantRow, tc.label)
				break
			}
		}
	}
}

func Test_View_MeasureFloat64_AggregationDistribution_WindowSlidingTime(t *testing.T) {
	startTime := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)

	k1, _ := tags.CreateKeyString("k1")
	k2, _ := tags.CreateKeyString("k2")
	agg1 := NewAggregationDistribution([]float64{2})
	vw1 := NewViewFloat64("VF1", "desc VF1", []tags.Key{k1, k2}, nil, agg1, NewWindowSlidingTime(10*time.Second, 5))

	type tagString struct {
		k *tags.KeyString
		v string
	}
	type record struct {
		f    float64
		tags []tagString
		now  time.Time
	}

	type wantRows struct {
		label        string
		retrieveTime time.Time
		rows         []*Row
	}

	type testCase struct {
		label    string
		records  []record
		wantRows []wantRows
	}

	tcs := []testCase{
		{
			"1",
			[]record{
				{1, []tagString{{k1, "v1"}}, startTime.Add(1 * time.Second)},
				{2, []tagString{{k1, "v1"}}, startTime.Add(6 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(6 * time.Second)},
				{4, []tagString{{k1, "v1"}}, startTime.Add(10 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(10 * time.Second)},
				{4, []tagString{{k1, "v1"}}, startTime.Add(14 * time.Second)},
				{3, []tagString{{k1, "v1"}}, startTime.Add(14 * time.Second)},
			},
			[]wantRows{
				{
					"last 6 recorded",
					startTime.Add(14 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								6, 2, 5, 23, []int64{0, 6}, agg1.bounds,
							},
						},
					},
				},
				{
					"last 4 recorded",
					startTime.Add(18 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								4, 3, 5, 16, []int64{0, 4}, agg1.bounds,
							},
						},
					},
				},
				{
					"last 2 recorded",
					startTime.Add(22 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								2, 3, 4, 7, []int64{0, 2}, agg1.bounds,
							},
						},
					},
				},
			},
		},
		{
			"2",
			[]record{
				{1, []tagString{{k1, "v1"}}, startTime.Add(3 * time.Second)},
				{2, []tagString{{k1, "v1"}}, startTime.Add(5 * time.Second)},
				{3, []tagString{{k1, "v1"}}, startTime.Add(5 * time.Second)},
				{4, []tagString{{k1, "v1"}}, startTime.Add(8 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(8 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(8 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(9 * time.Second)},
			},
			[]wantRows{
				{
					"no partial bucket",
					startTime.Add(10 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								7, 1, 5, 25, []int64{1, 6}, agg1.bounds,
							},
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 50%) (count: 0)",
					startTime.Add(12 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								7, 1, 5, 25, []int64{1, 6}, agg1.bounds,
							},
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 50%) (count: 1)",
					startTime.Add(12 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								7, 1, 5, 25, []int64{1, 6}, agg1.bounds,
							},
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 80%) (count: 2)",
					startTime.Add(15*time.Second + 400*time.Millisecond),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								6, 2, 5, 24, []int64{0, 6}, agg1.bounds,
							},
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 50%) (count: 2)",
					startTime.Add(16 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								5, 2.5, 5, 21.5, []int64{0, 5}, agg1.bounds,
							},
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 90%) (count: 3)",
					startTime.Add(17*time.Second + 200*time.Millisecond),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								4, 4, 5, 18.5, []int64{0, 4}, agg1.bounds,
							},
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 50%) (count: 3)",
					startTime.Add(18 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							&AggregationDistributionValue{
								3, 4, 5, 14, []int64{0, 3}, agg1.bounds,
							},
						},
					},
				},
			},
		},
	}

	tsb := &tags.TagSetBuilder{}
	for _, tc := range tcs {
		vw1.clearRows()
		for _, r := range tc.records {
			tsb.StartFromEmpty()
			for _, t := range r.tags {
				tsb.InsertString(t.k, t.v)
			}
			vw1.addSample(tsb.Build(), r.f, r.now)
		}

		for _, wantRows := range tc.wantRows {
			gotRows := vw1.collectedRows(wantRows.retrieveTime)

			for _, gotRow := range gotRows {
				if !rowFoundInRows(gotRow, wantRows.rows) {
					t.Errorf("got unexpected row '%v' for test case: '%v' with label '%v'", gotRow, tc.label, wantRows.label)
					break
				}
			}

			for _, wantRow := range wantRows.rows {
				if !rowFoundInRows(wantRow, gotRows) {
					t.Errorf("want row '%v' for test case: '%v' with label '%v'. Not received", wantRow, tc.label, wantRows.label)
					break
				}
			}
		}

	}
}

func Test_View_MeasureFloat64_AggregationCount_WindowSlidingTime(t *testing.T) {
	startTime := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)

	k1, _ := tags.CreateKeyString("k1")
	k2, _ := tags.CreateKeyString("k2")
	agg1 := NewAggregationCount()
	vw1 := NewViewFloat64("VF1", "desc VF1", []tags.Key{k1, k2}, nil, agg1, NewWindowSlidingTime(10*time.Second, 5))

	type tagString struct {
		k *tags.KeyString
		v string
	}
	type record struct {
		f    float64
		tags []tagString
		now  time.Time
	}

	type wantRows struct {
		label        string
		retrieveTime time.Time
		rows         []*Row
	}

	type testCase struct {
		label    string
		records  []record
		wantRows []wantRows
	}

	tcs := []testCase{
		{
			"1",
			[]record{
				{1, []tagString{{k1, "v1"}}, startTime.Add(1 * time.Second)},
				{2, []tagString{{k1, "v1"}}, startTime.Add(6 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(6 * time.Second)},
				{4, []tagString{{k1, "v1"}}, startTime.Add(10 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(10 * time.Second)},
				{4, []tagString{{k1, "v1"}}, startTime.Add(14 * time.Second)},
				{3, []tagString{{k1, "v1"}}, startTime.Add(14 * time.Second)},
			},
			[]wantRows{
				{
					"last 6 recorded",
					startTime.Add(14 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(6),
						},
					},
				},
				{
					"last 4 recorded",
					startTime.Add(18 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(4),
						},
					},
				},
				{
					"last 2 recorded",
					startTime.Add(22 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(2),
						},
					},
				},
			},
		},
		{
			"2",
			[]record{
				{1, []tagString{{k1, "v1"}}, startTime.Add(3 * time.Second)},
				{2, []tagString{{k1, "v1"}}, startTime.Add(5 * time.Second)},
				{3, []tagString{{k1, "v1"}}, startTime.Add(5 * time.Second)},
				{4, []tagString{{k1, "v1"}}, startTime.Add(8 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(8 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(8 * time.Second)},
				{5, []tagString{{k1, "v1"}}, startTime.Add(9 * time.Second)},
			},
			[]wantRows{
				{
					"no partial bucket",
					startTime.Add(10 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(7),
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 50%) (count: 0)",
					startTime.Add(12 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(7),
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 50%) (count: 1)",
					startTime.Add(12 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(7),
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 80%) (count: 2)",
					startTime.Add(15*time.Second + 400*time.Millisecond),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(6),
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 50%) (count: 2)",
					startTime.Add(16 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(5),
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 90%) (count: 3)",
					startTime.Add(17*time.Second + 200*time.Millisecond),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(4),
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 50%) (count: 3)",
					startTime.Add(18 * time.Second),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(3),
						},
					},
				},
				{
					"oldest partial bucket: (remaining time: 20%) (count: 3)",
					startTime.Add(18*time.Second + 600*time.Millisecond),
					[]*Row{
						{
							[]tags.Tag{{k1, []byte("v1")}},
							newAggregationCountValue(2),
						},
					},
				},
			},
		},
	}

	tsb := &tags.TagSetBuilder{}
	for _, tc := range tcs {
		vw1.clearRows()
		for _, r := range tc.records {
			tsb.StartFromEmpty()
			for _, t := range r.tags {
				tsb.InsertString(t.k, t.v)
			}
			vw1.addSample(tsb.Build(), r.f, r.now)
		}

		for _, wantRows := range tc.wantRows {
			gotRows := vw1.collectedRows(wantRows.retrieveTime)

			for _, gotRow := range gotRows {
				if !rowFoundInRows(gotRow, wantRows.rows) {
					t.Errorf("got unexpected row '%v' for test case: '%v' with label '%v'", gotRow, tc.label, wantRows.label)
					break
				}
			}

			for _, wantRow := range wantRows.rows {
				if !rowFoundInRows(wantRow, gotRows) {
					t.Errorf("want row '%v' for test case: '%v' with label '%v'. Not received", wantRow, tc.label, wantRows.label)
					break
				}
			}
		}

	}
}

func Test_View_MeasureFloat64_AggregationDistribution_WindowSlidingCount(t *testing.T) {
	k1, _ := tags.CreateKeyString("k1")
	k2, _ := tags.CreateKeyString("k2")
	agg1 := NewAggregationDistribution([]float64{2})
	vw1 := NewViewFloat64("VF1", "desc VF1", []tags.Key{k1, k2}, nil, agg1, NewWindowSlidingCount(4, 2))

	type tagString struct {
		k *tags.KeyString
		v string
	}
	type record struct {
		f    float64
		tags []tagString
	}

	type testCase struct {
		label   string
		records []record
		rows    []*Row
	}

	tcs := []testCase{
		{
			"1",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{2, []tagString{{k1, "v1"}}},
				{3, []tagString{{k1, "v1"}}},
				{4, []tagString{{k1, "v1"}}},
			},
			[]*Row{
				{
					[]tags.Tag{{k1, []byte("v1")}},
					&AggregationDistributionValue{
						4, 1, 4, 10, []int64{1, 3}, agg1.bounds,
					},
				},
			},
		},
		{
			"2",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{2, []tagString{{k1, "v1"}}},
				{3, []tagString{{k1, "v1"}}},
				{4, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}}},
				{6, []tagString{{k1, "v1"}}},
			},
			[]*Row{
				{
					[]tags.Tag{{k1, []byte("v1")}},
					&AggregationDistributionValue{
						4, 3, 6, 18, []int64{0, 4}, agg1.bounds,
					},
				},
			},
		},
		{
			"3",
			[]record{
				{1, []tagString{{k1, "v1"}}},
				{2, []tagString{{k1, "v1"}}},
				{3, []tagString{{k1, "v1"}}},
				{4, []tagString{{k1, "v1"}}},
				{5, []tagString{{k1, "v1"}}},
			},
			[]*Row{
				{
					[]tags.Tag{{k1, []byte("v1")}},
					&AggregationDistributionValue{
						4, 1.5, 5, 13.5, []int64{1, 3}, agg1.bounds,
					},
				},
			},
		},
	}

	tsb := &tags.TagSetBuilder{}
	for _, tc := range tcs {
		vw1.clearRows()
		for _, r := range tc.records {
			tsb.StartFromEmpty()
			for _, t := range r.tags {
				tsb.InsertString(t.k, t.v)
			}
			vw1.addSample(tsb.Build(), r.f, time.Now())
		}

		gotRows := vw1.collectedRows(time.Now())

		for _, gotRow := range gotRows {
			if !rowFoundInRows(gotRow, tc.rows) {
				t.Errorf("got unexpected row '%v' for test case: '%v'", gotRow, tc.label)
				break
			}
		}

		for _, wantRow := range tc.rows {
			if !rowFoundInRows(wantRow, gotRows) {
				t.Errorf("want row '%v' for test case: '%v'. Not received", wantRow, tc.label)
				break
			}
		}
	}
}
