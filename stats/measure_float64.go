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
	"fmt"

	"github.com/google/working-instrumentation-go/tags"
)

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

// Is creates a new measurement/datapoint of type measurementFloat64.
func (m *MeasureFloat64) Is(v float64) Measurement {
	return &measurementFloat64{
		m: m,
		v: v,
	}
}

type measurementFloat64 struct {
	m *MeasureFloat64
	v float64
}

func (mf *measurementFloat64) record(ts *tags.TagSet) {
	for v := range mf.m.views {

		v.addSample(ts, mf.v)
	}
}
