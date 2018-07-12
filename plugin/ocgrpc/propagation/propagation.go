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

// Package propagation describes how gRPC propagates tracing metadata.
package propagation

import (
	"go.opencensus.io/trace"
)

// Format describes how to propagate SpanContext across RPC boundaries.
type Format interface {
	// InjectSpanContext appends headers that represent the given SpanContext.
	InjectSpanContext(sc trace.SpanContext, appendHeader func(string, string))

	// ExtractSpanContext reads a SpanContext from the given headers.
	ExtractSpanContext(readHeader func(string) []string) (trace.SpanContext, bool)
}

type defaultFormat struct{}

const traceContextKey = "grpc-trace-bin"

// Default returns a default propagation Format that uses the grpc-trace-bin
// header.
func Default() Format {
	return (*defaultFormat)(nil)
}

func (f *defaultFormat) InjectSpanContext(sc trace.SpanContext, appendHeader func(string, string)) {
	traceContextBinary := toBinary(sc)
	appendHeader(traceContextKey, string(traceContextBinary))
}

func (f *defaultFormat) ExtractSpanContext(readHeader func(string) []string) (trace.SpanContext, bool) {
	traceContext := readHeader(traceContextKey)
	if len(traceContext) == 0 {
		return trace.SpanContext{}, false
	}
	traceContextBinary := []byte(traceContext[0])
	return fromBinary(traceContextBinary)
}

// toBinary returns the binary format representation of a SpanContext.
//
// If sc is the zero value, toBinary returns nil.
func toBinary(sc trace.SpanContext) []byte {
	if sc == (trace.SpanContext{}) {
		return nil
	}
	var b [29]byte
	copy(b[2:18], sc.TraceID[:])
	b[18] = 1
	copy(b[19:27], sc.SpanID[:])
	b[27] = 2
	b[28] = uint8(sc.TraceOptions)
	return b[:]
}

// fromBinary returns the SpanContext represented by b.
//
// If b has an unsupported version ID or contains no TraceID, fromBinary
// returns with ok==false.
func fromBinary(b []byte) (sc trace.SpanContext, ok bool) {
	if len(b) == 0 || b[0] != 0 {
		return trace.SpanContext{}, false
	}
	b = b[1:]
	if len(b) >= 17 && b[0] == 0 {
		copy(sc.TraceID[:], b[1:17])
		b = b[17:]
	} else {
		return trace.SpanContext{}, false
	}
	if len(b) >= 9 && b[0] == 1 {
		copy(sc.SpanID[:], b[1:9])
		b = b[9:]
	}
	if len(b) >= 2 && b[0] == 2 {
		sc.TraceOptions = trace.TraceOptions(b[1])
	}
	return sc, true
}
