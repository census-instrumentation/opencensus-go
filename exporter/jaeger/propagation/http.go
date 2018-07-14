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

// Package propagation implement
package propagation // import "go.opencensus.io/exporter/jaeger/propagation"

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

const (
	httpHeaderMaxSize = 200
	httpHeader        = `uber-trace-id`

	jaegerSampled = 1
	jaegerDebug   = 2
)

var (
	errEmptyTraceHeader      = errors.New("empty trace header")
	errMaleformedTraceHeader = errors.New("malformed trace header")
)

var _ propagation.HTTPFormat = (*HTTPFormat)(nil)

// HTTPFormat implements propagation.HTTPFormat to propagate
// traces in HTTP headers for Google Cloud Platform and Stackdriver Trace.
type HTTPFormat struct{}

func (f *HTTPFormat) SpanContextToRequest(sc trace.SpanContext, req *http.Request) {
	req.Header.Set(httpHeader, contextToString(sc))
}

func (f *HTTPFormat) SpanContextFromRequest(req *http.Request) (sc trace.SpanContext, ok bool) {
	h := req.Header.Get(httpHeader)
	if h == "" {
		return trace.SpanContext{}, false
	}
	sc, err := contextFromString(h)
	if err != nil {
		return trace.SpanContext{}, false
	}
	return sc, true
}

func contextToString(sc trace.SpanContext) string {
	var isSampled = 0
	if sc.IsSampled() {
		isSampled = jaegerSampled
	}
	return fmt.Sprintf("%s:%s::%x",
		sc.TraceID.String(),
		sc.SpanID.String(),
		isSampled)
}

// ContextFromString reconstructs the Context encoded in a string
func contextFromString(value string) (trace.SpanContext, error) {
	var context trace.SpanContext
	if value == "" {
		return trace.SpanContext{}, errEmptyTraceHeader
	}
	parts := strings.Split(value, ":")
	if len(parts) != 4 {
		return trace.SpanContext{}, errMaleformedTraceHeader
	}

	buf, err := hex.DecodeString(parts[0])
	if err != nil {
		return trace.SpanContext{}, err
	}
	copy(context.TraceID[:], buf)

	sid, err := strconv.ParseUint(parts[1], 16, 64)
	if err != nil {
		return trace.SpanContext{}, err
	}
	binary.BigEndian.PutUint64(context.SpanID[:], sid)

	jflags, err := strconv.ParseUint(parts[3], 16, 8)
	if err != nil {
		return trace.SpanContext{}, err
	}
	flags := 0
	if (jflags&jaegerSampled == jaegerSampled) ||
		(flags&jaegerDebug == jaegerDebug) {
		flags = 1
	}
	context.TraceOptions = trace.TraceOptions(flags)
	return context, nil
}
