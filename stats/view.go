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

	"github.com/census-instrumentation/opencensus-go/tags"
)

// View is the generic interface defining the various type of views.
type View interface {
	Name() string        // Name returns the name of a View.
	Description() string // Description returns the description of a View.
	addSubscription(c chan *ViewData)
	deleteSubscription(c chan *ViewData)
	subscriptionExists(c chan *ViewData) bool
	subscriptionsCount() int
	subscriptions() map[chan *ViewData]subscription

	startCollectingForAdhoc()
	stopCollectingForAdhoc()

	isCollecting() bool

	clearRows()

	collector() *collector
	window() Window
	measure() Measure
	collectedRows(now time.Time) []*Row

	addSample(ts *tags.TagSet, val interface{}, now time.Time)
}

// view is the data structure that holds the info describing the view as well
// as the aggregated data.
type view struct {
	// name of View. Must be unique.
	name        string
	description string

	// tagKeys to perform the aggregation on.
	tagKeys []tags.Key

	// Examples of measures are cpu:tickCount, diskio:time...
	m Measure

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

// NewViewInt64 creates a new *view.
func NewViewInt64(name, description string, keys []tags.Key, measure *MeasureInt64, agg Aggregation, wnd Window) View {
	return newView(name, description, keys, measure, agg, wnd)
}

// NewViewFloat64 creates a new *view.
func NewViewFloat64(name, description string, keys []tags.Key, measure *MeasureFloat64, agg Aggregation, wnd Window) View {
	return newView(name, description, keys, measure, agg, wnd)
}

func newView(name, description string, keys []tags.Key, measure Measure, agg Aggregation, wnd Window) *view {
	var keysCopy []tags.Key
	for _, k := range keys {
		keysCopy = append(keysCopy, k)
	}

	return &view{
		name,
		description,
		keysCopy,
		measure,
		time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		make(map[chan *ViewData]subscription),
		false,
		&collector{
			make(map[string]aggregator),
			agg,
			wnd,
		},
	}
}

// Name returns the name of view.
func (v *view) Name() string {
	return v.name
}

// Description returns the name of view.
func (v *view) Description() string {
	return v.description
}

func (v *view) addSubscription(c chan *ViewData) {
	v.ss[c] = subscription{}
}

func (v *view) deleteSubscription(c chan *ViewData) {
	delete(v.ss, c)
}

func (v *view) subscriptionExists(c chan *ViewData) bool {
	_, ok := v.ss[c]
	return ok
}

func (v *view) subscriptionsCount() int {
	return len(v.ss)
}

func (v *view) subscriptions() map[chan *ViewData]subscription {
	return v.ss
}

func (v *view) startCollectingForAdhoc() {
	v.collectingForAdhoc = true
}

func (v *view) stopCollectingForAdhoc() {
	v.collectingForAdhoc = false
}

func (v *view) isCollecting() bool {
	return v.subscriptionsCount() > 0 || v.collectingForAdhoc
}

func (v *view) clearRows() {
	v.c.clearRows()
}

func (v *view) collector() *collector {
	return v.c
}

func (v *view) window() Window {
	return v.c.w
}

func (v *view) measure() Measure {
	return v.m
}

func (v *view) collectedRows(now time.Time) []*Row {
	return v.c.collectedRows(v.tagKeys, now)
}

func (v *view) addSample(ts *tags.TagSet, val interface{}, now time.Time) {
	sig := tags.ToValuesString(ts, v.tagKeys)
	v.c.addSample(sig, val, now)
}

// A ViewData is a set of rows about usage of the single measure associated
// with the given view during a particular window. Each row is specific to a
// unique set of tags.
type ViewData struct {
	v    View
	rows []*Row
}

// Row is the collected value for a specific set of key value pairs a.k.a tags.
type Row struct {
	Tags             []tags.Tag
	AggregationValue AggregationValue
}
