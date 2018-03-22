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
	"fmt"
	"reflect"
	"sort"
	"sync/atomic"

	"go.opencensus.io/exporter"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/internal"
	"go.opencensus.io/tag"
)

// Deprecated: use exporter.Register
func RegisterExporter(e interface{}) {
	// TODO(ramonza): move this to the exporter package
	if e, ok := e.(exporter.View); ok {
		exporter.Register(e)
	}
}

// Deprecated: use exporter.Unregister
func UnregisterExporter(e interface{}) {
	if e, ok := e.(exporter.View); ok {
		exporter.Unregister(e)
	}
}

// View allows users to aggregate the recorded stats.Measurements.
// Views need to be passed to the Subscribe function to be before data will be
// collected and sent to Exporters.
type View struct {
	Name        string // Name of View. Must be unique. If unset, will default to the name of the Measure.
	Description string // Description is a human-readable description for this view.

	// TagKeys are the tag keys describing the grouping of this view.
	// A single Row will be produced for each combination of associated tag values.
	TagKeys []tag.Key

	// Measure is a stats.Measure to aggregate in this view.
	Measure stats.Measure

	// Aggregation is the aggregation function tp apply to the set of Measurements.
	Aggregation *Aggregation
}

// Deprecated: Use &View{}.
func New(name, description string, keys []tag.Key, measure stats.Measure, agg *Aggregation) (*View, error) {
	if measure == nil {
		panic("measure may not be nil")
	}
	return &View{
		Name:        name,
		Description: description,
		TagKeys:     keys,
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
// defaults for Name and Description and sorting the TagKeys
func (v *View) canonicalize() error {
	if v.Measure == nil {
		return fmt.Errorf("cannot subscribe view %q: measure not set", v.Name)
	}
	if v.Aggregation == nil {
		return fmt.Errorf("cannot subscribe view %q: aggregation not set", v.Name)
	}
	if v.Name == "" {
		v.Name = v.Measure.Name()
	}
	if v.Description == "" {
		v.Description = v.Measure.Description()
	}
	if err := checkViewName(v.Name); err != nil {
		return err
	}
	sort.Slice(v.TagKeys, func(i, j int) bool {
		return v.TagKeys[i].Name() < v.TagKeys[j].Name()
	})
	return nil
}

// viewInternal is the internal representation of a View.
type viewInternal struct {
	view       *View  // view is the canonicalized View definition associated with this view.
	subscribed uint32 // 1 if someone is subscribed and data need to be exported, use atomic to access
	collector  *collector
}

func newViewInternal(v *View) (*viewInternal, error) {
	return &viewInternal{
		view:      v,
		collector: &collector{make(map[string]aggregator), v.Aggregation},
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

func (v *viewInternal) collectedRows() []*exporter.Row {
	return v.collector.collectedRows(v.view.TagKeys)
}

func (v *viewInternal) addSample(m *tag.Map, val float64) {
	if !v.isSubscribed() {
		return
	}
	sig := string(encodeWithKeys(m, v.view.TagKeys))
	v.collector.addSample(sig, val)
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
