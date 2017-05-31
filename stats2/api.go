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

// RegisterViewDesc registers an AggregationViewDesc. It returns an error if
// the AggregationViewDesc cannot be registered.
// Subsequent calls to RecordUsage with tags that match a AggregationViewDesc
// will cause the usage to be recorded. If the registration is successful, the
// channel is used to subscribe to the view -i.e. the collected measurements
// for the registered AggregationViewDesc will be reported to the client
// through channel c. Data in the channel is differential, meaning the returned
// value is the aggregation of collected data for that view since the last
// report. To avoid data loss, clients must ensure that channel sends proceed
// in a timely manner. The calling code is responsible for using a buffered
// channel for anything else than blocking on the channel waiting for the
// collected view. Limits on the aggregation period can be set by
// SetCallbackPeriod.
var RegisterViewDesc func(vd *ViewDesc, c chan *View) error

// UnregisterViewDesc deletes a previously registered viewDesc with the same
// vwName. It returns an error if no registered ViewDesc can be found with the
// same name. All data collected and not reported for the corresponding view
// will be lost. All clients subscribed to this view are unsubscribed
// automatically and their subscriptions channels closed.
var UnregisterViewDesc func(vwName string) error

// SubscribeToView subscribes a client to an already registered ViewDesc. It
// allows for many clients to consume the same View with a single registration.
// It returns an error if no registered ViewDesc can be found with the
// same name.
var SubscribeToView func(vwName string, c chan *View) error

// UnsubscribeFromView unsubscribes a previously subscribed channel from the
// ViewDesc subscriptions.
// It returns an error if no ViewDesc with name vwName is found or if c is
// not subscribed to it.
var UnsubscribeFromView func(vwName string, c chan *View) error

// RecordMeasurement records a measurement against the tags passed as part of
// the context.
var RecordMeasurement func(ctx context.Context, m *Measurement)

// RecordManyMeasurement records multiple measurements with the same tags at
// once.
var RecordManyMeasurement func(ctx context.Context, ms []*Measurement)

// SetCallbackPeriod sets the minimum and maximum periods for aggregation
// reporting for all registered views in the program. The maximum period is
// only advisory; reports may be generated less frequently than this.
// The default period is determined by internal memory usage.  Calling
// SetCallbackPeriod with either argument equal to zero re-enables the default
// behavior.
var SetCallbackPeriod func(min, max time.Duration)

// CreateMeasureDescfloat64 creates a float64 measure descriptor. The name must
// be unique. Used to link the MeasureDescFloat64 to a ViewDesc. Examples are:
// cpu:tickCount, diskio:time...
// The description is used for display purposes only. It is meant to be human
// readable and is used to show the resource in dashboards. Example are:
// CPU profile ticks, Disk I/O, Disk usage in usecs...
var func RegisterMeasureDescFloat64(name string, description string) (MeasureDescFloat64, error)

// RegisterMeasureDescInt64 creates an int64 measure descriptor. The name must
// be unique. Used to link the MeasureDescInt64 to a ViewDesc. Examples are:
// cpu:tickCount, diskio:time...
// The description is used for display purposes only. It is meant to be human
// readable and is used to show the resource in dashboards. Example are:
// CPU profile ticks, Disk I/O, Disk usage in usecs...
var func RegisterMeasureDescInt64(name string, description string) MeasureDescFloat64

// UnregisterMeasureDesc deletes a previously registered MeasureDesc with the
// same mName. It returns an error if no registered mName can be found with the
// same name or if ViewDesc referring to it is still registered.
var func UnRegisterMeasureDesc(name string) error

var func CreateMeasurementFloat64(mdf MeasureDescFloat64, f float64) Measurement

var func CreateMeasurementInt64(mdi MeasureDescInt64, i int64) Measurement



type MeasureDesc interface {
	isMeasureDesc() bool
}

type MeasureDescFloat64 interface {
	MeasureDesc
	CreateMeasurement(v float64) Measurement
}

type MeasureDescInt64 interface {
	MeasureDesc
	CreateMeasurement(v int64) Measurement
}

type Measurement interface {
	record(ctx context.Context)
}

func CreateViewDesc(name string, description string, tagKeys []string, measureDescName string, aggregationDesc AggregationDesc, windowDesc WindowDesc) ViewDesc {
	return nil
}

// ViewDesc is a helper data structure that holds common fields to all
// ViewAggregationDesc. It should never be used standalone but always as part
// of a ViewAggregationDesc.
type ViewDesc interface {
	IsViewDesc() bool
	Name() string
	TagKeys() []tags.Key
	MeasureDescName() string
	AggregationDesc() AggregationDesc
	WindowDesc() WindowDesc
}

type AggregationDesc interface {
	IsAggregationDesc() bool
}

type WindowDesc interface {
	IsWindowDesc() bool
}

type WindowDescCumulative struct {
}

type WindowDescSlidingTime struct {
}

type WindowDescJumpingTime struct {
}

/* TODO(acetechnologist): add support for other types: slidingSpace,
//jumpingSpace.
type WindowDescSlidingSpace struct {
}

type WindowDescJumpingSpace struct {
}
*/

type Aggregation interface {
	IsAggregation() bool
	Tags() []tags.Tag
}

type AggContinuousStatsFloat64 struct {
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

type AggGaugeStatsFloat64 struct {
	Value float64
	tags  []tags.Tag
}

// A View is a set of Aggregations about usage of the single resource
// associated with the given view during a particular time interval. Each
// Aggregation is specific to a unique set of tags. The Census infrastructure
// reports a stream of View events to the application for further processing
// such as further aggregations, logging and export to other services.
type View struct {
	ViewDesc ViewDesc
	// Aggregations is expected to be a []*AggContinuousStatsFloat64,
	// []*AggContinuousStatsInt64a, []*AggGaugeStatsFloat64,
	// []*AggGaugeStatsInt64, []*AggGaugeStatsBool or []*AggGaugeStatsString.
	Aggregations interface{}
}
