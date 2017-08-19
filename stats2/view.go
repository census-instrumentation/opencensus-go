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
	addSubscription(c chan *ViewData)
	deleteSubscription(c chan *ViewData)
	subscriptionExists(c chan *ViewData) bool
	subscriptionsCount() int
	subscriptions() map[chan *ViewData]subscription
	startCollectingForAdhoc()
	stopCollectingForAdhoc()
	isCollectingForAdhoc() bool
	collectedRows() []*Row
	clearRows()
	window() Window
	measure() Measure
	Name() string // Name returns the name of a View.
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
	m *MeasureFloat64

	// aggregation is the description of the aggregation to perform for this
	// view.
	a *AggregationFloat64

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

	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its AggregateValueFloat64.
	signatures map[string]AggregateValueFloat64
}

func (v *ViewFloat64) recordFloat64(ts *tags.TagSet, f float64) {
}

func (v *ViewFloat64) addSubscription(c chan *ViewData) {
	v.ss[c] = subscription{}
}

func (v *ViewFloat64) deleteSubscription(c chan *ViewData) {
	delete(v.ss, c)
}

func (v *ViewFloat64) subscriptionExists(c chan *ViewData) bool {
	_, ok := v.ss[c]
	return ok
}

func (v *ViewFloat64) subscriptionsCount() int {
	return len(v.ss)
}

func (v *ViewFloat64) subscriptions() map[chan *ViewData]subscription {
	return v.ss
}

func (v *ViewFloat64) startCollectingForAdhoc() {
	v.collectingForAdhoc = true
}

func (v *ViewFloat64) stopCollectingForAdhoc() {
	v.collectingForAdhoc = false
}

func (v *ViewFloat64) isCollectingForAdhoc() bool {
	return v.collectingForAdhoc
}

func (v *ViewFloat64) collectedRows() []*Row {
	// TODO: create []*Row and return them
	return nil
}

func (v *ViewFloat64) clearRows() {
	v.signatures = make(map[string]AggregateValueFloat64)
}

func (v *ViewFloat64) addSample(ts *tags.TagSet, f float64) {
	// TODO: add sample
}

func (v *ViewFloat64) window() Window {
	return v.w
}

func (v *ViewFloat64) measure() Measure {
	return v.m
}

// Name returns the name of ViewFloat64.
func (v *ViewFloat64) Name() string {
	return v.name
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
	m *MeasureInt64

	// aggregation is the description of the aggregation to perform for this
	// view.
	a *AggregationInt64

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

	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its AggregateValueInt64.
	signatures map[string]AggregateValueInt64
}

func (v *ViewInt64) recordInt64(ts *tags.TagSet, i int64) {
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

func (v *ViewInt64) isCollectingForAdhoc() bool {
	return v.collectingForAdhoc
}

func (v *ViewInt64) collectedRows() []*Row {
	// TODO: create []*Row and return them
	return nil
}

func (v *ViewInt64) clearRows() {
	v.signatures = make(map[string]AggregateValueInt64)
}

func (v *ViewInt64) addSample(ts *tags.TagSet, i int64) {
	// TODO: add sample
}

func (v *ViewInt64) window() Window {
	return v.w
}

func (v *ViewInt64) measure() Measure {
	return v.m
}

// Name returns the name of ViewInt64.
func (v *ViewInt64) Name() string {
	return v.name
}

// A ViewData is a set of rows about usage of the single measure associated
// with the given view during a particular window. Each row is specific to a
// unique set of tags.
type ViewData struct {
	v    View
	rows []*Row
}

// NewViewFloat64 creates a new *ViewFloat64.
func NewViewFloat64(name, description string, keys []tags.Key, measure MeasureFloat64, agg AggregationFloat64, wnd Window) (*ViewFloat64, error) {
	return nil, nil
}

// NewViewInt64 creates a new *ViewInt64.
func NewViewInt64(name, description string, keys []tags.Key, measure MeasureInt64, agg AggregationInt64, wnd Window) (*ViewInt64, error) {
	return nil, nil
}
