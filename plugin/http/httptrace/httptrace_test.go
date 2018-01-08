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

package httptrace

import (
	"bytes"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"testing"

	"go.opencensus.io/trace"
)

type testTransport struct {
	ch chan *http.Request
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.ch <- req
	return nil, errors.New("noop")
}

type testPropagator struct{}

func (t *testPropagator) FromRequest(req *http.Request) (sc trace.SpanContext, ok bool) {
	header := req.Header.Get("trace")
	buf, err := hex.DecodeString(header)
	if err != nil {
		log.Fatalf("Cannot decode trace header: %q", header)
	}
	r := bytes.NewReader(buf)
	r.Read(sc.TraceID[:])
	r.Read(sc.SpanID[:])
	opts, err := r.ReadByte()
	if err != nil {
		log.Fatalf("Cannot read trace options from trace header: %q", header)
	}
	sc.TraceOptions = trace.TraceOptions(opts)
	return sc, true
}

func (t *testPropagator) ToRequest(sc trace.SpanContext, req *http.Request) *http.Request {
	var buf bytes.Buffer
	buf.Write(sc.TraceID[:])
	buf.Write(sc.SpanID[:])
	buf.WriteByte(byte(sc.TraceOptions))
	req.Header.Set("trace", hex.EncodeToString(buf.Bytes()))
	return req
}

func TestTransport_RoundTrip(t *testing.T) {
	parent := trace.NewSpan("parent", trace.StartSpanOptions{})
	tests := []struct {
		name       string
		parent     *trace.Span
		wantHeader string
	}{
		{
			name:   "no parent",
			parent: nil,
		},
		{
			name:   "parent",
			parent: parent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &testTransport{ch: make(chan *http.Request, 1)}
			rt := NewTransport(transport, &testPropagator{})

			req, _ := http.NewRequest("GET", "http://foo.com", nil)
			if tt.parent != nil {
				req = req.WithContext(trace.WithSpan(req.Context(), tt.parent))
			}
			rt.RoundTrip(req)

			req = <-transport.ch
			span := trace.FromContext(req.Context())

			if header := req.Header.Get("trace"); header == "" {
				t.Fatalf("Trace header = empty; want valid trace header")
			}
			if span == nil {
				t.Fatalf("Got no spans in req context; want one")
			}
			if tt.parent != nil {
				if got, want := span.SpanContext().TraceID, tt.parent.SpanContext().TraceID; got != want {
					t.Errorf("span.SpanContext().TraceID=%v; want %v", got, want)
				}
			}
		})
	}
}

func TestHandler(t *testing.T) {
	traceID := [16]byte{16, 84, 69, 170, 120, 67, 188, 139, 242, 6, 177, 32, 0, 16, 0, 0}
	tests := []struct {
		header           string
		wantTraceID      trace.TraceID
		wantTraceOptions trace.TraceOptions
	}{
		{
			header:           "105445aa7843bc8bf206b12000100000000000000000000000",
			wantTraceID:      traceID,
			wantTraceOptions: trace.TraceOptions(0),
		},
		{
			header:           "105445aa7843bc8bf206b12000100000000000000000000001",
			wantTraceID:      traceID,
			wantTraceOptions: trace.TraceOptions(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			propagator := &testPropagator{}

			handler := NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				span := trace.FromContext(r.Context())
				sc := span.SpanContext()
				if got, want := sc.TraceID, tt.wantTraceID; got != want {
					t.Errorf("TraceID = %q; want %q", got, want)
				}
				if got, want := sc.TraceOptions, tt.wantTraceOptions; got != want {
					t.Errorf("TraceOptions = %v; want %v", got, want)
				}
			}), propagator)
			req, _ := http.NewRequest("GET", "http://foo.com", nil)
			req.Header.Add("trace", tt.header)
			handler.ServeHTTP(nil, req)
		})
	}
}
