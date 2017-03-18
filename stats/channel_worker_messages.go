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
	"time"

	"github.com/google/instrumentation-go/stats/tagging"
)

// measureDescRegistration is a message requesting that the channelWorker
// goroutine registers MeasureDesc md.
type measureDescRegistration struct {
	md  MeasureDesc
	err chan error
}

// measureDescUnregistration is a message requesting that the channelWorker
// goroutine unregisters MeasureDesc md.
type measureDescUnregistration struct {
	mn  string
	err chan error
}

// viewDescRegistration is a message requesting that the channelWorker
// goroutine registers ViewDesc vd.
type viewDescRegistration struct {
	vd  ViewDesc
	err chan error
}

// viewDescUnregistration is a message requesting that the channelWorker
// goroutine unregisters ViewDesc vd.
type viewDescUnregistration struct {
	vn  string
	err chan error
}

// retrieveViewsQuery is a message requesting that channelWorker goroutine
// returns matching views.
type retrieveViewsQuery struct {
	vns, mns []string
	vws      chan []*View
	err      chan error
}

// viewDescSubscription is a message requesting that the channelWorker
// goroutine subscribes the caller to the view named vn.
type viewDescSubscription struct {
	s   Subscription
	err chan error
}

// viewDescUnsubscription is a message requesting that the channelWorker
// goroutine unsubscribes the caller from the view named vn.
type viewDescUnsubscription struct {
	s   Subscription
	err chan error
}

// singleRecord is a message requesting that the channelWorker goroutine
// records the value v for the MeasureDesc md and tags in ct.
type singleRecord struct {
	ts tagging.TagsSet
	m  Measurement
}

// multiRecords is a message requesting that the channelWorker goroutine
// records the values vs for the MeasureDesc mds and tags in ct.
type multiRecords struct {
	ts tagging.TagsSet
	ms []Measurement
}

// reportingPeriod is a message requesting that the channelWorker goroutine
// modifies the min/max duration between reporting collected metrics.
type reportingPeriod struct {
	min, max time.Duration
}
