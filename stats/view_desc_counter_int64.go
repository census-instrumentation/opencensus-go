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
	"bytes"
	"fmt"
	"time"

	"github.com/google/instrumentation-go/stats/tagging"
)

// CounterInt64ViewDesc defines an int64 counter view.
type CounterInt64ViewDesc struct {
	Vdc *ViewDescCommon
}

func (gd *CounterInt64ViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newCounterAggregatorInt64(), nil
}

func (gd *CounterInt64ViewDesc) retrieveView(now time.Time) (*View, error) {
	gav, err := gd.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		ViewDesc: gd,
		ViewAgg:  gav,
	}, nil
}

func (gd *CounterInt64ViewDesc) ViewDescCommon() *ViewDescCommon {
	return gd.Vdc
}

func (gd *CounterInt64ViewDesc) isValid() error {
	return nil
}

func (gd *CounterInt64ViewDesc) retrieveAggreationView(t time.Time) (*CounterInt64View, error) {
	var aggs []*CounterInt64Agg

	for sig, a := range gd.Vdc.signatures {
		tags, err := tagging.DecodeFromValuesSignatureToSlice([]byte(sig), gd.Vdc.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature '%v'. %v", sig, err)
		}
		aggregator, ok := a.(*counterAggregatorInt64)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregator type. got %T, want stats.counterAggregatorInt64", a)
		}
		ga := &CounterInt64Agg{
			CounterInt64Stats: aggregator.retrieveCollected(),
			Tags:              tags,
		}
		aggs = append(aggs, ga)
	}

	return &CounterInt64View{
		Descriptor:   gd,
		Aggregations: aggs,
	}, nil
}

func (gd *CounterInt64ViewDesc) stringWithIndent(tabs string) string {
	if gd == nil {
		return "nil"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", gd)
	fmt.Fprintf(&buf, "%v  Name: %v,\n", tabs, gd.Vdc.Name)
	fmt.Fprintf(&buf, "%v  Description: %v,\n", tabs, gd.Vdc.Description)
	fmt.Fprintf(&buf, "%v  MeasureDescName: %v,\n", tabs, gd.Vdc.MeasureDescName)
	fmt.Fprintf(&buf, "%v  TagKeys: %v,\n", tabs, gd.Vdc.TagKeys)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (gd *CounterInt64ViewDesc) String() string {
	return gd.stringWithIndent("")
}

// CounterInt64View is the set of collected CounterInt64Agg associated with
// ViewDesc.
type CounterInt64View struct {
	Descriptor   *CounterInt64ViewDesc
	Aggregations []*CounterInt64Agg
}

func (gv *CounterInt64View) stringWithIndent(tabs string) string {
	if gv == nil {
		return "nil"
	}

	tabs2 := tabs + "    "
	tabs3 := tabs2 + "  "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", gv)
	fmt.Fprintf(&buf, "%v  Aggregations:\n", tabs)
	for _, agg := range gv.Aggregations {
		fmt.Fprintf(&buf, "%v%v,\n", tabs2, agg.stringWithIndent(tabs3))
	}
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (gv *CounterInt64View) String() string {
	return gv.stringWithIndent("")
}

// A CounterInt64Agg is a statistical summary of measures associated with a
// unique tag set.
type CounterInt64Agg struct {
	*CounterInt64Stats
	Tags []tagging.Tag
}

func (ga *CounterInt64Agg) stringWithIndent(tabs string) string {
	if ga == nil {
		return "nil"
	}

	tabs2 := tabs + "  "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", ga)
	fmt.Fprintf(&buf, "%v  Stats: %v,\n", tabs, ga.CounterInt64Stats.stringWithIndent(tabs2))
	fmt.Fprintf(&buf, "%v  Tags: %v,\n", tabs, ga.Tags)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (ga *CounterInt64Agg) String() string {
	return ga.stringWithIndent("")
}
