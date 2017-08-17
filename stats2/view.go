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

// View is the generic interface defining the various type of views.
type View interface {
	isView() bool
	Name() string
}

// ViewFloat64 is the data structure that holds the info describing the float64
// view as well as the aggregated data.
type ViewFloat64 struct {
	// name of View. Must be unique.
	name        string
	description string

	// tagKeys to perform the aggregation on.
	tagKeys []tags.Key

	// Examples of measures are cpu:tickCount, diskio:time...
	measure MeasureFloat64

	// aggregation is the description of the aggregation to perform for this
	// view.
	aggregation AggregationFloat64

	// window is the window under which the aggregation is performed.
	window Window

	// start is time when view collection was started originally.
	start time.Time

	// vChans are the channels through which the collected views data for this
	// view are sent to the consumers of this view.
	vChans map[chan *ViewData]struct{}

	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its AggregateValueFloat64.
	signatures map[string]AggregateValueFloat64
}

func (v *ViewFloat64) recordFloat64(ts *tags.TagSet, f float64) {

}

// ViewInt64 is the data structure that holds the info describing the int64
// view as well as the aggregated data.
type ViewInt64 struct {
	// name of View. Must be unique.
	name        string
	description string

	// tagKeys to perform the aggregation on.
	tagKeys []tags.Key

	// Examples of measures are cpu:tickCount, diskio:time...
	measure MeasureInt64

	// aggregation is the description of the aggregation to perform for this
	// view.
	aggregation AggregationInt64

	// window is the window under which the aggregation is performed.
	window Window

	// start is time when view collection was started originally.
	start time.Time

	// vChans are the channels through which the collected views data for this
	// view are sent to the consumers of this view.
	vChans map[chan *ViewData]struct{}

	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its AggregateValueInt64.
	signatures map[string]AggregateValueInt64
}

func (v *ViewInt64) recordInt64(ts *tags.TagSet, i int64) {

}

// A ViewData is a set of rows about usage of the single measure associated
// with the given view during a particular window. Each row is specific to a
// unique set of tags.
type ViewData struct {
	v    View
	rows []*Rows
}

// NewViewFloat64 creates a new *ViewFloat64.
func NewViewFloat64(name, description string, keys []tags.Key, measure MeasureFloat64, agg AggregationFloat64, wnd Window) (*ViewFloat64, error) {
	return nil, nil
}

// NewViewInt64 creates a new *ViewInt64.
func NewViewInt64(name, description string, keys []tags.Key, measure MeasureInt64, agg AggregationInt64, wnd Window) (*ViewInt64, error) {
	return nil, nil
}
