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

// GaugeFloat64ViewDesc defines an float64 gauge view.
type GaugeFloat64ViewDesc struct {
	*ViewDescCommon
}

func (gd *GaugeFloat64ViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newGaugeAggregatorFloat64(), nil
}

func (gd *GaugeFloat64ViewDesc) retrieveView(now time.Time) (*View, error) {
	gav, err := gd.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		ViewDesc: gd,
		ViewAgg:  gav,
	}, nil
}

func (gd *GaugeFloat64ViewDesc) viewDesc() *ViewDescCommon {
	return gd.ViewDescCommon
}

func (gd *GaugeFloat64ViewDesc) isValid() error {
	return nil
}

func (gd *GaugeFloat64ViewDesc) retrieveAggreationView(t time.Time) (*GaugeFloat64View, error) {
	var aggs []*GaugeFloat64Agg

	for sig, a := range gd.signatures {
		tags, err := tagging.TagsFromValuesSignature([]byte(sig), gd.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature '%v'. %v", sig, err)
		}
		aggregator, ok := a.(*gaugeAggregatorFloat64)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregator type. got %T, want stats.gaugeAggregatorFloat64", a)
		}
		ga := &GaugeFloat64Agg{
			GaugeFloat64Stats: aggregator.retrieveCollected(),
			Tags:              tags,
		}
		aggs = append(aggs, ga)
	}

	return &GaugeFloat64View{
		Descriptor:   gd,
		Aggregations: aggs,
	}, nil
}

func (gd *GaugeFloat64ViewDesc) stringWithIndent(tabs string) string {
	if gd == nil {
		return "nil"
	}
	vd := gd.ViewDescCommon
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", gd)
	fmt.Fprintf(&buf, "%v  Name: %v,\n", tabs, vd.Name)
	fmt.Fprintf(&buf, "%v  Description: %v,\n", tabs, vd.Description)
	fmt.Fprintf(&buf, "%v  MeasureDescName: %v,\n", tabs, vd.MeasureDescName)
	fmt.Fprintf(&buf, "%v  TagKeys: %v,\n", tabs, vd.TagKeys)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (gd *GaugeFloat64ViewDesc) String() string {
	return gd.stringWithIndent("")
}

// GaugeFloat64View is the set of collected GaugeFloat64Agg associated with
// ViewDesc.
type GaugeFloat64View struct {
	Descriptor   *GaugeFloat64ViewDesc
	Aggregations []*GaugeFloat64Agg
}

func (gv *GaugeFloat64View) stringWithIndent(tabs string) string {
	if gv == nil {
		return "nil"
	}

	tabs2 := tabs + "    "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", gv)
	fmt.Fprintf(&buf, "%v  Aggregations:\n", tabs)
	for _, agg := range gv.Aggregations {
		fmt.Fprintf(&buf, "%v%v,\n", tabs2, agg.stringWithIndent(tabs2))
	}
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (gv *GaugeFloat64View) String() string {
	return gv.stringWithIndent("")
}

// A GaugeFloat64Agg is a statistical summary of measures associated with a
// unique tag set.
type GaugeFloat64Agg struct {
	*GaugeFloat64Stats
	Tags []tagging.Tag
}

func (ga *GaugeFloat64Agg) stringWithIndent(tabs string) string {
	if ga == nil {
		return "nil"
	}

	tabs2 := tabs + "  "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", ga)
	fmt.Fprintf(&buf, "%v  Stats: %v,\n", tabs, ga.GaugeFloat64Stats.stringWithIndent(tabs2))
	fmt.Fprintf(&buf, "%v  Tags: %v,\n", tabs, ga.Tags)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (ga *GaugeFloat64Agg) String() string {
	return ga.stringWithIndent("")
}
