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

// IntervalAggViewDesc holds the parameters describing an interval aggregation.
type IntervalViewDesc struct {
	*ViewDescCommon

	// Number of internal sub-intervals to use when collecting stats for each
	// interval. The max error in interval measurements will be approximately
	// 1/SubIntervals (although in practice, this will only be approached in
	// the presence of very large and bursty workload changes), and underlying
	// memory usage will be roughly proportional to the value of this
	// field. Must be in the range [2, 20]. A value of 5 will be used if this
	// is unspecified.
	SubIntervals int

	// The size of each interval, as a time duration. Must have at least one
	// element.
	Intervals []time.Duration
}

func (id *IntervalViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newIntervalsAggregator(t, id.Intervals, id.SubIntervals), nil
}

func (id *IntervalViewDesc) retrieveView(now time.Time) (*View, error) {
	iav, err := id.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		ViewDesc: id,
		ViewAgg:  iav,
	}, nil
}

func (id *IntervalViewDesc) viewDesc() *ViewDescCommon {
	return id.ViewDescCommon
}

func (id *IntervalViewDesc) isValid() error {
	if id.SubIntervals < 2 || id.SubIntervals < 20 {
		return fmt.Errorf("%v error. subIntervals is not in [2,20]", id)
	}
	return nil
}

func (id *IntervalViewDesc) retrieveAggreationView(now time.Time) (*IntervalAggView, error) {
	var aggs []*IntervalAgg

	for sig, a := range id.signatures {
		tags, err := tagging.TagsFromValuesSignature([]byte(sig), id.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature '%v'. %v", sig, err)
		}
		aggregator, ok := a.(*intervalsAggregator)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregator type. got %T, want stats.intervalsAggregator", a)
		}
		ia := &IntervalAgg{
			IntervalStats: aggregator.retrieveCollected(now),
			Tags:          tags,
		}
		aggs = append(aggs, ia)
	}

	return &IntervalAggView{
		Descriptor:   id,
		Aggregations: aggs,
	}, nil
}

func (id *IntervalViewDesc) stringWithIndent(tabs string) string {
	if id == nil {
		return "nil"
	}
	vd := id.ViewDescCommon
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", id)
	fmt.Fprintf(&buf, "%v  Name: %v,\n", tabs, vd.Name)
	fmt.Fprintf(&buf, "%v  Description: %v,\n", tabs, vd.Description)
	fmt.Fprintf(&buf, "%v  MeasureDescName: %v,\n", tabs, vd.MeasureDescName)
	fmt.Fprintf(&buf, "%v  TagKeys: %v,\n", tabs, vd.TagKeys)
	fmt.Fprintf(&buf, "%v  Intervals: %v,\n", tabs, id.Intervals)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (id *IntervalViewDesc) String() string {
	return id.stringWithIndent("")
}

// IntervalAggView is the set of collected IntervalAgg associated with
// ViewDesc.
type IntervalAggView struct {
	Descriptor   *IntervalViewDesc
	Aggregations []*IntervalAgg
}

func (iv *IntervalAggView) stringWithIndent(tabs string) string {
	if iv == nil {
		return "nil"
	}

	tabs2 := tabs + "    "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", iv)
	fmt.Fprintf(&buf, "%v  Aggregations:\n", tabs)
	for _, agg := range iv.Aggregations {
		fmt.Fprintf(&buf, "%v%v,\n", tabs2, agg.stringWithIndent(tabs2))
	}
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (iv *IntervalAggView) String() string {
	return iv.stringWithIndent("")
}

// IntervalAgg is a statistical summary of measures associated with a unique
// tag set for a specific time interval.
type IntervalAgg struct {
	IntervalStats []*IntervalStats
	Tags          []tagging.Tag
}

func (ia *IntervalAgg) stringWithIndent(tabs string) string {
	if ia == nil {
		return "nil"
	}

	tabs2 := tabs + "  "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", ia)

	fmt.Fprintf(&buf, "%v  IntervalStats:\n", tabs)
	for _, is := range ia.IntervalStats {
		fmt.Fprintf(&buf, "%v%v,\n", tabs2, is.stringWithIndent(tabs2))
	}
	fmt.Fprintf(&buf, "%v  Tags: %v,\n", tabs, ia.Tags)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (ia *IntervalAgg) String() string {
	return ia.stringWithIndent("")
}
