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
	"time"

	"github.com/google/instrumentation-go/stats/tagging"

	"golang.org/x/net/context"
)

// RegisterMeasureDesc adds a measurement descriptor a.k.a resource to the list
// of descriptors known by the stats library so that usage of that resource may
// be recorded by calling RecordUsage. RegisterMeasureDesc returns an error if
// a descriptor with the same name was already registered. Statistics for this
// descriptor will be reported only for views that were registered using the
// descriptor name.
var RegisterMeasureDesc func(md *measureDesc) error

// UnregisterMeasureDesc deletes a previously registered measureDesc with the
// same mName. It returns an error if no registered mName can be found with the
// same name or if AggregationViewDesc referring to it is still registered.
var UnregisterMeasureDesc func(mName string) error

// RegisterViewDesc registers an AggregationViewDesc. It returns an error if
// the AggregationViewDesc cannot be registered.
// Subsequent calls to RecordUsage with a measureDesc and tags that match a
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

// UnregisterViewDesc deletes a previously registered AggregationViewDesc with
// the same vwName. It returns an error if no registered AggregationViewDesc
// can be found with the same name. All data collected and not reported for the
// corresponding view will be lost. All clients subscribed to this view are
// unsubscribed automatically and their subscriptions channels closed.
var UnregisterViewDesc func(vwName string) error

// SubscribeToView subscribes a client to an already registered
// AggregationViewDesc. It allows for many clients to consume the same View
// with a single registration. It returns an error if no registered
// AggregationViewDesc can be found with the same name.
var SubscribeToView func(vwName string, c chan *View) error

// UnsubscribeFromView unsubscribes a previously subscribed channel from the
// AggregationViewDesc subscriptions.
// It returns an error if no AggregationViewDesc with name vwName is found or
// if c is not subscribed to it.
var UnsubscribeFromView func(vwName string, c chan *View) error

// RecordMeasurement records a quantity of usage of the specified measureDesc.
// Tags are passed as part of the context.
// TODO(iamm2): Expand the API to allow passing the tags explicitly in the
// function call to avoid creating a new context with the new tags that will be
// disregarded right away. This is not optimal as for each record we need to
// take a lock. Extracting the tags from the context and assigning them to
// views is expensive and performing this for each record is not ideal. This is
// intentional to keep the API simple for the first version.
var RecordMeasurement func(ctx context.Context, m Measurement)

// RecordMeasurements records multiple measurements with the same tags at once.
// It is expected that mds and values are the same length. If not, none of the
// measurements are recorded.
var RecordMeasurements func(ctx context.Context, m ...Measurement)

// SetCallbackPeriod sets the minimum and maximum periods for aggregation
// reporting for all registered views in the program. The maximum period is
// only advisory; reports may be generated less frequently than this.
// The default period is determined by internal memory usage.  Calling
// SetCallbackPeriod with either argument equal to zero re-enables the default
// behavior.
var SetCallbackPeriod func(min, max time.Duration)

func init() {
	uc := newUsageCollector()
	RecordMeasurements = func(ctx context.Context, m ...Measurement) {
		ts := tagging.FromContext(ctx)
		uc.recordManyMeasurement(time.Now(), ts, m)
	}

	RetrieveView = func(name string) []*View {
		return uc.retrieveViews(time.Now())
	}
}

var RetrieveView func(name string) []*View
