// Copyright 2018, OpenCensus Authors
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

package log

import (
	"io"
	"reflect"
	"testing"
	"time"
)

func TestMerge(t *testing.T) {
	a := String("a", "abc")
	b := String("b", "def")

	testCases := map[string]struct {
		A    []Field
		B    []Field
		Want []Field
	}{
		"nil a": {
			A:    nil,
			B:    []Field{a},
			Want: []Field{a},
		},
		"nil b": {
			A:    []Field{a},
			B:    nil,
			Want: []Field{a},
		},
		"a before b": {
			A:    []Field{a},
			B:    []Field{b},
			Want: []Field{a, b},
		},
		"remove dupe - 1": {
			A:    []Field{a},
			B:    []Field{a, b},
			Want: []Field{a, b},
		},
		"remove dupe - 2": {
			A:    []Field{a, a},
			B:    nil,
			Want: []Field{a},
		},
		"remove dupe - 3": {
			A:    nil,
			B:    []Field{a, a},
			Want: []Field{a},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got := mergeFields(tc.A, tc.B)
			if !reflect.DeepEqual(got, tc.Want) {
				t.Errorf("%v: got %v, want %v", label, got, tc.Want)
			}
		})
	}
}

var Result []Field

func BenchmarkMerge(t *testing.B) {
	a := []Field{
		String("a1", "1"),
		String("a2", "2"),
		String("a3", "3"),
	}
	b := []Field{
		String("b1", "1"),
		String("b2", "2"),
		String("b3", "3"),
	}

	for i := 0; i < t.N; i++ {
		Result = mergeFields(a, b)
	}

	want := append(a, b...)
	if !reflect.DeepEqual(Result, want) {
		t.Errorf("want %#v, got %#v", Result, want)
	}
}

func TestFactories(t *testing.T) {
	testCases := map[string]struct {
		Value interface{}
		Want  Field
	}{
		"bool": {
			Value: true,
			Want: Field{
				Type: BoolType,
				Int:  1,
			},
		},
		"false": {
			Value: false,
			Want: Field{
				Type: BoolType,
				Int:  0,
			},
		},
		"duration": {
			Value: time.Second,
			Want: Field{
				Type: DurationType,
				Int:  int64(time.Second),
			},
		},
		"error - present": {
			Value: io.EOF,
			Want: Field{
				Type:      ErrorType,
				Interface: io.EOF,
			},
		},
		"error - nil": {
			Value: error(nil),
			Want: Field{
				Type:      NoOpType,
			},
		},
		"float32": {
			Value: float32(1.23),
			Want: Field{
				Type:  Float32Type,
				Float: float64(float32(1.23)),
			},
		},
		"float64": {
			Value: float64(1.23),
			Want: Field{
				Type:  Float64Type,
				Float: 1.23,
			},
		},
		"int": {
			Value: 1,
			Want: Field{
				Type: IntType,
				Int:  1,
			},
		},
		"int8": {
			Value: int8(1),
			Want: Field{
				Type: Int8Type,
				Int:  1,
			},
		},
		"int16": {
			Value: int16(1),
			Want: Field{
				Type: Int16Type,
				Int:  1,
			},
		},
		"int32": {
			Value: int32(1),
			Want: Field{
				Type: Int32Type,
				Int:  1,
			},
		},
		"int64": {
			Value: int64(1),
			Want: Field{
				Type: Int64Type,
				Int:  1,
			},
		},
		"string": {
			Value: "string",
			Want: Field{
				Type:   StringType,
				String: "string",
			},
		},
		"strings": {
			Value: []string{"a", "b", "c"},
			Want: Field{
				Type:      StringsType,
				Interface: []string{"a", "b", "c"},
			},
		},
		"strings - nil": {
			Value: []string(nil),
			Want: Field{
				Type:      StringsType,
				Interface: []string(nil),
			},
		},
		//"stringer": {
		//	Value: Stringer{value: "value"},
		//	Want: Field{
		//		Type: BoolType,
		//		Int:  1,
		//	},
		//},
		"time": {
			Value: time.Date(2018, time.July, 13, 5, 6, 7, 0, time.UTC),
			Want: Field{
				Type:      TimeType,
				Interface: time.Date(2018, time.July, 13, 5, 6, 7, 0, time.UTC),
			},
		},
		"time - zero": {
			Value: time.Time{},
			Want: Field{
				Type:      TimeType,
				Interface: time.Time{},
			},
		},
		"uint": {
			Value: uint(1),
			Want: Field{
				Type: UintType,
				Int:  1,
			},
		},
		"uint8": {
			Value: uint8(1),
			Want: Field{
				Type: Uint8Type,
				Int:  1,
			},
		},
		"uint16": {
			Value: uint16(1),
			Want: Field{
				Type: Uint16Type,
				Int:  1,
			},
		},
		"uint32": {
			Value: uint32(1),
			Want: Field{
				Type: Uint32Type,
				Int:  1,
			},
		},
		"uint64": {
			Value: uint64(1),
			Want: Field{
				Type: Uint64Type,
				Int:  1,
			},
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			got := Any("key", tc.Value)
			tc.Want.Key = got.Key
			if !reflect.DeepEqual(got, tc.Want) {
				t.Errorf("%v: got %v, want %v", label, got, tc.Want)
			}
		})
	}
}
