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
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/google/instrumentation-go/stats/tagging"
	"golang.org/x/net/context"
)

type record struct {
	t  time.Time
	ts tagging.TagsSet
	v  float64
}

type view struct {
	viewDesc     ViewDesc
	wantViewAgg  *DistributionView
	registerTime time.Time
	retrieveTime time.Time
}

type ucTestData struct {
	measureDesc MeasureDesc
	views       []*view
	records     []record
}

func (td *ucTestData) String() string {
	if td == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", td.measureDesc)
}

func TestUsageCollection(t *testing.T) {
	registerTime := time.Now()
	retrieveTime := registerTime.Add(10 * time.Second)

	k1, err := tagging.DefaultKeyManager().CreateKeyStringUTF8("k1")
	if err != nil {
		t.Fatalf("creating keyString failed. %v ", err)
	}
	k2, err := tagging.DefaultKeyManager().CreateKeyStringUTF8("k2")
	if err != nil {
		t.Fatalf("creating keyString failed. %v ", err)
	}
	uctds := []*ucTestData{
		{
			&measureDescFloat64{
				&measureDesc{
					name: "measure1",
					unit: &MeasurementUnit{1, []BasicUnit{BytesUnit}, []BasicUnit{}},
				},
			},
			[]*view{
				{
					viewDesc: &DistributionViewDesc{
						Vdc: &ViewDescCommon{
							Name:            "view1",
							MeasureDescName: "measure1",
							TagKeys:         []tagging.Key{k1, k2},
						},
						Bounds: []float64{15},
					},
					registerTime: registerTime,
					retrieveTime: retrieveTime,
					wantViewAgg: &DistributionView{
						Aggregations: []*DistributionAgg{
							{
								&DistributionStats{
									3,
									10,
									20,
									30,
									60,
									[]int64{1, 2},
								},
								[]tagging.Tag{k1.CreateTag("v1"), k2.CreateTag("v2")},
							},
							{
								&DistributionStats{
									3,
									10,
									20,
									30,
									60,
									[]int64{1, 2},
								},
								[]tagging.Tag{k1.CreateTag("v1")},
							},
						},
						Start: registerTime,
						End:   retrieveTime,
					},
				},
			},
			[]record{
				{
					registerTime.Add(1 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
					},
					10,
				},
				{
					registerTime.Add(2 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
					},
					20,
				},
				{
					registerTime.Add(3 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
					},
					30,
				},
				{
					registerTime.Add(4 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
						k2: k2.CreateTag("v2"),
					},
					10,
				},
				{
					registerTime.Add(5 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
						k2: k2.CreateTag("v2"),
					},
					20,
				},
				{
					registerTime.Add(6 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
						k2: k2.CreateTag("v2"),
					},
					30,
				},
			},
		},
		{
			&measureDescFloat64{
				&measureDesc{
					name: "measure2",
					unit: &MeasurementUnit{2, []BasicUnit{BytesUnit}, []BasicUnit{}},
				},
			},
			[]*view{
				{
					viewDesc: &DistributionViewDesc{
						Vdc: &ViewDescCommon{
							Name:            "allTagsView",
							MeasureDescName: "measure2",
							TagKeys:         []tagging.Key{},
						},
						Bounds: []float64{25},
					},
					registerTime: registerTime,
					retrieveTime: retrieveTime,
					wantViewAgg: &DistributionView{
						Aggregations: []*DistributionAgg{
							{
								&DistributionStats{
									6,
									10,
									20,
									30,
									120,
									[]int64{4, 2},
								},
								[]tagging.Tag(nil),
							},
						},
						Start: registerTime,
						End:   retrieveTime,
					},
				},
				{
					viewDesc: &DistributionViewDesc{
						Vdc: &ViewDescCommon{
							Name:            "view1",
							MeasureDescName: "measure2",
							TagKeys:         []tagging.Key{k1, k2},
						},
						Bounds: []float64{15},
					},
					registerTime: registerTime,
					retrieveTime: retrieveTime,
					wantViewAgg: &DistributionView{
						Aggregations: []*DistributionAgg{
							{
								&DistributionStats{
									3,
									10,
									20,
									30,
									60,
									[]int64{1, 2},
								},
								[]tagging.Tag{k1.CreateTag("v1"), k2.CreateTag("v2")},
							},
							{
								&DistributionStats{
									3,
									10,
									20,
									30,
									60,
									[]int64{1, 2},
								},
								[]tagging.Tag{k1.CreateTag("v1")},
							},
						},
						Start: registerTime,
						End:   retrieveTime,
					},
				},
				{
					viewDesc: &DistributionViewDesc{
						Vdc: &ViewDescCommon{
							Name:            "view2",
							MeasureDescName: "measure2",
							TagKeys:         []tagging.Key{k1, k2},
						},
						Bounds: []float64{25},
					},
					registerTime: registerTime,
					retrieveTime: retrieveTime,
					wantViewAgg: &DistributionView{
						Aggregations: []*DistributionAgg{
							{
								&DistributionStats{
									3,
									10,
									20,
									30,
									60,
									[]int64{2, 1},
								},
								[]tagging.Tag{k1.CreateTag("v1"), k2.CreateTag("v2")},
							},
							{
								&DistributionStats{
									3,
									10,
									20,
									30,
									60,
									[]int64{2, 1},
								},
								[]tagging.Tag{k1.CreateTag("v1")},
							},
						},
						Start: registerTime,
						End:   retrieveTime,
					},
				},
				{
					viewDesc: &DistributionViewDesc{
						Vdc: &ViewDescCommon{
							Name:            "view3",
							MeasureDescName: "measure2",
							TagKeys:         []tagging.Key{k1},
						},
						Bounds: []float64{25},
					},
					registerTime: registerTime,
					retrieveTime: retrieveTime,
					wantViewAgg: &DistributionView{
						Aggregations: []*DistributionAgg{
							{
								&DistributionStats{
									6,
									10,
									20,
									30,
									120,
									[]int64{4, 2},
								},
								[]tagging.Tag{k1.CreateTag("v1")},
							},
						},
						Start: registerTime,
						End:   retrieveTime,
					},
				},
			},
			[]record{
				{
					registerTime.Add(1 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
					},
					10,
				},
				{
					registerTime.Add(2 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
					},
					20,
				},
				{
					registerTime.Add(3 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
					},
					30,
				},
				{
					registerTime.Add(4 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
						k2: k2.CreateTag("v2"),
					},
					10,
				},
				{
					registerTime.Add(5 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
						k2: k2.CreateTag("v2"),
					},
					20,
				},
				{
					registerTime.Add(6 * time.Second),
					tagging.TagsSet{
						k1: k1.CreateTag("v1"),
						k2: k2.CreateTag("v2"),
					},
					30,
				},
			},
		},
	}

	for _, td := range uctds {
		uc := &usageCollector{
			mDescriptors: make(map[string]MeasureDesc),
			vDescriptors: make(map[string]ViewDesc),
		}
		td.measureDesc.Meta().aggViewDescs = make(map[ViewDesc]struct{})
		uc.registerMeasureDesc(td.measureDesc)
		for _, vw := range td.views {
			uc.registerViewDesc(vw.viewDesc, vw.registerTime)
		}

		for _, r := range td.records {
			m := &measurementFloat64{
				md: td.measureDesc,
				v:  r.v,
			}
			uc.recordMeasurement(r.t, r.ts, m)
		}

		for _, vw := range td.views {
			gotVw, err := uc.retrieveViewByName(vw.viewDesc.ViewDescCommon().Name, vw.retrieveTime)
			if err != nil {
				t.Errorf("got error %v (test case: %v), want no error", err, td)
			}

			switch gotVwAgg := gotVw.ViewAgg.(type) {
			case *DistributionView:
				if len(gotVwAgg.Aggregations) != len(vw.wantViewAgg.Aggregations) {
					t.Errorf("got %v aggregations (test case: %v, view:%v), want %v aggregations", len(gotVwAgg.Aggregations), td, vw.viewDesc.ViewDescCommon().Name, len(vw.wantViewAgg.Aggregations))
					continue
				}

				for _, gotAgg := range gotVwAgg.Aggregations {
					found := false
					for _, wantAgg := range vw.wantViewAgg.Aggregations {
						if reflect.DeepEqual(gotAgg, wantAgg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("got unexpected aggregation %v (test case: %v)", gotAgg, td)
					}
				}
			default:
				t.Errorf("got view aggregation type %v (test case: %v), want %T", gotVwAgg, td, vw.wantViewAgg)
			}

		}
	}
}

func registerKeys(count int) []tagging.KeyStringUTF8 {
	mgr := tagging.DefaultKeyManager()
	var keys []tagging.KeyStringUTF8

	for i := 0; i < count; i++ {
		k1, err := mgr.CreateKeyStringUTF8("keyIdentifier" + strconv.Itoa(i))
		if err != nil {
			glog.Fatalf("RegisterKeys(_) failed. %v\n", err)
		}
		keys = append(keys, k1)
	}
	return keys
}

func createMutations(keys []tagging.KeyStringUTF8) []tagging.Mutation {
	var mutations []tagging.Mutation
	for i, k := range keys {
		mutations = append(mutations, k.CreateMutation("valueIdentifier"+strconv.Itoa(i), tagging.BehaviorAddOrReplace))
	}
	return mutations
}

func registerMeasure(uc *usageCollector, n string) *measureDescFloat64 {
	mu := &MeasurementUnit{
		Power10: 6,
		Numerators: []BasicUnit{
			BytesUnit,
		},
	}
	mf64 := NewMeasureDescFloat64(n, "", mu)
	if err := uc.registerMeasureDesc(mf64); err != nil {
		glog.Fatalf("RegisterMeasure(_) failed. %v\n", err)
	}
	return mf64
}

func registerView(uc *usageCollector, n string, measureName string, keys []tagging.KeyStringUTF8) *DistributionViewDesc {
	vw := &DistributionViewDesc{
		Vdc: &ViewDescCommon{
			Name:            n,
			Description:     "",
			MeasureDescName: measureName,
		},
		Bounds: []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
	}
	for _, k := range keys {
		vw.Vdc.TagKeys = append(vw.Vdc.TagKeys, k)
	}
	if err := uc.registerViewDesc(vw, time.Now()); err != nil {
		glog.Fatalf("RegisterView(_) failed. %v\n", err)
	}
	return vw
}

// 10 keys. 1 measure. 1 view. 10 records.
func TestUsageCollector_10Keys_1Measure_1View_10Records(t *testing.T) {
	keys := registerKeys(10)
	mutations := createMutations(keys)

	uc := newUsageCollector()
	m := registerMeasure(uc, "m")
	_ = registerView(uc, "v", "m", keys)

	ctx := tagging.NewContextWithMutations(context.Background(), mutations...)
	ts := tagging.FromContext(ctx)

	for j := 0; j < 10; j++ {
		measurement := m.CreateMeasurement(float64(j))
		uc.recordMeasurement(time.Now(), ts, measurement)
	}
	retrieved := uc.retrieveViewsAdhoc(nil, nil, time.Now())

	if len(retrieved) != 1 {
		t.Fatalf("got %v views retrieved, want 1 view", len(retrieved))
	}

	dv, ok := retrieved[0].ViewAgg.(*DistributionView)
	if !ok {
		t.Errorf("got retrieved view of type %T, want view of type *DistributionView", dv)
	}

	if len(dv.Aggregations) != 1 {
		t.Errorf("got %v unique aggregations, want 1 single aggregation", len(dv.Aggregations))
	}

	for _, agg := range dv.Aggregations {
		if agg.DistributionStats.Count != 10 {
			t.Errorf("got %v records for aggregation %v, want 10 records", agg.DistributionStats.Count, agg)
		}
	}
}

func Benchmark_Create_1Measurement_Record_1Measurement(b *testing.B) {
	keys := registerKeys(10)
	mutations := createMutations(keys)
	uc := newUsageCollector()
	m := registerMeasure(uc, "m")

	for i := 0; i < 10; i++ {
		_ = registerView(uc, "v"+strconv.Itoa(i), "m", keys)
	}

	ctx := tagging.NewContextWithMutations(context.Background(), mutations...)
	ts := tagging.FromContext(ctx)

	measurement := m.CreateMeasurement(float64(1))
	for i := 0; i < b.N; i++ {
		uc.recordMeasurement(time.Now(), ts, measurement)
	}
}
