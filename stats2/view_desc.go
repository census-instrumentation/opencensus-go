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

// viewDesc is the data structure that holds the info describing the view as
// well as the aggregated data.
type viewDesc struct {
	// name of ViewDesc. Must be unique.
	name        string
	description string

	// tagKeys to perform the aggregation on.
	tagKeys []tags.Key

	// measureDescName is the name of a Measure. Examples are cpu:tickCount,
	// diskio:time...
	measureDescName string

	// AggregationDesc is an aggregation description.
	aggregationDesc AggregationDesc

	// WindowDesc is an aggregation window description.
	windowDesc WindowDesc

	// start is time when ViewDesc was registered.
	start time.Time

	// vChans are the channels through which the collected views for this ViewDesc
	// are sent to the consumers of this view.
	vChans map[chan *View]struct{}

	// signatures holds the aggregations for each unique tag signature (values
	// for all keys) to its *stats.Aggregator.
	signatures map[string]Aggregation
}
