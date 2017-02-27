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

// DistributionViewDesc holds the parameters describing an aggregation
// distribution..
type DistributionViewDesc struct {
	*ViewDescCommon

	// An aggregation distribution may contain a histogram of the values in the
	// population. The bucket boundaries for that histogram are described
	// by Bounds. This defines len(Bounds)+1 buckets.
	//
	// if len(Bounds) >= 2 then the boundaries for bucket index i are:
	// [-infinity, bounds[i]) for i = 0
	// [bounds[i-1], bounds[i]) for 0 < i < len(Bounds)
	// [bounds[i-1], +infinity) for i = len(Bounds)
	//
	// if len(Bounds) == 0 then there is no histogram associated with the
	// distribution. There will be a single bucket with boundaries
	// (-infinity, +infinity).
	//
	// if len(Bounds) == 1 then there is no finite buckets, and that single
	// element is the common boundary of the overflow and underflow buckets.
	Bounds []float64
}

func (dd *DistributionViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newDistributionAggregator(dd.Bounds), nil
}

func (dd *DistributionViewDesc) retrieveView(now time.Time) (*View, error) {
	dav, err := dd.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		ViewDesc: dd,
		ViewAgg:  dav,
	}, nil
}

func (dd *DistributionViewDesc) viewDesc() *ViewDescCommon {
	return dd.ViewDescCommon
}

func (dd *DistributionViewDesc) isValid() error {
	for i := 1; i < len(dd.Bounds); i++ {
		if dd.Bounds[i-1] >= dd.Bounds[i] {
			return fmt.Errorf("%v error. bounds are not increasing", dd)
		}
	}
	return nil
}

func (dd *DistributionViewDesc) retrieveAggreationView(t time.Time) (*DistributionView, error) {
	var aggs []*DistributionAgg

	for sig, a := range dd.signatures {
		tags, err := tagging.TagsFromValuesSignature([]byte(sig), dd.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature '%v'. %v", sig, err)
		}
		aggregator, ok := a.(*distributionAggregator)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregator type. got %T, want stats.distributionAggregator", a)
		}
		da := &DistributionAgg{
			DistributionStats: aggregator.retrieveCollected(),
			Tags:              tags,
		}
		aggs = append(aggs, da)
	}

	return &DistributionView{
		Descriptor:   dd,
		Aggregations: aggs,
		Start:        dd.start,
		End:          t,
	}, nil
}

func (dd *DistributionViewDesc) stringWithIndent(tabs string) string {
	if dd == nil {
		return "nil"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", dd)
	fmt.Fprintf(&buf, "%v  Name: %v,\n", tabs, dd.Name)
	fmt.Fprintf(&buf, "%v  Description: %v,\n", tabs, dd.Description)
	fmt.Fprintf(&buf, "%v  MeasureDescName: %v,\n", tabs, dd.MeasureDescName)
	fmt.Fprintf(&buf, "%v  TagKeys: %v,\n", tabs, dd.TagKeys)
	fmt.Fprintf(&buf, "%v  Bound: %v,\n", tabs, dd.Bounds)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (dd *DistributionViewDesc) String() string {
	return dd.stringWithIndent("")
}

// DistributionView is the set of collected DistributionAgg associated with
// ViewDesc.
type DistributionView struct {
	Descriptor   *DistributionViewDesc
	Aggregations []*DistributionAgg
	Start, End   time.Time // start is time when ViewDesc was registered.
}

func (dv *DistributionView) stringWithIndent(tabs string) string {
	if dv == nil {
		return "nil"
	}

	tabs2 := tabs + "    "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", dv)
	fmt.Fprintf(&buf, "%v  Start: %v,\n", tabs, dv.Start)
	fmt.Fprintf(&buf, "%v  End: %v,\n", tabs, dv.End)
	fmt.Fprintf(&buf, "%v  Aggregations:\n", tabs)
	for _, agg := range dv.Aggregations {
		fmt.Fprintf(&buf, "%v%v,\n", tabs2, agg.stringWithIndent(tabs2))
	}
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (dv *DistributionView) String() string {
	return dv.stringWithIndent("")
}

// An DistributionAgg is a statistical summary of measures associated with a
// unique tag set.
type DistributionAgg struct {
	*DistributionStats
	Tags []tagging.Tag
}

func (da *DistributionAgg) stringWithIndent(tabs string) string {
	if da == nil {
		return "nil"
	}

	tabs2 := tabs + "  "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", da)
	fmt.Fprintf(&buf, "%v  Aggregations: %v,\n", tabs, da.DistributionStats.stringWithIndent(tabs2))
	fmt.Fprintf(&buf, "%v  Tags: %v,\n", tabs, da.Tags)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (da *DistributionAgg) String() string {
	return da.stringWithIndent("")
}
