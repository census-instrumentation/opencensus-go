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
package stats

import (
	"bytes"
	"fmt"
	"time"

	"golang.org/x/net/context"
)

// RegisterMeasureDesc adds a measurement descriptor a.k.a resource to the list
// of descriptors known by the stats library so that usage of that resource may
// be recorded by calling RecordUsage. RegisterMeasureDesc returns an error if
// a descriptor with the same name was already registered. Statistics for this
// descriptor will be reported only for views that were registered using the
// descriptor name.
var RegisterMeasureDesc func(md *MeasureDesc) error

// UnregisterMeasureDesc deletes a previously registered MeasureDesc with the
// same mName. It returns an error if no registered mName can be found with the
// same name or if ViewDesc referring to it is still registered.
var UnregisterMeasureDesc func(mName string) error

// RegisterViewDesc registers an AggregationViewDesc. It returns an error if
// the AggregationViewDesc cannot be registered.
// Subsequent calls to RecordUsage with a MeasureDesc and tags that match a
// AggregationViewDesc will cause the usage to be recorded. If the registration
// is successful, the channel is used to subscribe to the view -i.e. the
// collected measurements for the registered AggregationViewDesc will be
// reported to the client through channel c. Data in the channel is
// differential, meaning the returned value is the aggregation of collected
// data for that view since the last report. To avoid data loss, clients must
// ensure that channel sends proceed in a timely manner. The calling code is
// responsible for using a buffered channel for anything else than blocking on
// the channel waiting for the collected view. Limits on the aggregation period
// can be set by SetCallbackPeriod.
var RegisterViewDesc func(vd AggregationViewDesc, c chan *View) error

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

// RecordMeasurement records a quantity of usage of the specified MeasureDesc.
// Tags are passed as part of the context.
// TODO(iamm2): Expand the API to allow passing the tags explicitly in the
// function call to avoid creating a new context with the new tags that will be
// disregarded right away. This is not optimal as for each record we need to
// take a lock. Extracting the tags from the context and assigning them to
// views is expensive and performing this for each record is not ideal. This is
// intentional to keep the API simple for the first version.
var RecordMeasurement func(ctx context.Context, md *MeasureDesc, value float64)

// RecordManyMeasurement records multiple measurements with the same tags at
// once. It is expected that mds and values are the same length. If not, none
// of the measurements are recorded.
var RecordManyMeasurement func(ctx context.Context, mds []*MeasureDesc, values []float64)

// SetCallbackPeriod sets the minimum and maximum periods for aggregation
// reporting for all registered views in the program. The maximum period is
// only advisory; reports may be generated less frequently than this.
// The default period is determined by internal memory usage.  Calling
// SetCallbackPeriod with either argument equal to zero re-enables the default
// behavior.
var SetCallbackPeriod func(min, max time.Duration)

// ViewDesc is a helper data structure that holds common fields to all
// ViewAggregationDesc. It should never be used standalone but always as part
// of a ViewAggregationDesc.
type ViewDesc struct {
	// Name of ViewDesc. Must be unique.
	// TODO(iamm2): provide examples for Name.
	Name string
	// TODO(iamm2): provide an example for description.
	Description string

	// MeasureDescName is the name of a Measure. Examples are cpu:tickCount,
	// diskio:time...
	MeasureDescName string

	// Keys to perform the aggregation on.
	TagKeys []string

	// start is time when ViewDesc was registered.
	start time.Time

	// vChans are the channels through which the collected views for this ViewDesc
	// are sent to the consumers of this view.
	vChans map[chan *View]struct{}

	// signatures holds the aggregations for each unique tag signature (values
	// for all keys) to its *stats.Aggregator.
	signatures map[string]aggregator
}

// aggregator is the interface that the aggregators created by an aggregation
// are expected to implement.
type aggregator interface {
	addSample(v float64, t time.Time)
}

// AggregationViewDesc is the interface that all aggregations are expected to
// implement.
type AggregationViewDesc interface {
	// creates an aggregator instance for a unique tags signature.
	createAggregator(t time.Time) (aggregator, error)
	// retrieves the collected *View holding collected data by all the
	// aggregator instances..
	retrieveView(now time.Time) (*View, error)
	// returns the *ViewDesc associated with this AggregationViewDesc
	viewDesc() *ViewDesc
	// validates the input recieved as requested by the client code.
	isValid() error
}

// DistributionAggViewDesc holds the parameters describing an aggregation
// distribution..
type DistributionAggViewDesc struct {
	*ViewDesc

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

func (dd *DistributionAggViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newDistributionAggregator(dd.Bounds), nil
}

func (dd *DistributionAggViewDesc) retrieveView(now time.Time) (*View, error) {
	dav, err := dd.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		ViewDesc: dd.ViewDesc,
		ViewAgg:  dav,
	}, nil
}

func (dd *DistributionAggViewDesc) viewDesc() *ViewDesc {
	return dd.ViewDesc
}

