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

// Package httptrace contains OpenCensus tracing integrations with net/http.
package httptrace

import (
	"net/http"
	"strings"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

// TODO(jbd): Add godoc examples.

type transport struct {
	base    http.RoundTripper
	formats []propagation.HTTPFormat
}

// RoundTrip creates a trace.Span and inserts it into the outgoing request's headers.
// The created span can follow a parent span, if a parent is presented in
// the request's context.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	name := "Sent" + strings.Replace(req.URL.String(), req.URL.Scheme, ".", -1)
	// TODO(jbd): Discuss whether we want to prefix
	// outgoing requests with Sent.
	ctx := trace.StartSpan(req.Context(), name)
	req = req.WithContext(ctx)

	span := trace.FromContext(ctx)
	for _, f := range t.formats {
		f.ToRequest(span.SpanContext(), req)
	}

	resp, err := t.base.RoundTrip(req)

	// TODO(jbd): Add status and attributes.
	trace.EndSpan(ctx)
	return resp, err
}

// CancelRequest cancels an in-flight request by closing its connection.
func (t *transport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	if cr, ok := t.base.(canceler); ok {
		cr.CancelRequest(req)
	}
}

// NewTransport returns an http.RoundTripper that traces the outgoing requests.
//
// All the requests are done by the given base roundtripper. If nil is given,
// http.DefaultTransport is used. If its base HTTP RoundTripper implements CancelRequest,
// the returned round tripper will be cancelable.
//
// Traces are propagated via the provided HTTP propagation mechanisms.
func NewTransport(base http.RoundTripper, format ...propagation.HTTPFormat) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &transport{base: base, formats: format}
}

// NewHandler returns a http.Handler from the given handler
// that is aware of the incoming request's span.
// The span can be extracted from the incoming request in handler
// functions from incoming request's context:
//
//    span := trace.FromContext(r.Context())
//
// The span will be automatically ended by the handler.
//
// Incoming propagation mechanism is determined by the given HTTP propagators.
func NewHandler(base http.Handler, format ...propagation.HTTPFormat) http.Handler {
	return &handler{handler: base, formats: format}
}

type handler struct {
	handler http.Handler
	formats []propagation.HTTPFormat
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := "Recv" + strings.Replace(r.URL.String(), r.URL.Scheme, ".", -1)

	var (
		sc trace.SpanContext
		ok bool
	)

	ctx := r.Context()
	for _, f := range h.formats {
		sc, ok = f.FromRequest(r)
		if ok {
			break
		}
	}
	if ok {
		ctx = trace.StartSpanWithRemoteParent(ctx, name, sc, trace.StartSpanOptions{})
	} else {
		ctx = trace.StartSpan(ctx, name)
	}
	defer trace.EndSpan(ctx)

	// TODO(jbd): Add status and attributes.
	r = r.WithContext(ctx)
	h.handler.ServeHTTP(w, r)
}
