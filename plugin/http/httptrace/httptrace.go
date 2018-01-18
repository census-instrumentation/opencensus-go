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
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/api/googleapi"
)

// TODO(jbd): Add godoc examples.

// Transport is an http.RoundTripper that traces the outgoing requests.
//
// Use NewTransport to create new transports.
type Transport struct {
	// Base represents the underlying roundtripper that does the actual requests.
	// If none is given, http.DefaultTransport is used.
	//
	// If base HTTP roundtripper implements CancelRequest,
	// the returned round tripper will be cancelable.
	Base http.RoundTripper

	// Formats are the mechanisms that propagate
	// the outgoing trace in an HTTP request.
	Formats []propagation.HTTPFormat
}

// RoundTrip creates a trace.Span and inserts it into the outgoing request's headers.
// The created span can follow a parent span, if a parent is presented in
// the request's context.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	name := "Sent" + strings.Replace(req.URL.String(), req.URL.Scheme, ".", -1)
	// TODO(jbd): Discuss whether we want to prefix
	// outgoing requests with Sent.
	ctx := trace.StartSpan(req.Context(), name)
	req = req.WithContext(ctx)

	span := trace.FromContext(ctx)
	for _, f := range t.Formats {
		f.ToRequest(span.SpanContext(), req)
	}

	resp, err := t.base().RoundTrip(req)

	// TODO(jbd): Add status and attributes.
	if err != nil {
		trace.EndSpan(ctx)
		return resp, err
	}

	// trace.EndSpan(ctx) will be invoked after
	// resp.Body.Close() has been invoked.
	resp.Body = &spanEndBody{rc: resp.Body, spanCtx: ctx}
	return resp, err
}

// spanEndBody wraps a response.Body and invokes
// trace.EndSpan on encountering io.EOF on reading
// the body of the original response.
type spanEndBody struct {
	rc      io.ReadCloser
	spanCtx context.Context

	endSpanOnce sync.Once
}

var _ io.ReadCloser = (*spanEndBody)(nil)

func (bpr *spanEndBody) Read(b []byte) (int, error) {
	n, err := bpr.rc.Read(b)
	if err == nil {
		return n, nil
	}

	// Otherwise, time to end the span or set the status
	switch err {
	case io.EOF:
		bpr.endSpan()
	default:
		// For all other errors, set the span status
		// If the error has a status and code
		// let's propagate those in the span
		if ge, ok := err.(*googleapi.Error); ok {
			trace.SetSpanStatus(bpr.spanCtx, trace.Status{
				Message: ge.Message,
				Code:    int32(ge.Code),
			})
		} else {
			trace.SetSpanStatus(bpr.spanCtx, trace.Status{
				// Code 2 is for Internal server error as per
				// https://github.com/googleapis/googleapis/blob/f704d14a7224a140bca5cc26835fae471eaf7281/google/rpc/code.proto#L44-L51
				Code:    2,
				Message: err.Error(),
			})
		}
	}
	return n, err
}

// endSpan invokes trace.EndSpan exactly once
func (bpr *spanEndBody) endSpan() {
	bpr.endSpanOnce.Do(func() {
		trace.EndSpan(bpr.spanCtx)
	})
}

func (bpr *spanEndBody) Close() error {
	// Invoking endSpan on Close will help catch the cases
	// in which a read returned a non-nil error, we set the
	// span status but didn't end the span.
	bpr.endSpan()
	return bpr.rc.Close()
}

// CancelRequest cancels an in-flight request by closing its connection.
func (t *Transport) CancelRequest(req *http.Request) {
	type canceler interface {
		CancelRequest(*http.Request)
	}
	if cr, ok := t.base().(canceler); ok {
		cr.CancelRequest(req)
	}
}

func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

// NewTransport returns an http.RoundTripper that traces the outgoing requests.
//
// Traces are propagated via the provided HTTP propagation mechanisms.
func NewTransport(format ...propagation.HTTPFormat) *Transport {
	return &Transport{Formats: format}
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
