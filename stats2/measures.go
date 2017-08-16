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

// Measure is the interface for all measure types. A measure is required when
// defining a view.
type Measure interface {
	isMeasure() bool
}

func RegisterMeasure(m Measure) error {
	// TODO
	return nil
}

func UnregisterView(m Measure) error {
	// TODO
	return nil
}

// GetMeasureByName returns the registered measure associated with name.
func GetMeasureByName func(name string) (Measure, error) {
	// TODO
	return nil
}

// MeasureFloat64 is a measure of type float64.
type MeasureFloat64 struct {
	name        string
	description string
	viewAggs    map[*View]struct{}
}

// NewMeasureFloat64 creates a new measure of type MeasureFloat64.
func NewMeasureFloat64(name string, description string) MeasureFloat64 {
	return &MeasureFloat64{
		name:        name,
		description: description,
		viewAggs:    make(map[*View]struct{}),
	}
}

// Is creates a new measurement/datapoint of type measurementFloat64.
func (m *MeasureFloat64) Is(v float64) Measurement {
	return &measurementFloat64{
		m: m,
		v: v,
	}
}

func (m *MeasureFloat64) isMeasure() bool { return true }

// MeasureInt64 is a measure of type int64.
type MeasureInt64 struct {
	name        string
	description string
	viewAggs    map[*View]struct{}
}

// NewMeasureInt64 creates a new measure of type MeasureInt64.
func NewMeasureInt64(name string, description string) MeasureInt64 {
	return &MeasureInt64{
		name:        name,
		description: description,
		viewAggs:    make(map[*View]struct{}),
	}
}

// Is creates a new measurement/datapoint of type measurementInt64.
func (m *MeasureInt64) Is(v int64) Measurement {
	return &measurementInt64{
		m: m,
		v: v,
	}
}

func (md *measureInt64) isMeasure() bool { return true }
