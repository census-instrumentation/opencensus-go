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
	"fmt"
	"time"
)

// FieldType identifies the native type of the value represented by the Field union type.
type FieldType uint8

const (
	// UnknownType defined for completeness
	UnknownType  FieldType = iota
	BoolType               // BoolType defines true as field.Int == 1; everything else is false
	DurationType           // DurationType holds duration in field.Int
	ErrorType              // ErrorType holds error in field.Interface
	Float32Type            // Float32Type holds float32 as field.Float
	Float64Type            // Float64Type holds float64 as field.Float
	IntType                // IntType holds int as field.Int
	Int8Type               // Int8Type holds int8 as field.Int
	Int16Type              // Int16Type holds int16 as field.Int
	Int32Type              // Int32Type holds int32 as field.Int
	Int64Type              // Int64Type holds int64 as field.Int
	NoOpType               // NoOpType is a virtual type and indicates the Field should be ignored
	StringType             // StringType holds string as field.String
	StringsType            // StringsType holds []string as field.Interface
	StringerType           // StringerType holds fmt.Stringer as field.Interface
	TimeType               // TimeType holds time.Time as field.Interface
	UintType               // UintType holds uint as field.Int
	Uint8Type              // Uint8Type holds uint8 as field.Int
	Uint16Type             // Uint16Type holds uint16 as field.Int
	Uint32Type             // Uint32Type holds uint32 as field.Int
	Uint64Type             // Uint64Type holds uint64 as field.Int
)

// Field is a union type that defines a key value pair.  The key must be a string.  However,
// the value can be any type.
type Field struct {
	Key       string      // Key name
	Type      FieldType   // Type contained in this union
	Int       int64       // Int holds all Int types (see Type above)
	Float     float64     // Float holds all Float types (see Type above)
	String    string      // String hold string type
	Interface interface{} // Interface holds everything else
}

// Any attempts to encode the value provided.  If Any is unable to encode the value,
// Any will return a Field of type NoOp
func Any(key string, v interface{}) Field {
	switch value := v.(type) {
	case bool:
		return Bool(key, value)
	case time.Duration:
		return Duration(key, value)
	case error:
		return Error(key, value)
	case float32:
		return Float32(key, value)
	case float64:
		return Float64(key, value)
	case int:
		return Int(key, value)
	case int8:
		return Int8(key, value)
	case int16:
		return Int16(key, value)
	case int32:
		return Int32(key, value)
	case int64:
		return Int64(key, value)
	case string:
		return String(key, value)
	case []string:
		return Strings(key, value)
	case time.Time: // time.Time must be before fmt.Stringer
		return Time(key, value)
	case fmt.Stringer:
		return Stringer(key, value)
	case uint:
		return Uint(key, value)
	case uint8:
		return Uint8(key, value)
	case uint16:
		return Uint16(key, value)
	case uint32:
		return Uint32(key, value)
	case uint64:
		return Uint64(key, value)
	default:
		return NoOp()
	}
}

// Bool constructs a Field that holds a bool
func Bool(key string, v bool) Field {
	var value int64
	if v {
		value = 1
	}
	return Field{Key: key, Type: BoolType, Int: value}
}

// Duration constructs a Field that holds a time.Duration. Duration will be
// recorded in millis
func Duration(key string, v time.Duration) Field {
	return Field{Key: key, Type: DurationType, Int: int64(v)}
}

// Error constructs a Field that holds a error
func Error(key string, v error) Field {
	if v == nil {
		return Field{Type: NoOpType}
	}
	return Field{Key: key, Type: ErrorType, Interface: v}
}

// Float32 constructs a Field that holds a float32
func Float32(key string, v float32) Field {
	return Field{Key: key, Type: Float32Type, Float: float64(v)}
}

// Float64 constructs a Field that holds a float64
func Float64(key string, v float64) Field {
	return Field{Key: key, Type: Float64Type, Float: float64(v)}
}

// Int constructs a Field that holds a int
func Int(key string, v int) Field {
	return Field{Key: key, Type: IntType, Int: int64(v)}
}

// Int8 constructs a Field that holds a int8
func Int8(key string, v int8) Field {
	return Field{Key: key, Type: Int8Type, Int: int64(v)}
}

// Int16 constructs a Field that holds a int16
func Int16(key string, v int16) Field {
	return Field{Key: key, Type: Int16Type, Int: int64(v)}
}

// Int32 constructs a Field that holds a int32
func Int32(key string, v int32) Field {
	return Field{Key: key, Type: Int32Type, Int: int64(v)}
}

// Int64 constructs a Field that holds a int64
func Int64(key string, v int64) Field {
	return Field{Key: key, Type: Int64Type, Int: int64(v)}
}

// NoOp constructs a field that should be ignored.  Useful for cases where
// a functional constructor may optionally return a Field
func NoOp() Field {
	return Field{Type: NoOpType}
}

// String constructs a Field that holds a string
func String(key string, v string) Field {
	return Field{Key: key, Type: StringType, String: v}
}

// Strings constructs a Field that holds a []string
func Strings(key string, v []string) Field {
	return Field{Key: key, Type: StringsType, Interface: v}
}

// Stringer constructs a Field that holds a fmt.String
func Stringer(key string, v fmt.Stringer) Field {
	return Field{Key: key, Type: StringerType, Interface: v}
}

// Time constructs a Field that holds a time.Time
func Time(key string, v time.Time) Field {
	return Field{Key: key, Type: TimeType, Interface: v}
}

// Uint constructs a Field that holds a uint
func Uint(key string, v uint) Field {
	return Field{Key: key, Type: UintType, Int: int64(v)}
}

// Uint8 constructs a Field that holds a uint8
func Uint8(key string, v uint8) Field {
	return Field{Key: key, Type: Uint8Type, Int: int64(v)}
}

// Uint16 constructs a Field that holds a uint16
func Uint16(key string, v uint16) Field {
	return Field{Key: key, Type: Uint16Type, Int: int64(v)}
}

// Uint32 constructs a Field that holds a uint32
func Uint32(key string, v uint32) Field {
	return Field{Key: key, Type: Uint32Type, Int: int64(v)}
}

// Uint64 constructs a Field that holds a uint64
func Uint64(key string, v uint64) Field {
	return Field{Key: key, Type: Uint64Type, Int: int64(v)}
}

func fieldsContainsKey(fields []Field, key string) bool {
	for _, field := range fields {
		if field.Key == key {
			return true
		}
	}
	return false
}

func mergeFields(a, b []Field) []Field {
	fields := make([]Field, 0, len(a)+len(b))

	for _, field := range a {
		if fieldsContainsKey(fields, field.Key) {
			continue
		}
		fields = append(fields, field)
	}

	for _, field := range b {
		if fieldsContainsKey(fields, field.Key) {
			continue
		}
		fields = append(fields, field)
	}

	return fields
}
