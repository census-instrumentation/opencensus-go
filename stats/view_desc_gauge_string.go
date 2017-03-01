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

// GaugeStringViewDesc defines an string gauge view.
type GaugeStringViewDesc struct {
	vdc *ViewDescCommon
}

func (gd *GaugeStringViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newGaugeAggregatorString(), nil
}

func (gd *GaugeStringViewDesc) retrieveView(now time.Time) (*View, error) {
	gav, err := gd.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		ViewDesc: gd,
		ViewAgg:  gav,
	}, nil
}

func (gd *GaugeStringViewDesc) ViewDescCommon() *ViewDescCommon {
	return gd.vdc
}

func (gd *GaugeStringViewDesc) isValid() error {
	return nil
}

func (gd *GaugeStringViewDesc) retrieveAggreationView(t time.Time) (*GaugeStringView, error) {
	var aggs []*GaugeStringAgg

	for sig, a := range gd.vdc.signatures {
		tags, err := tagging.TagsFromValuesSignature([]byte(sig), gd.vdc.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature '%v'. %v", sig, err)
		}
		aggregator, ok := a.(*gaugeAggregatorString)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregator type. got %T, want stats.gaugeAggregatorString", a)
		}
		ga := &GaugeStringAgg{
			GaugeStringStats: aggregator.retrieveCollected(),
			Tags:             tags,
		}
		aggs = append(aggs, ga)
	}

	return &GaugeStringView{
		Descriptor:   gd,
		Aggregations: aggs,
	}, nil
}

func (gd *GaugeStringViewDesc) stringWithIndent(tabs string) string {
	if gd == nil {
		return "nil"
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", gd)
	fmt.Fprintf(&buf, "%v  Name: %v,\n", tabs, gd.vdc.Name)
	fmt.Fprintf(&buf, "%v  Description: %v,\n", tabs, gd.vdc.Description)
	fmt.Fprintf(&buf, "%v  MeasureDescName: %v,\n", tabs, gd.vdc.MeasureDescName)
	fmt.Fprintf(&buf, "%v  TagKeys: %v,\n", tabs, gd.vdc.TagKeys)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (gd *GaugeStringViewDesc) String() string {
	return gd.stringWithIndent("")
}

// GaugeStringView is the set of collected GaugeStringAgg associated with
// ViewDesc.
type GaugeStringView struct {
	Descriptor   *GaugeStringViewDesc
	Aggregations []*GaugeStringAgg
}

func (gv *GaugeStringView) stringWithIndent(tabs string) string {
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

func (gv *GaugeStringView) String() string {
	return gv.stringWithIndent("")
}

// A GaugeStringAgg is a statistical summary of measures associated with a
// unique tag set.
type GaugeStringAgg struct {
	*GaugeStringStats
	Tags []tagging.Tag
}

func (ga *GaugeStringAgg) stringWithIndent(tabs string) string {
	if ga == nil {
		return "nil"
	}

	tabs2 := tabs + "  "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", ga)
	fmt.Fprintf(&buf, "%v  Stats: %v,\n", tabs, ga.GaugeStringStats.stringWithIndent(tabs2))
	fmt.Fprintf(&buf, "%v  Tags: %v,\n", tabs, ga.Tags)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (ga *GaugeStringAgg) String() string {
	return ga.stringWithIndent("")
}
