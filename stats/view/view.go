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
	"sort"
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
	Name        string // Name of View. Must be unique. If unset, will default to the name of the Measure.
	Description string // Description is a human-readable description for this view.

	// Dimensions describe the grouping of this view by tags associated with each Record call.
	// A single logical output Row will be produced for each combination of associated tag values.
	Dimensions []TagSelector

	// Measure is a stats.Measure to aggregate in this view.
	Measure stats.Measure

	// Aggregation is the aggregation function tp apply to the set of Measurements.
	Aggregation Aggregation
}

type TagSelector interface {
	OutputKey() tag.Key
	Extract(*tag.Map) (string, bool)
}

// Deprecated: Use &View{}.
func New(name, description string, keys []TagSelector, measure stats.Measure, agg Aggregation) (*View, error) {
	if measure == nil {
		panic("measure may not be nil")
	}
	return &View{
		Name:        name,
		Description: description,
		Dimensions:  keys,
		Measure:     measure,
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

// same compares two views and returns true if they represent the same aggregation.
func (v *View) same(other *View) bool {
	if v == other {
		return true
	}
	if v == nil {
		return false
	}
	return reflect.DeepEqual(v.Aggregation, other.Aggregation) &&
		v.Measure.Name() == other.Measure.Name()
}

// canonicalized returns a validated View canonicalized by setting explicit
// defaults for Name and Description and sorting the Dimensions
func (v *View) canonicalized() (*View, error) {
	if v.Measure == nil {
		return nil, fmt.Errorf("cannot subscribe view %q: measure not set", v.Name)
	}
	if v.Aggregation == nil {
		return nil, fmt.Errorf("cannot subscribe view %q: aggregation not set", v.Name)
	}
	vc := *v
	if vc.Name == "" {
		vc.Name = vc.Measure.Name()
	}
	if vc.Description == "" {
		vc.Description = vc.Measure.Description()
	}
	if err := checkViewName(vc.Name); err != nil {
		return nil, err
	}
	vc.Dimensions = make([]TagSelector, len(v.Dimensions))
	copy(vc.Dimensions, v.Dimensions)
	sort.Slice(vc.Dimensions, func(i, j int) bool {
		return vc.Dimensions[i].Name() < vc.Dimensions[j].Name()
	})
	return &vc, nil
}

// viewInternal is the internal representation of a View.
type viewInternal struct {
	view       *View  // view is the canonicalized View definition associated with this view.
	subscribed uint32 // 1 if someone is subscribed and data need to be exported, use atomic to access
	collector  *collector
}

func newViewInternal(v *View) (*viewInternal, error) {
	vc, err := v.canonicalized()
	if err != nil {
		return nil, err
	}
	return &viewInternal{
		view:      vc,
		collector: &collector{make(map[string]AggregationData), v.Aggregation},
	}, nil
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

func (v *viewInternal) collectedRows() []*Row {
	return v.collector.collectedRows(v.view.Dimensions)
}

func (v *viewInternal) addSample(m *tag.Map, val float64) {
	if !v.isSubscribed() {
		return
	}
	sig := string(encodeWithKeys(m, v.view.Dimensions))
	v.collector.addSample(sig, val)
}

// A Data is a set of rows about usage of the single measure associated
// with the given view. Each row is specific to a unique set of tags.
type Data struct {
	View       *View
	Start, End time.Time
	Rows       []*Row
}

// Row is the collected value for a specific set of key value pairs a.k.a tags.
type Row struct {
	DimensionValues []DimensionValue
	Data            AggregationData
}

type DimensionValue struct {
	Name  string
	Value string
}

func (r *Row) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{ ")
	buffer.WriteString("{ ")
	for _, t := range r.DimensionValues {
		buffer.WriteString(fmt.Sprintf("{%v %v}", t.Name, t.Value))
	}
	buffer.WriteString(" }")
	buffer.WriteString(fmt.Sprintf("%v", r.Data))
	buffer.WriteString(" }")
	return buffer.String()
}

// same returns true if both Rows are equal. Tags are expected to be ordered
// by the key name. Even both rows have the same tags but the tags appear in
// different orders it will return false.
func (r *Row) Equal(other *Row) bool {
	if r == other {
		return true
	}
	return reflect.DeepEqual(r.DimensionValues, other.DimensionValues) && r.Data.equal(other.Data)
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
