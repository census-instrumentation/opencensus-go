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

package stackdriver

import (
	"context"
	"encoding/binary"
	"net/http"
	"reflect"
	"testing"
	"time"

	"go.opencensus.io/trace"
)

func TestBundling(t *testing.T) {
	exporter := newExporter(Options{
		ProjectID:            "fakeProjectID",
		BundleDelayThreshold: time.Second / 10,
		BundleCountThreshold: 10,
	}, nil)
	ch := make(chan []*trace.SpanData)
	exporter.uploadFn = func(spans []*trace.SpanData) {
		ch <- spans
	}
	trace.RegisterExporter(exporter)

	trace.SetDefaultSampler(trace.AlwaysSample())
	for i := 0; i < 35; i++ {
		ctx := trace.StartSpan(context.Background(), "span")
		trace.EndSpan(ctx)
	}

	// Read the first three bundles.
	<-ch
	<-ch
	<-ch

	// Test that the fourth bundle isn't sent early.
	select {
	case <-ch:
		t.Errorf("bundle sent too early")
	case <-time.After(time.Second / 20):
		<-ch
	}

	// Test that there aren't extra bundles.
	select {
	case <-ch:
		t.Errorf("too many bundles sent")
	case <-time.After(time.Second / 5):
	}
}

func TestHTTPFormat(t *testing.T) {
	exporter, err := NewExporter(Options{ProjectID: "test"})
	if err != nil {
		t.Fatal(err)
	}

	traceID := [16]byte{16, 84, 69, 170, 120, 67, 188, 139, 242, 6, 177, 32, 0, 16, 0, 0}
	var spanID [8]byte
	binary.PutUvarint(spanID[:], 123)
	tests := []struct {
		incoming        string
		wantSpanContext trace.SpanContext
	}{
		{
			incoming: "105445aa7843bc8bf206b12000100000/123;o=1",
			wantSpanContext: trace.SpanContext{
				TraceID:      traceID,
				SpanID:       spanID,
				TraceOptions: 1,
			},
		},
		{
			incoming: "105445aa7843bc8bf206b12000100000/123;o=0",
			wantSpanContext: trace.SpanContext{
				TraceID:      traceID,
				SpanID:       spanID,
				TraceOptions: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.incoming, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			req.Header.Add(httpHeader, tt.incoming)
			sc, ok := exporter.FromRequest(req)
			if !ok {
				t.Errorf("exporter.FromRequest() = false; want true")
			}
			if got, want := sc, tt.wantSpanContext; !reflect.DeepEqual(got, want) {
				t.Errorf("exporter.FromRequest() returned span context %v; want %v", got, want)
			}

			req, _ = http.NewRequest("GET", "http://example.com", nil)
			exporter.ToRequest(sc, req)
			if got, want := req.Header.Get(httpHeader), tt.incoming; got != want {
				t.Errorf("exporter.ToRequest() returned header %q; want %q", got, want)
			}
		})
	}
}
