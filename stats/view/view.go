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

package view

import (
	"bytes"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/internal"
	"go.opencensus.io/tag"
)

// View allows users to aggregate the recorded stats.Measurements.
// Views need to be passed to the Subscribe function to be before data will be
// collected and sent to Exporters.
type View struct {
	Name        string // Name of View. Must be unique.
	Description string // Description is a human-readable description for this view.

	// GroupByTags are the tag keys describing the grouping of this view.
	// A single Row will be produced for each combination of associated tag values.
	GroupByTags []tag.Key

	// MeasureName is the name of the stats.Measure to aggregate in this view.
	MeasureName string

	// Aggregation is the aggregation function tp apply to the set of Measurements.
	Aggregation Aggregation
}

// Deprecated: Use &View{}.
func New(name, description string, keys []tag.Key, measure stats.Measure, agg Aggregation) (*View, error) {
	if measure == nil {
		panic("measure may not be nil")
	}
	return &View{
		Name:        name,
		Description: description,
		GroupByTags: keys,
		MeasureName: measure.Name(),
		Aggregation: agg,
	}, nil
}

// WithName returns a copy of the View with a new name. This is useful for
// renaming views to cope with limitations placed on metric names by various
// backends.
func (v *View) WithName(name string) *View {
	vNew := *v
	vNew.Name = name
	return &vNew
}

// Equal compares two views and returns true if they represent the same aggregation.
func (v *View) Equal(other *View) bool {
	if v == other {
		return true
	}
	if v == nil {
		return false
	}
	return reflect.DeepEqual(v.Aggregation, other.Aggregation) &&
		v.Name == other.Name &&
		v.MeasureName == other.MeasureName
}

type viewInternal struct {
	definition View
	subscribed uint32 // 1 if someone is subscribed and data need to be exported, use atomic to access
	collector  *collector
	measure    stats.Measure
}

func newViewInternal(v *View, m stats.Measure) *viewInternal {
	return &viewInternal{
		definition: *v,
		collector:  &collector{make(map[string]AggregationData), v.Aggregation},
		measure:    m,
	}
}

func (v *viewInternal) subscribe() {
	atomic.StoreUint32(&v.subscribed, 1)
}

func (v *viewInternal) unsubscribe() {
	atomic.StoreUint32(&v.subscribed, 0)
}

// isSubscribed returns true if the view is exporting
// data by subscription.
func (v *viewInternal) isSubscribed() bool {
	return atomic.LoadUint32(&v.subscribed) == 1
}

func (v *viewInternal) clearRows() {
	v.collector.clearRows()
}

func (v *viewInternal) collectedRows(now time.Time) []*Row {
	return v.collector.collectedRows(v.definition.GroupByTags, now)
}

func (v *viewInternal) addSample(m *tag.Map, val float64, now time.Time) {
	if !v.isSubscribed() {
		return
	}
	sig := string(encodeWithKeys(m, v.definition.GroupByTags))
	v.collector.addSample(sig, val, now)
}

// A Data is a set of rows about usage of the single measure associated
// with the given view. Each row is specific to a unique set of tags.
type Data struct {
	Measure    stats.Measure
	View       *View
	Start, End time.Time
	Rows       []*Row
}

// Row is the collected value for a specific set of key value pairs a.k.a tags.
type Row struct {
	Tags []tag.Tag
	Data AggregationData
}

func (r *Row) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{ ")
	buffer.WriteString("{ ")
	for _, t := range r.Tags {
		buffer.WriteString(fmt.Sprintf("{%v %v}", t.Key.Name(), t.Value))
	}
	buffer.WriteString(" }")
	buffer.WriteString(fmt.Sprintf("%v", r.Data))
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
	return reflect.DeepEqual(r.Tags, other.Tags) && r.Data.equal(other.Data)
}

func checkViewName(name string) error {
	if len(name) > internal.MaxNameLength {
		return fmt.Errorf("view name cannot be larger than %v", internal.MaxNameLength)
	}
	if !internal.IsPrintable(name) {
		return fmt.Errorf("view name needs to be an ASCII string")
	}
	return nil
}
