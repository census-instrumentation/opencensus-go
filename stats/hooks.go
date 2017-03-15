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

	"golang.org/x/net/context"
)

// RegisterMeasureDesc adds a measurement descriptor a.k.a resource to the list
// of descriptors known by the stats library so that usage of that resource may
// be recorded by calling RecordUsage. RegisterMeasureDesc returns an error if
// a descriptor with the same name was already registered. Statistics for this
// descriptor will be reported only for views that were registered using the
// descriptor name.
var RegisterMeasureDesc func(md MeasureDesc) error

// UnregisterMeasureDesc deletes a previously registered measureDesc with the
// same mName. It returns an error if no registered mName can be found with the
// same name or if AggregationViewDesc referring to it is still registered.
var UnregisterMeasureDesc func(mName string) error

// RegisterViewDesc registers an AggregationViewDesc. It returns an error if
// the AggregationViewDesc cannot be registered.
// Subsequent calls to RecordUsage with a measureDesc and tags that match a
// AggregationViewDesc will cause the usage to be recorded.
var RegisterViewDesc func(vd ViewDesc) error

// UnregisterViewDesc deletes a previously registered AggregationViewDesc with
// the same vwName. It returns an error if no registered AggregationViewDesc
// can be found with the same name. All data collected and not reported for the
// corresponding view will be lost. All clients subscribed to this view are
// unsubscribed automatically and their subscriptions channels closed.
var UnregisterViewDesc func(vwName string) error

// Subscribe subscribes a client to an already registered ViewDesc or a set of
// ViewDesc. It allows for many clients to consume the same collected View(s).
// It returns an error if the subscription was already used to subscribe. If
// the subscription is successful, the channel within hte subscription is used
// to subscribe to the collected view(s) -i.e. the collected  measurements for
// the registered AggregationViewDesc will be reported to the client through
// channel c. To avoid data loss, clients must ensure that channel sends
// proceed in a timely manner. The calling code is responsible for using a
// buffered channel or blocking on the channel waiting for the collected view.
// Limits on the aggregation period can be set by SetCallbackPeriod.
var Subscribe func(s Subscription) error

// Unsubscribe removes a previously subscribed subscription. It returns an
// error if the subscription wasn't used previously to subscribe.
var Unsubscribe func(s Subscription) error

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

// RetrieveViews allows retrieving views in an adhoc fashion. This retrieval
// doesn't reset the view's collected data, and just returns a snapshot of the
// view as it is currently collected by the library.
// TODO(mmoakil): implement this.
var RetrieveViews func(viewNames, measureNames []string) ([]*View, error)
