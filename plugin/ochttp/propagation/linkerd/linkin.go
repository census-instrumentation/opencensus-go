// Package linkin provides linkerd trace propagation for Opencensus.
//
// Opencensus is a single distribution of libraries that automatically collects
// traces and metrics from your app, displays them locally, and sends them to
// any analysis tool. Opencensus supports the Zipkin request tracing system.
//
// Zipkin is a popular distributed tracing system, allowing requests to be
// traced through a distributed system. A request is broken into 'spans', each
// representing one part of the greater request path. Software that wishes to
// leverage Zipkin must be instrumented to do so (for example via Opencensus).
//
// linkerd is a popular service mesh. One of linkerd's selling points is that it
// provides Zipkin request tracing 'for free'. Software need not be 'fully'
// instrumented, and instead need only copy linkerd's l5d-ctx-* HTTP headers
// from incoming HTTP requests to any outgoing HTTP requests they spawn.
//
// Unfortunately while linkerd emits traces to Zipkin, it propagates trace data
// via a non-standard header. This package may be used as a drop-in replacement
// for https://godoc.org/go.opencensus.io/plugin/ochttp/propagation/b3 in
// environments that use linkerd for part or all of their request tracing needs.
//
// linkerd trace headers are base64 encoded 32 or 40 byte arrays (depending on
// whether the trace ID is 64 or 128bit) with the following Finagle
// serialization format:
//
//  spanID:8 parentID:8 traceIDLow:8 flags:8 traceIDHigh:8
//
// https://github.com/twitter/finagle/blob/345d7a2/finagle-core/src/main/scala/com/twitter/finagle/tracing/Id.scala#L113
// https://github.com/twitter/finagle/blob/345d7a2/finagle-core/src/main/scala/com/twitter/finagle/tracing/Flags.scala
package linkin

import (
	"encoding/base64"
	"net/http"

	"go.opencensus.io/trace"
)

const (
	l5dHeaderTrace = "l5d-ctx-trace"

	l5dFlagShouldSample byte               = 6
	ocShouldSample      trace.TraceOptions = 1
)

// HTTPFormat implements propagation.HTTPFormat to propagate traces in HTTP
// headers in linkerd propagation format. HTTPFormat omits the parent ID
// because it is not represented in the OpenCensus span context. Spans created
// from the incoming header will be the direct children of the client-side span.
// Similarly, the receiver of the outgoing spans should use client-side span
// created by OpenCensus as the parent.
type HTTPFormat struct{}

func shouldSample(f byte) bool {
	// If the debug bit is set, we should sample.
	if f&1 != 0 {
		return true
	}
	// If the sampling known and sampled bits are set, we should sample.
	return f&(1<<1) != 0 && f&(1<<2) != 0
}

// SpanContextFromRequest extracts linkerd span context from incoming requests.
func (f *HTTPFormat) SpanContextFromRequest(r *http.Request) (trace.SpanContext, bool) {
	sc := trace.SpanContext{}
	b, err := base64.StdEncoding.DecodeString(r.Header.Get(l5dHeaderTrace))
	if err != nil {
		return sc, false
	}
	if len(b) != 32 && len(b) != 40 {
		return sc, false
	}

	if len(b) == 40 {
		copy(sc.TraceID[0:8], b[32:])
	}
	copy(sc.TraceID[8:16], b[16:24])
	copy(sc.SpanID[:], b[0:8])

	if shouldSample(b[31]) {
		sc.TraceOptions = ocShouldSample
	}

	return sc, true
}

// SpanContextToRequest modifies the given request to include an l5d-ctx-trace
// HTTP header derived from the given SpanContext.
func (f *HTTPFormat) SpanContextToRequest(sc trace.SpanContext, r *http.Request) {
	b := [40]byte{}
	copy(b[0:8], sc.SpanID[:])
	copy(b[16:24], sc.TraceID[8:16])
	copy(b[32:], sc.TraceID[0:8])
	if sc.IsSampled() {
		b[31] = l5dFlagShouldSample
	}
	r.Header.Set(l5dHeaderTrace, base64.StdEncoding.EncodeToString(b[:]))
}
