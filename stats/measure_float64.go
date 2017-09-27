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

// MeasureFloat64 is a measure of type float64.
type MeasureFloat64 struct {
	name        string
	unit        string
	description string
	views       map[View]bool
}

// Name returns the name of the measure.
func (m *MeasureFloat64) Name() string {
	return m.name
}

// Unit returns the unit of the measure.
func (m *MeasureFloat64) Unit() string {
	return m.unit
}

func (m *MeasureFloat64) addView(v View) {
	m.views[v] = true
}

func (m *MeasureFloat64) removeView(v View) {
	delete(m.views, v)
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

func (mf *measurementFloat64) isMeasurement() bool { return true }
