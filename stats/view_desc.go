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

// AggregationViewDesc is the interface that all aggregations are expected to
// implement.
type ViewDesc interface {
	// creates an aggregator instance for a unique tags signature.
	createAggregator(t time.Time) (aggregator, error)
	// retrieves the collected *View holding collected data by all the
	// aggregator instances..
	retrieveView(now time.Time) (*View, error)
	// returns the *ViewDesc associated with this AggregationViewDesc
	viewDesc() *ViewDescCommon
	// validates the input recieved as requested by the client code.
	isValid() error
	// stringWithIndent returns String() with 'tabs' prefix
	stringWithIndent(tabs string) string
}

// aggregator is the interface that the aggregators created by an aggregation
// are expected to implement.
type aggregator interface {
	addSample(v Measurement, t time.Time)
}

type viewAggregation interface {
	// stringWithIndent print to string with 'tabs' prefix
	stringWithIndent(tabs string) string
}

// ViewDescCommon is a helper data structure that holds common fields to all
// ViewAggregationDesc. It should never be used standalone but always as part
// of a ViewAggregationDesc.
type ViewDescCommon struct {
	// Name of ViewDesc. Must be unique.
	// TODO(iamm2): provide examples for Name.
	Name string
	// TODO(iamm2): provide an example for description.
	Description string

	// MeasureDescName is the name of a Measure. Examples are cpu:tickCount,
	// diskio:time...
	MeasureDescName string

	// Keys to perform the aggregation on.
	TagKeys []tagging.Key

	// start is time when ViewDesc was registered.
	start time.Time

	// vChans are the channels through which the collected views for this ViewDesc
	// are sent to the consumers of this view.
	vChans map[chan *View]struct{}

	// signatures holds the aggregations for each unique tag signature (values
	// for all keys) to its *stats.Aggregator.
	signatures map[string]aggregator
}

// A View is a set of Aggregations about usage of the single resource
// associated with the given view during a particular time interval. Each
// Aggregation is specific to a unique set of tags. The Census infrastructure
// reports a stream of View events to the application for further processing
// such as further aggregations, logging and export to other services.
type View struct {
	ViewDesc ViewDesc
	// ViewAgg is expected to be a *DistributionAggView or a
	// *IntervalAggView
	ViewAgg viewAggregation
}

func (vw *View) stringWithIndent(tabs string) string {
	if vw == nil {
		return "nil"
	}
	tabs2 := tabs + "  "
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", vw)
	fmt.Fprintf(&buf, "%v  ViewDesc: %v,\n", tabs, vw.ViewDesc.stringWithIndent(tabs2))
	fmt.Fprintf(&buf, "%v  ViewAgg: %v,\n", tabs, vw.ViewAgg.stringWithIndent(tabs2))
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (vw *View) String() string {
	return vw.stringWithIndent("")
}