func (dd *DistributionAggViewDesc) isValid() error {
	for i := 1; i < len(dd.Bounds); i++ {
		if dd.Bounds[i-1] >= dd.Bounds[i] {
			return fmt.Errorf("%v error. bounds are not increasing", dd)
		}
	}
	return nil
}

func (dd *DistributionAggViewDesc) retrieveAggreationView(t time.Time) (*DistributionAggView, error) {
	var aggs []*DistributionAgg

	for sig, a := range dd.signatures {
		tags, err := tagsFromSignature([]byte(sig), dd.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature %v", sig)
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

	return &DistributionAggView{
		Descriptor:   dd,
		Aggregations: aggs,
		Start:        dd.start,
		End:          t,
	}, nil
}

// IntervalAggViewDesc holds the parameters describing an interval aggregation.
type IntervalAggViewDesc struct {
	*ViewDesc

	// Number of internal sub-intervals to use when collecting stats for each
	// interval. The max error in interval measurements will be approximately
	// 1/SubIntervals (although in practice, this will only be approached in
	// the presence of very large and bursty workload changes), and underlying
	// memory usage will be roughly proportional to the value of this
	// field. Must be in the range [2, 20]. A value of 5 will be used if this
	// is unspecified.
	SubIntervals int

	// The size of each interval, as a time duration. Must have at least one
	// element.
	Intervals []time.Duration
}

func (id *IntervalAggViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newIntervalsAggregator(t, id.Intervals, id.SubIntervals), nil
}

func (id *IntervalAggViewDesc) retrieveView(now time.Time) (*View, error) {
	iav, err := id.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		ViewDesc: id.ViewDesc,
		ViewAgg:  iav,
	}, nil
}

func (id *IntervalAggViewDesc) viewDesc() *ViewDesc {
	return id.ViewDesc
}

func (id *IntervalAggViewDesc) isValid() error {
	if id.SubIntervals < 2 || id.SubIntervals < 20 {
		return fmt.Errorf("%v error. subIntervals is not in [2,20]", id)
	}
	return nil
}

func (id *IntervalAggViewDesc) retrieveAggreationView(now time.Time) (*IntervalAggView, error) {
	var aggs []*IntervalAgg

	for sig, a := range id.signatures {
		tags, err := tagsFromSignature([]byte(sig), id.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature %v", sig)
		}
		aggregator, ok := a.(*intervalsAggregator)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregator type. got %T, want stats.intervalsAggregator", a)
		}
		ia := &IntervalAgg{
			IntervalStats: aggregator.retrieveCollected(now),
			Tags:          tags,
		}
		aggs = append(aggs, ia)
	}

	return &IntervalAggView{
		Descriptor:   id,
		Aggregations: aggs,
	}, nil
}

// A View is a set of Aggregations about usage of the single resource
// associated with the given view during a particular time interval. Each
// Aggregation is specific to a unique set of tags. The Census infrastructure
// reports a stream of View events to the application for further processing
// such as further aggregations, logging and export to other services.
type View struct {
	ViewDesc *ViewDesc
	// ViewAgg is expected to be a *DistributionAggView or a
	// *IntervalAggView
	ViewAgg interface{}
}

// DistributionAggView is the set of collected DistributionAgg associated with
// ViewDesc.
type DistributionAggView struct {
	Descriptor   *DistributionAggViewDesc
	Aggregations []*DistributionAgg
	Start, End   time.Time // start is time when ViewDesc was registered.
}

// An DistributionAgg is a statistical summary of measures associated with a
// unique tag set for a specific bucket.
type DistributionAgg struct {
	*DistributionStats
	Tags []Tag
}

// DistributionStats records a distribution of float64 sample values.
// It is the result of a DistributionAgg aggregation.
type DistributionStats struct {
	Count               int64
	Min, Mean, Max, Sum float64
	// CountPerBucket is the set of occurrences count per bucket. The
	// buckets bounds are the same as the ones setup in
	// AggregationDesc.
	CountPerBucket []int64
}

// IntervalAggView is the set of collected IntervalAgg associated with
// ViewDesc.
type IntervalAggView struct {
	Descriptor   *IntervalAggViewDesc
	Aggregations []*IntervalAgg
}

// IntervalAgg is a statistical summary of measures associated with a unique
// tag set for a specific time interval.
type IntervalAgg struct {
	IntervalStats []*IntervalStats
	Tags          []Tag
}

// IntervalStats records stats result of an IntervalAgg aggregation for a
// specific time window.
type IntervalStats struct {
	Duration   time.Duration
	Count, Sum float64
}

// A Tag is the (key,value) pair that the client code uses to tag a
// measurement.
type Tag struct {
	Key, Value string
}

