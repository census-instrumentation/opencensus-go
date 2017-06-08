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
	"context"
	"time"

	"github.com/google/working-instrumentation-go/tags"
)

// CreateMeasureFloat64 creates a float64 measure descriptor. The name must be
// unique. Used to link the MeasureFloat64 to a View. Examples are:
// cpu:tickCount, diskio:time...
// The description is used for display purposes only. It is meant to be human
// readable and is used to show the resource in dashboards. Example are:
// CPU profile ticks, Disk I/O, Disk usage in usecs...
var CreateMeasureFloat64 func(name string, description string) (MeasureFloat64, error)

// CreateMeasureInt64 creates an int64 measure descriptor. The name must be
// unique. Used to link the MeasureInt64 to a View. Examples are:
// cpu:tickCount, diskio:time...
// The description is used for display purposes only. It is meant to be human
// readable and is used to show the resource in dashboards. Example are:
// CPU profile ticks, Disk I/O, Disk usage in usecs...
var CreateMeasureInt64 func(name string, description string) (MeasureInt64, error)

// UnRegisterMeasure deletes a previously registered Measure with the same
// mName. It returns an error if no registered mName can be found with the same
// name or if View referring to it is still registered.
var UnRegisterMeasure func(name string) error

// RecordFloat64 records a float64 value against a measure and the tags passed
// as part of the context.
var RecordFloat64 func(ctx context.Context, mf MeasureFloat64, v float64)

// RecordInt64 records an int64 value against a measure and the tags passed as
// part of the context.
var RecordInt64 func(ctx context.Context, mf MeasureInt64, v int64)

// Record records one or multiple measurements with the same tags at once.
var Record func(ctx context.Context, ms []*Measurement)

// Measure is the interface for all measure types. A measure is required when
// defining a view.
type Measure interface {
	isMeasure() bool
}

// MeasureFloat64 is the interface for measureFloat64.
type MeasureFloat64 interface {
	Measure
	Is(v float64) Measurement
}

// MeasureInt64 is the interface for measureInt64.
type MeasureInt64 interface {
	Measure
	Is(v int64) Measurement
}

// Measurement is the interface for all measurement types. Measurements are
// required when recording stats.
type Measurement interface {
	record(ctx context.Context)
}

// RegisterView registers view. It returns an error if the view cannot be
// registered. Subsequent calls to Record with the same measure as the one in
// the view will cause the usage to be recorded. If the registration is
// successful, the channel is used to subscribe to the view -i.e. the collected
// measurements for the registered AggregationView will be reported to the
// client through channel c. Data in the channel is differential, meaning the
// returned value is the aggregation of collected data for that view since the
// last report. To avoid data loss, clients must ensure that channel sends
// proceed in a timely manner. The calling code is responsible for using a
// buffered channel for anything else than blocking on the channel waiting for
// the collected view. Limits on the aggregation period can be set by
// SetCallbackPeriod.
var RegisterView func(vwName, description, string, tagKeys []tags.Key, measureName string, agg Aggregation, wnd Window, c chan *View) error

// UnregisterView deletes a previously registered view with the same vwName. It
// returns an error if no registered View can be found with the same name. All
// data collected and not reported for the corresponding view will be lost. All
// clients subscribed to this view are unsubscribed automatically and their
// subscriptions channels closed.
var UnregisterView func(vwName string) error

// SubscribeToView subscribes a client to an already registered View. It allows
// for many clients to consume the same View with a single registration. It
// returns an error if no registered View can be found with the same name.
var SubscribeToView func(vwName string, c chan *View) error

// UnsubscribeFromView unsubscribes a previously subscribed channel from the
// View subscriptions. It returns an error if no View with name vwName is found
// or if c is not subscribed to it.
var UnsubscribeFromView func(vwName string, c chan *View) error

// SetCallbackPeriod sets the minimum and maximum periods for aggregation
// reporting for all registered views in the program. The maximum period is
// only advisory; reports may be generated less frequently than this. The
// default period is determined by internal memory usage.  Calling
// SetCallbackPeriod with either argument equal to zero re-enables the default
// behavior.
var SetCallbackPeriod func(min, max time.Duration)

type Aggregation interface {
	isAggregation() bool
}

type Window interface {
	isWindow() bool
}

type WindowCumulative struct {
}

type WindowSlidingTime struct {
}

type WindowJumpingTime struct {
}

/* TODO(acetechnologist): add support for other types: slidingSpace,
//jumpingSpace.
type WindowSlidingSpace struct {
}

type WindowJumpingSpace struct {
}
*/

type AggregationValue interface {
	isAggregationValue() bool
	Tags() []tags.Tag
}

type AggValueContinuousStatsFloat64 struct {
	Count         int64
	Min, Max, Sum float64
	// The sum of squared deviations from the mean of the values in the
	// population. For values x_i this is:
	//
	//     Sum[i=1..n]((x_i - mean)^2)
	//
	// Knuth, "The Art of Computer Programming", Vol. 2, page 323, 3rd edition
	// describes Welford's method for accumulating this sum in one pass.
	SumOfSquaredDeviation float64
	// CountPerBucket is the set of occurrences count per bucket. The
	// buckets bounds are the same as the ones setup in
	// AggregationDesc.
	CountPerBucket []int64
	tags           []tags.Tag
}

type AggValueGaugeStatsFloat64 struct {
	Value float64
	tags  []tags.Tag
}

// A View is a set of Aggregations about usage of the single resource
// associated with the given view during a particular time interval. Each
// Aggregation is specific to a unique set of tags. The Census infrastructure
// reports a stream of View events to the application for further processing
// such as further aggregations, logging and export to other services.
type View struct {
	vwName, description string
	tagKeys             []tags.Key
	measure             Measure
	agg                 Aggregation
	wnd                 Window
	c                   chan *View

	// Aggregations is expected to be a []*AggContinuousStatsFloat64,
	// []*AggContinuousStatsInt64a, []*AggGaugeStatsFloat64,
	// []*AggGaugeStatsInt64, []*AggGaugeStatsBool or []*AggGaugeStatsString.
	aggregationValues interface{}
}
