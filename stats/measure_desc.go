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

package stats

// MeasureDesc is the interface for all measures descriptions.
type MeasureDesc interface {
	Meta() *measureDesc
}

// MeasureDescString is the interface for measures of type string.
type MeasureDescString interface {
	MeasureDesc
	CreateMeasurement(v string) Measurement
}

// MeasureDescBool is the interface for measures of type bool.
type MeasureDescBool interface {
	MeasureDesc
	CreateMeasurement(v bool) Measurement
}

// MeasureDescFloat64 is the interface for measures of type float64.
type MeasureDescFloat64 interface {
	MeasureDesc
	CreateMeasurement(v float64) Measurement
}

// MeasureDescInt64 is the interface for measures of type int64.
type MeasureDescInt64 interface {
	MeasureDesc
	CreateMeasurement(v int64) Measurement
}

// measureDesc describes a data point (measurement) type accounted
// for by the stats library, such as RAM or CPU time.
type measureDesc struct {
	// The name must be unique. Used to link the MeasureDesc to a ViewDesc.
	// Examples are cpu:tickCount, diskio:time...
	name string
	// The description is used for display purposes only. It is meant to be
	// human readable and is used to show the resource in dashboards.
	// Example are CPU profile ticks, Disk I/O, Disk usage in usecs...
	description  string
	unit         *MeasurementUnit
	aggViewDescs map[ViewDesc]struct{}
}

func (md *measureDesc) Name() string {
	return md.name
}

func (md *measureDesc) Description() string {
	return md.description
}

func (md *measureDesc) Unit() *MeasurementUnit {
	return md.unit
}

func (md *measureDesc) String() string {
	return md.name
}

// MeasurementUnit is the unit of measurement for a resource.
type MeasurementUnit struct {
	Power10      int
	Numerators   []BasicUnit
	Denominators []BasicUnit
}

// BasicUnit is used for representing the basic units used to construct
// MeasurementUnits.
type BasicUnit byte

// These constants are the type of basic units allowed.
const (
	UnknownUnit BasicUnit = iota
	ScalarUnit
	BitsUnit
	BytesUnit
	SecsUnit
	CoresUnit
)
