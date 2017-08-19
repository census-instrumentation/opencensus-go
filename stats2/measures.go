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

import "fmt"

// Measure is the interface for all measure types. A measure is required when
// defining a view.
type Measure interface {
	Name() string
	addView(v View)
	removeView(v View)
	viewsCount() int
}

// MeasureFloat64 is a measure of type float64.
type MeasureFloat64 struct {
	name        string
	description string
	views       map[*ViewFloat64]bool
}

// NewMeasureFloat64 creates a new measure of type MeasureFloat64.
func NewMeasureFloat64(name string, description string) *MeasureFloat64 {
	return &MeasureFloat64{
		name:        name,
		description: description,
		views:       make(map[*ViewFloat64]bool),
	}
}

// Is creates a new measurement/datapoint of type measurementFloat64.
func (m *MeasureFloat64) Is(v float64) Measurement {
	return &measurementFloat64{
		m: m,
		v: v,
	}
}

// Name returns the name of the measure.
func (m *MeasureFloat64) Name() string {
	return m.name
}

func (m *MeasureFloat64) addView(v View) {
	vf64, ok := v.(*ViewFloat64)
	if !ok {
		panic(fmt.Sprintf("adding a view of type '%T' to MeasureFloat64. This is a bug in the stats library. It should never happen.", v))
	}

	m.views[vf64] = true
}

func (m *MeasureFloat64) removeView(v View) {
	vf64, ok := v.(*ViewFloat64)
	if !ok {
		panic(fmt.Sprintf("removing a view of type '%T' from MeasureFloat64. This is a bug in the stats library. It should never happen.", v))
	}

	delete(m.views, vf64)
}

func (m *MeasureFloat64) viewsCount() int { return len(m.views) }

// MeasureInt64 is a measure of type int64.
type MeasureInt64 struct {
	name        string
	description string
	views       map[*ViewInt64]bool
}

// NewMeasureInt64 creates a new measure of type MeasureInt64.
func NewMeasureInt64(name string, description string) *MeasureInt64 {
	return &MeasureInt64{
		name:        name,
		description: description,
		views:       make(map[*ViewInt64]bool),
	}
}

// Is creates a new measurement/datapoint of type measurementInt64.
func (m *MeasureInt64) Is(v int64) Measurement {
	return &measurementInt64{
		m: m,
		v: v,
	}
}

// Name returns the name of the measure.
func (m *MeasureInt64) Name() string {
	return m.name
}

func (m *MeasureInt64) addView(v View) {
	vi64, ok := v.(*ViewInt64)
	if !ok {
		panic(fmt.Sprintf("adding a view of type '%T' to MeasureInt64. This is a bug in the stats library. It should never happen.", v))
	}

	m.views[vi64] = true
}

func (m *MeasureInt64) removeView(v View) {
	vi64, ok := v.(*ViewInt64)
	if !ok {
		panic(fmt.Sprintf("removing a view of type '%T' from MeasureInt64. This is a bug in the stats library. It should never happen.", v))
	}

	delete(m.views, vi64)
}

func (m *MeasureInt64) viewsCount() int { return len(m.views) }
