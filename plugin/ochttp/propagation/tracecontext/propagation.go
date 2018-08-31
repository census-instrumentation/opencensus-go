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

// Package tracecontext contains HTTP propagator for TraceContext standard.
// See https://github.com/w3c/distributed-tracing for more information.
package tracecontext // import "go.opencensus.io/plugin/ochttp/propagation/tracecontext"

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"go.opencensus.io/trace/tracestate"
)

const (
	supportedVersion  = 0
	maxVersion        = 254
	maxTracestateLen  = 512
	traceparentHeader = "traceparent"
	tracestateHeader  = "tracestate"
)

var _ propagation.HTTPFormat = (*HTTPFormat)(nil)

// HTTPFormat implements the TraceContext trace propagation format.
type HTTPFormat struct{}

// SpanContextFromRequest extracts a span context from incoming requests.
func (f *HTTPFormat) SpanContextFromRequest(req *http.Request) (sc trace.SpanContext, ok bool) {
	h := req.Header.Get(traceparentHeader)
	if h == "" {
		return trace.SpanContext{}, false
	}
	sections := strings.Split(h, "-")
	if len(sections) < 3 {
		return trace.SpanContext{}, false
	}

	ver, err := hex.DecodeString(sections[0])
	if err != nil {
		return trace.SpanContext{}, false
	}
	if len(ver) == 0 || int(ver[0]) > supportedVersion || int(ver[0]) > maxVersion {
		return trace.SpanContext{}, false
	}

	tid, err := hex.DecodeString(sections[1])
	if err != nil {
		return trace.SpanContext{}, false
	}
	if len(tid) != 16 {
		return trace.SpanContext{}, false
	}
	copy(sc.TraceID[:], tid)

	sid, err := hex.DecodeString(sections[2])
	if err != nil {
		return trace.SpanContext{}, false
	}
	if len(sid) != 8 {
		return trace.SpanContext{}, false
	}
	copy(sc.SpanID[:], sid)

	if len(sections) == 4 {
		opts, err := hex.DecodeString(sections[3])
		if err != nil || len(opts) < 1 {
			return trace.SpanContext{}, false
		}
		sc.TraceOptions = trace.TraceOptions(opts[0])
	}

	// Don't allow all zero trace or span ID.
	if sc.TraceID == [16]byte{} || sc.SpanID == [8]byte{} {
		return trace.SpanContext{}, false
	}

	// Extract Tracestate
	tracestate, ok := tracestateFromRequest(req)
	if ok == false {
		return trace.SpanContext{}, false
	}
	sc.Tracestate = tracestate
	return sc, true
}

func tracestateFromRequest(req *http.Request) (*tracestate.Tracestate, bool) {
	h := req.Header.Get(tracestateHeader)
	if h == "" {
		return nil, true
	}

	var entries []tracestate.Entry
	pairs := strings.Split(h, ",")
	headerLenWithoutTrailingSpaces := len(pairs) - 1 // Number of commas
	for _, pair := range pairs {
		trimmedPair := strings.TrimSpace(pair)
		headerLenWithoutTrailingSpaces += len(trimmedPair)
		if headerLenWithoutTrailingSpaces > maxTracestateLen {
			// Drop the entire Tracestate
			return nil, true
		}
		kv := strings.Split(strings.TrimSpace(pair), "=")
		if len(kv) != 2 {
			return nil, false
		}
		entries = append(entries, tracestate.Entry{Key: kv[0], Value: kv[1]})
	}
	ts, err := tracestate.New(nil, entries...)
	if err != nil {
		return nil, false
	}

	return ts, true
}

func tracestateToRequest(sc trace.SpanContext, req *http.Request) {
	pairs := []string{}
	if sc.Tracestate != nil {
		entries := sc.Tracestate.Entries()

		for _, entry := range entries {
			pairs = append(pairs, strings.Join([]string{entry.Key, entry.Value}, "="))
		}
		h := strings.Join(pairs, ",")

		// According to the spec https://github.com/w3c/distributed-tracing/blob/master/trace_context/HTTP_HEADER_FORMAT.md
		// tracer can decide to forward tracestate if the header len exceeds maxTracestateLen.
		// The choice here is to not forward under such circumstances.
		if h != "" && len(h) <= maxTracestateLen {
			req.Header.Set(tracestateHeader, h)
		}
	}
}

// SpanContextToRequest modifies the given request to include traceparent and tracestate headers.
func (f *HTTPFormat) SpanContextToRequest(sc trace.SpanContext, req *http.Request) {
	h := fmt.Sprintf("%x-%x-%x-%x",
		[]byte{supportedVersion},
		sc.TraceID[:],
		sc.SpanID[:],
		[]byte{byte(sc.TraceOptions)})
	req.Header.Set(traceparentHeader, h)
	tracestateToRequest(sc, req)
}
