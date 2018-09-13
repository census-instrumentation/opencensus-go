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

package tracestate

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	ts "go.opencensus.io/trace/tracestate"
	"strings"
)

var (
	oversizeValue   = strings.Repeat("a", MaxTracestateLen/2)
	oversizeEntry1  = ts.Entry{Key: "foo", Value: oversizeValue}
	oversizeEntry2  = ts.Entry{Key: "hello", Value: oversizeValue}
	entry1          = ts.Entry{Key: "foo", Value: "bar"}
	entry2          = ts.Entry{Key: "hello", Value: "world   example"}
	oversizeTs, _   = ts.New(nil, oversizeEntry1, oversizeEntry2)
	defaultTs, _    = ts.New(nil, nil...)
	nonDefaultTs, _ = ts.New(nil, entry1, entry2)
)

func TestFromRequest(t *testing.T) {
	tests := []struct {
		name     string
		tsHeader string
		wantTs   *ts.Tracestate
	}{
		{
			name:     "tracestate invalid entries delimiter",
			tsHeader: "foo=bar;hello=world",
			wantTs:   defaultTs,
		},
		{
			name:     "tracestate invalid key-value delimiter",
			tsHeader: "foo=bar,hello-world",
			wantTs:   defaultTs,
		},
		{
			name:     "tracestate invalid value character",
			tsHeader: "foo=bar,hello=world   example   \u00a0  ",
			wantTs:   defaultTs,
		},
		{
			name:     "tracestate blank key-value",
			tsHeader: "foo=bar,    ",
			wantTs:   defaultTs,
		},
		{
			name:     "tracestate oversize header",
			tsHeader: fmt.Sprintf("foo=%s,hello=%s", oversizeValue, oversizeValue),
			wantTs:   defaultTs,
		},
		{
			name:     "tracestate valid",
			tsHeader: "foo=bar   ,   hello=world   example",
			wantTs:   nonDefaultTs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			req.Header.Set("tracestate", tt.tsHeader)

			gotTs := FromRequest(req)
			if !reflect.DeepEqual(gotTs, tt.wantTs) {
				t.Errorf("HTTPFormat.FromRequest() gotTs = %v, want %v", gotTs, tt.wantTs)
			}
		})
	}
}

func TestToRequest(t *testing.T) {
	tests := []struct {
		name       string
		ts         *ts.Tracestate
		wantHeader string
	}{
		{
			name: "valid span context with default tracestate",
			ts: defaultTs,
			wantHeader: "",
		},
		{
			name: "valid span context with non default tracestate",
			ts: nonDefaultTs,
			wantHeader: "foo=bar,hello=world   example",
		},
		{
			name: "valid span context with oversize tracestate",
			ts: oversizeTs,
			wantHeader: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			ToRequest(tt.ts, req)

			h := req.Header.Get("tracestate")
			if got, want := h, tt.wantHeader; got != want {
				t.Errorf("HTTPFormat.ToRequest() tracestate header = %v, want %v", got, want)
			}
		})
	}
}
