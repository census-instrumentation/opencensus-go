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
	"testing"
	"time"

	"github.com/google/instrumentation-go/stats/tagging"
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

	k1, err := tagging.DefaultKeyManager().CreateKeyString("k1")
	if err != nil {
		t.Fatalf("creating keyString failed. %v ", err)
	}
	k2, err := tagging.DefaultKeyManager().CreateKeyString("k2")
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
						vdc: &ViewDescCommon{
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
						vdc: &ViewDescCommon{
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
						vdc: &ViewDescCommon{
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
						vdc: &ViewDescCommon{
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
						vdc: &ViewDescCommon{
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
