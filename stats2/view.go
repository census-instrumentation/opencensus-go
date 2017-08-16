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

// View is the data structure that holds the info describing the view as well
// as the aggregated data.
type View struct {
	// name of View. Must be unique.
	name        string
	description string

	// tagKeys to perform the aggregation on.
	tagKeys []tags.Key

	// Examples of measures are cpu:tickCount, diskio:time...
	measure Measure

	// aggregation is the description of the aggregation to perform for this
	// view.
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

// A View is a set of Aggregations about usage of the single resource
// associated with the given view during a particular time interval. Each
// Aggregation is specific to a unique set of tags. The Census infrastructure
// reports a stream of View events to the application for further processing
// such as further aggregations, logging and export to other services.
type ViewData struct {
	v *View

	// Aggregations is expected to be a []*AggContinuousStatsFloat64,
	// []*AggContinuousStatsInt64a, []*AggGaugeStatsFloat64,
	// []*AggGaugeStatsInt64, []*AggGaugeStatsBool or []*AggGaugeStatsString.
	rows Rows
}

// NewView creates a new *view.
func NewView(name, description, string, keys []tags.Key, measure Measure, agg Aggregation, w Window) (*View, error) {
	// TODO
	return nil, nil
}

// RegisterView registers view. It returns an error if the view cannot be
// registered. Subsequent calls to Record with the same measure as the one in
// the view will NOT cause the usage to be recorded unless a consumer is
// subscribed to the view or StartCollectionForAdhoc for this view is called.
func RegisterView(v *View) error {

}

// UnregisterView deletes the previously registered view. It returns an error
// if no registered View can be found with the same name. All data collected
// and not reported for the corresponding view will be lost. All clients
// subscribed to this view are unsubscribed automatically and their
// subscriptions channels closed.
func UnregisterView(v *View) error {

}

// GetViewByName returns the registered view associated with this name.
func GetViewByName(name string) (*View, error) {

}

// SubscribeToView subscribes a client to a View. If the view wasn't already
// registered, it will be automatically registered. It allows for many clients
// to consume the same ViewData with a single registration. -i.e. the aggregate
// of the collected measurements will be reported to the calling code through
// channel c. To avoid data loss, clients must ensure that channel sends
// proceed in a timely manner. The calling code is responsible for using a
// buffered channel or blocking on the channel waiting for the collected data.
func SubscribeToView(v *View, c chan *ViewData) error {
}

// UnsubscribeFromView unsubscribes a previously subscribed channel from the
// View subscriptions. If no more subscriber for v exists and the the ad hoc
// collection for this view isn't active, data stops being collected for this
// view.
func UnsubscribeFromView(v *View, c chan *ViewData) error {
}

// StartCollectionForAdhoc starts data collection for this view even if no
// listeners are subscribed to it.
func StartCollectionForAdhoc(v *View) error {
}

// StopCollectionForAdhoc stops data collection for this view unless at least
// 1 listener is subscribed to it.
func StopCollectionForAdhoc(v *View) error {
}

// RetrieveData returns the current collected data for the view.
func RetrieveData func(v *View) (*ViewData, error) {

}