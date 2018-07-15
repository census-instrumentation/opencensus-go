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

package logjson

import (
	"bytes"
	"io"
	"testing"
	"time"

	"go.opencensus.io/log"
)

type Stringer struct {
	value string
}

func (s Stringer) String() string {
	return s.value
}

func TestEncode(t *testing.T) {
	testCases := map[string]struct {
		Value interface{}
		Want  string
	}{
		"bool": {
			Value: true,
			Want:  `{"key":true}`,
		},
		"false": {
			Value: false,
			Want:  `{"key":false}`,
		},
		"duration": {
			Value: time.Second,
			Want:  `{"key":1000}`,
		},
		"error - present": {
			Value: io.EOF,
			Want:  `{"key":"EOF"}`,
		},
		"error - nil": {
			Value: error(nil),
			Want:  `{}`,
		},
		"float32": {
			Value: float32(1.23),
			Want:  `{"key":1.23}`,
		},
		"float64": {
			Value: float64(1.23),
			Want:  `{"key":1.23}`,
		},
		"int": {
			Value: 1,
			Want:  `{"key":1}`,
		},
		"int8": {
			Value: int8(1),
			Want:  `{"key":1}`,
		},
		"int16": {
			Value: int16(1),
			Want:  `{"key":1}`,
		},
		"int32": {
			Value: int32(1),
			Want:  `{"key":1}`,
		},
		"int64": {
			Value: int64(1),
			Want:  `{"key":1}`,
		},
		"string": {
			Value: "string",
			Want:  `{"key":"string"}`,
		},
		"strings": {
			Value: []string{"a", "b", "c"},
			Want:  `{"key":"a,b,c"}`,
		},
		"strings - nil": {
			Value: []string(nil),
			Want:  `{"key":null}`,
		},
		"stringer": {
			Value: Stringer{value: "value"},
			Want:  `{"key":"value"}`,
		},
		"time": {
			Value: time.Date(2018, time.July, 13, 5, 6, 7, 0, time.UTC),
			Want:  `{"key":"2018-07-13T05:06:07Z"}`,
		},
		"time - zero": {
			Value: time.Time{},
			Want:  `{"key":null}`,
		},
		"uint": {
			Value: uint(1),
			Want:  `{"key":1}`,
		},
		"uint8": {
			Value: uint8(1),
			Want:  `{"key":1}`,
		},
		"uint16": {
			Value: uint16(1),
			Want:  `{"key":1}`,
		},
		"uint32": {
			Value: uint32(1),
			Want:  `{"key":1}`,
		},
		"uint64": {
			Value: uint64(1),
			Want:  `{"key":1}`,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			field := log.Any("key", tc.Value)
			buf := bytes.NewBuffer(nil)
			err := Encode(buf, []log.Field{field})
			if err != nil {
				t.Fatalf("%v: got %v, want nil", label, err)
			}

			if got := buf.String(); got != tc.Want  {
				t.Errorf("%v: got %v, want %v", label, got, tc.Want)
			}
		})
	}
}

func TestEncodeString(t *testing.T) {
	testCases := map[string]struct {
		Input string
		Want  string
	}{
		"newlines": {
			Input: "\r\n",
			Want:  `"\r\n"`,
		},
		"quote": {
			Input: `"`,
			Want:  `"\""`,
		},
		"tab": {
			Input: "\t",
			Want:  `"\t"`,
		},
		"low byte": {
			Input: string([]byte{0x1}),
			Want:  `"\u0001"`,
		},
		"hello": {
			Input: `hello`,
			Want:  `"hello"`,
		},
		"hello - quoted": {
			Input: `"hello"`,
			Want:  `"\"hello\""`,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			encodeString(buf, tc.Input)
			if got := buf.String(); got != tc.Want {
				t.Errorf("%v: got %v, want %v", label, got, tc.Want)
			}
		})
	}
}
