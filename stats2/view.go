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

// Package stats defines the stats collection API and its native Go
// implementation.

package stats2

import (
	"time"

	"github.com/google/working-instrumentation-go/tags"
)

// view is the data structure that holds the info describing the view as well
// as the aggregated data.
type view struct {
	// name of View. Must be unique.
	name        string
	description string

	// tagKeys to perform the aggregation on.
	tagKeys []tags.Key

	// measureName is the name of a Measure. Examples are cpu:tickCount,
	// diskio:time...
	measureName string

	// aggregation is the aggregation to perform for this view.
	aggregation Aggregation

	// window is the window under which the aggregation is performed.
	window Window

	// start is time when view collection was started originally.
	start time.Time

	// vChans are the channels through which the collected views data for this
	// view are sent to the consumers of this view.
	vChans map[chan *View]struct{}

	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its *stats.Aggregator.
	signatures map[string]Aggregation
}
