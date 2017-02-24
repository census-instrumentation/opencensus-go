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

import "fmt"

// Measurement is the interface that needs to be implemented by all all the
// measurement types.
type Measurement interface {
	measureDesc() MeasureDesc
	string() string
	bool() bool
	float64() float64
	int64() int64
}

// measurementString represents a measurement which value is of type string.
type measurementString struct {
	md MeasureDesc
	v  string
}

func (ms *measurementString) measureDesc() MeasureDesc {
	return ms.md
}

func (ms *measurementString) string() string {
	return ms.v
}

func (ms *measurementString) bool() bool {
	panic(fmt.Sprintf("called bool() on %v", ms))
}

func (ms *measurementString) float64() float64 {
	panic(fmt.Sprintf("called float64() on %v", ms))
}

func (ms *measurementString) int64() int64 {
	panic(fmt.Sprintf("called int64() on %v", ms))
}

// measurementBool represents a measurement which value is of type bool.
type measurementBool struct {
	md MeasureDesc
	v  bool
}

func (mb *measurementBool) measureDesc() MeasureDesc {
	return mb.md
}

func (mb *measurementBool) string() string {
	panic(fmt.Sprintf("called string() on %v", mb))
}

func (mb *measurementBool) bool() bool {
	return mb.v
}

func (mb *measurementBool) float64() float64 {
	panic(fmt.Sprintf("called float64() on %v", mb))
}

func (mb *measurementBool) int64() int64 {
	panic(fmt.Sprintf("called int64() on %v", mb))
}

// measurementFloat64 represents a measurement which value is of type float64.
type measurementFloat64 struct {
	md MeasureDesc
	v  float64
}

func (mf *measurementFloat64) measureDesc() MeasureDesc {
	return mf.md
}

func (mf *measurementFloat64) string() string {
	panic(fmt.Sprintf("called string() on %v", mf))
}

func (mf *measurementFloat64) bool() bool {
	panic(fmt.Sprintf("called bool() on %v", mf))
}

func (mf *measurementFloat64) float64() float64 {
	return mf.v
}

func (mf *measurementFloat64) int64() int64 {
	panic(fmt.Sprintf("called int64() on %v", mf))
}

// measurementInt64 represents a measurement which value is of type int64.
type measurementInt64 struct {
	md MeasureDesc
	v  int64
}

func (mi *measurementInt64) measureDesc() MeasureDesc {
	return mi.md
}

func (mi *measurementInt64) string() string {
	panic(fmt.Sprintf("called string() on %v", mi))
}

func (mi *measurementInt64) bool() bool {
	panic(fmt.Sprintf("called bool() on %v", mi))
}

func (mi *measurementInt64) float64() float64 {
	panic(fmt.Sprintf("called float64() on %v", mi))
}

func (mi *measurementInt64) int64() int64 {
	return mi.v
}
