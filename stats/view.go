// Copyright 2017, OpenCensus Authors
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
	"bytes"
	"fmt"
	"reflect"
	"time"

	"github.com/census-instrumentation/opencensus-go/tags"
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
	m Measure

	// start is time when view collection was started originally.
	start time.Time

	// ss are the channels through which the collected views data for this view
	// are sent to the consumers of this view.
	ss map[chan *ViewData]subscription

	// boolean to indicate if the the view should be collecting data even if no
	// client is subscribed to it. This is necessary for supporting a pull
	// model.
	isForcedCollection bool

	// TODO(jbd): Guard isForcedCollection.

	c *collector
}

// NewView creates a new view. Views need to be registered
// via RegisterView to enable data collection.
func NewView(name, description string, keys []tags.Key, measure Measure, agg Aggregation, window Window) *View {
	var keysCopy []tags.Key
	for _, k := range keys {
		keysCopy = append(keysCopy, k)
	}

	return &View{
		name,
		description,
		keysCopy,
		measure,
		time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC),
		make(map[chan *ViewData]subscription),
		false,
		&collector{make(map[string]aggregator), agg, window},
	}
}

// Name returns the name of view.
func (v *View) Name() string {
	return v.name
}

// Description returns the name of view.
func (v *View) Description() string {
	return v.description
}

func (v *View) addSubscription(c chan *ViewData) {
	v.ss[c] = subscription{}
}

func (v *View) deleteSubscription(c chan *ViewData) {
	delete(v.ss, c)
}

func (v *View) subscriptionExists(c chan *ViewData) bool {
	_, ok := v.ss[c]
	return ok
}

func (v *View) subscriptionsCount() int {
	return len(v.ss)
}

func (v *View) subscriptions() map[chan *ViewData]subscription {
	return v.ss
}

func (v *View) startForcedCollection() {
	v.isForcedCollection = true
}

func (v *View) stopForcedCollection() {
	v.isForcedCollection = false
}

func (v *View) isCollecting() bool {
	return v.subscriptionsCount() > 0 || v.isForcedCollection
}

func (v *View) clearRows() {
	v.c.clearRows()
}

func (v *View) collector() *collector {
	return v.c
}

// Window returns the timing window is being used to collect
// metrics on this view.
func (v *View) Window() Window {
	return v.c.w
}

// Aggregation returns the Aggregation used to aggregate the measurements
// collected by this view.
func (v *View) Aggregation() Aggregation {
	return v.c.a
}

// Measure returns the measure type the view is collecting measurements for.
func (v *View) Measure() Measure {
	return v.m
}

func (v *View) collectedRows(now time.Time) []*Row {
	return v.c.collectedRows(v.tagKeys, now)
}

func (v *View) addSample(ts *tags.TagSet, val interface{}, now time.Time) {
	if !v.isCollecting() {
		return
	}
	sig := tags.ToValuesString(ts, v.tagKeys)
	v.c.addSample(sig, val, now)
}

// A ViewData is a set of rows about usage of the single measure associated
// with the given view during a particular window. Each row is specific to a
// unique set of tags.
type ViewData struct {
	V          *View
	Start, End time.Time
	Rows       []*Row
}

// Row is the collected value for a specific set of key value pairs a.k.a tags.
type Row struct {
	Tags             []tags.Tag
	AggregationValue AggregationValue
}

func (r *Row) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{ ")
	buffer.WriteString("{ ")
	for _, t := range r.Tags {
		buffer.WriteString(fmt.Sprintf("{%v %v}", t.K.Name(), t.K.ValueAsString(t.V)))
	}
	buffer.WriteString(" }")
	buffer.WriteString(r.AggregationValue.String())
	buffer.WriteString(" }")
	return buffer.String()
}

// Equal returns true if both Rows are equal. Tags are expected to be ordered
// by the key name. Even both rows have the same tags but the tags appear in
// different orders it will return false.
func (r *Row) Equal(other *Row) bool {
	if r == other {
		return true
	}

	return reflect.DeepEqual(r.Tags, other.Tags) && r.AggregationValue.equal(other.AggregationValue)
}

// ContainsRow returns true if rows contain r.
func ContainsRow(rows []*Row, r *Row) bool {
	for _, x := range rows {
		if r.Equal(x) {
			return true
		}
	}
	return false
}

// EqualRows returns true if rows1, rows2 are equivalent. The rows position
// into the slice is taken into account.
func EqualRows(rows1, rows2 []*Row) (bool, string) {
	if len(rows1) != len(rows2) {
		return false, fmt.Sprintf("len(rows1)=%v and len(rows2)=%v", len(rows1), len(rows2))
	}

	for _, r1 := range rows1 {
		if !ContainsRow(rows2, r1) {
			return false, fmt.Sprintf("got unexpected row '%v' in rows1", r1)
		}
	}

	return true, ""
}
