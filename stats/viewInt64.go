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

	"github.com/google/working-instrumentation-go/tags"
)

// ViewInt64 is the data structure that holds the info describing the int64
// view as well as the aggregated data.
type ViewInt64 struct {
	// name of View. Must be unique.
	name        string
	description string

	// tagKeys to perform the aggregation on.
	tagKeys []tags.Key

	// Examples of measures are cpu:tickCount, diskio:time...
	m *MeasureInt64

	// Aggregation is the description of the aggregation to perform for this
	// view.
	a Aggregation

	// window is the window under which the aggregation is performed.
	w Window

	// start is time when view collection was started originally.
	start time.Time

	// ss are the channels through which the collected views data for this view
	// are sent to the consumers of this view.
	ss map[chan *ViewData]subscription

	// boolean to indicate if the the view should be collecting data even if no
	// client is subscribed to it. This is necessary for supporting a pull
	// model.
	collectingForAdhoc bool

	c *collector
}

// NewViewInt64 creates a new *ViewInt64.
func NewViewInt64(name, description string, keys []tags.Key, measure *MeasureInt64, agg Aggregation, wnd Window) *ViewInt64 {
	return &ViewInt64{
		name,
		description,
		keys,
		measure,
		agg,
		wnd,
		time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		make(map[chan *ViewData]subscription),
		false,
		&collector{
			make(map[string]aggregator),
			wnd,
			agg,
		},
	}
}

// Name returns the name of ViewInt64.
func (v *ViewInt64) Name() string {
	return v.name
}

// Description returns the name of ViewInt64.
func (v *ViewInt64) Description() string {
	return v.description
}

func (v *ViewInt64) addSubscription(c chan *ViewData) {
	v.ss[c] = subscription{}
}

func (v *ViewInt64) deleteSubscription(c chan *ViewData) {
	delete(v.ss, c)
}

func (v *ViewInt64) subscriptionExists(c chan *ViewData) bool {
	_, ok := v.ss[c]
	return ok
}

func (v *ViewInt64) subscriptionsCount() int {
	return len(v.ss)
}

func (v *ViewInt64) subscriptions() map[chan *ViewData]subscription {
	return v.ss
}

func (v *ViewInt64) startCollectingForAdhoc() {
	v.collectingForAdhoc = true
}

func (v *ViewInt64) stopCollectingForAdhoc() {
	v.collectingForAdhoc = false
}

func (v *ViewInt64) isCollecting() bool {
	return v.subscriptionsCount() > 0 || v.collectingForAdhoc
}

func (v *ViewInt64) collector() *collector {
	return v.c
}

func (v *ViewInt64) window() Window {
	return v.w
}

func (v *ViewInt64) measure() Measure {
	return v.m
}

func (v *ViewInt64) collectedRows() []*Row {
	return v.c.collectedRows(v.tagKeys, time.Now())
}

func (v *ViewInt64) addSample(ts *tags.TagSet, i int64) {
	sig := tags.ToValuesString(ts, v.tagKeys)
	v.c.addSample(sig, i, time.Now())
}