// MeasureDesc describes a data point (measurement) type accounted
// for by the stats library, such as RAM or CPU time.
type MeasureDesc struct {
	// The name must be unique. Used to link the MeasureDesc to a ViewDesc.
	// Examples are cpu:tickCount, diskio:time...
	Name string
	// The description is used for display purposes only. It is meant to be
	// human readable and is used to show the resource in dashboards.
	// Example are CPU profile ticks, Disk I/O, Disk usage in usecs...
	Description string
	Unit        MeasurementUnit

	aggViewDescs map[AggregationViewDesc]struct{}
}

// MeasurementUnit is the unit of measurement for a resource.
type MeasurementUnit struct {
	Power10      int
	Numerators   []BasicUnit
	Denominators []BasicUnit
}

// BasicUnit is used for representing the basic units used to construct
// MeasurementUnits.
type BasicUnit byte

// These constants are the type of basic units allowed.
const (
	UnknownUnit BasicUnit = iota
	ScalarUnit
	BitsUnit
	BytesUnit
	SecsUnit
	CoresUnit
)

func (vw *View) String() string {
	if vw == nil {
		return "nil"
	}
	var buf bytes.Buffer
	buf.WriteString("View{\n")
	fmt.Fprintf(&buf, "%v,\n", vw.ViewDesc)
	fmt.Fprintf(&buf, "%v,\n", vw.ViewAgg)
	buf.WriteString("}")
	return buf.String()
}

func (dd *DistributionAggViewDesc) String() string {
	if dd == nil {
		return "nil"
	}
	vd := dd.ViewDesc
	var buf bytes.Buffer
	buf.WriteString("  viewDesc{\n")
	fmt.Fprintf(&buf, "    Name: %v,\n", vd.Name)
	fmt.Fprintf(&buf, "    Description: %v,\n", vd.Description)
	fmt.Fprintf(&buf, "    MeasureDescName: %v,\n", vd.MeasureDescName)
	fmt.Fprintf(&buf, "    TagKeys: %v,\n", vd.TagKeys)
	fmt.Fprintf(&buf, "    Bound: %v,\n", dd.Bounds)
	buf.WriteString("    },\n")
	buf.WriteString("  }")
	return buf.String()
}

func (id *IntervalAggViewDesc) String() string {
	if id == nil {
		return "nil"
	}
	vd := id.ViewDesc
	var buf bytes.Buffer
	buf.WriteString("  viewDesc{\n")
	fmt.Fprintf(&buf, "    Name: %v,\n", vd.Name)
	fmt.Fprintf(&buf, "    Description: %v,\n", vd.Description)
	fmt.Fprintf(&buf, "    MeasureDescName: %v,\n", vd.MeasureDescName)
	fmt.Fprintf(&buf, "    TagKeys: %v,\n", vd.TagKeys)
	fmt.Fprintf(&buf, "    Intervals: %v,\n", id.Intervals)
	buf.WriteString("    },\n")
	buf.WriteString("  }")
	return buf.String()
}

func (dd *DistributionAggView) String() string {
	if dd == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  viewAgg{\n")
	fmt.Fprintf(&buf, "    Start: %v,\n", dd.Start)
	fmt.Fprintf(&buf, "    End: %v,\n", dd.End)
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", dd.Aggregations)
	buf.WriteString("  }")
	return buf.String()
}

func (ia *IntervalAggView) String() string {
	if ia == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  viewAgg{\n")
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", ia.Aggregations)
	buf.WriteString("  }")
	return buf.String()
}

func (da *DistributionAgg) String() string {
	if da == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  DistributionAgg{\n")
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", da.DistributionStats)
	fmt.Fprintf(&buf, "    Tags: %v,\n", da.Tags)
	buf.WriteString("  }")
	return buf.String()
}

func (ia *IntervalAgg) String() string {
	if ia == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  IntervalAgg{\n")
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", ia.IntervalStats)
	fmt.Fprintf(&buf, "    Tags: %v,\n", ia.Tags)
	buf.WriteString("  }")
	return buf.String()
}

func (ds *DistributionStats) String() string {
	if ds == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  DistributionStats{\n")
	fmt.Fprintf(&buf, "    Count: %v,\n", ds.Count)
	fmt.Fprintf(&buf, "    Min: %v,\n", ds.Min)
	fmt.Fprintf(&buf, "    Mean: %v,\n", ds.Mean)
	fmt.Fprintf(&buf, "    Max: %v,\n", ds.Max)
	fmt.Fprintf(&buf, "    Sum: %v,\n", ds.Sum)
	fmt.Fprintf(&buf, "    CountPerBucket: %v,\n", ds.CountPerBucket)
	buf.WriteString("  }")
	return buf.String()
}

func (is *IntervalStats) String() string {
	if is == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  DistributionStats{\n")
	fmt.Fprintf(&buf, "    Duration: %v,\n", is.Duration)
	fmt.Fprintf(&buf, "    Count: %v,\n", is.Count)
	fmt.Fprintf(&buf, "    Sum: %v,\n", is.Sum)
	buf.WriteString("  }")
	return buf.String()
}
