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

// Package httptrace contains OpenCensus tracing integrations with net/http.
package httptrace

import (
	"net/http"
	"strings"

	"go.opencensus.io/trace"
)

// Transport is an http.RoundTripper that traces the outgoing requests.
type Transport struct {
	// Base is the base http.RoundTripper to be used to do the actual request.
	//
	// Optional. If nil, http.DefaultTransport is used.
	Base http.RoundTripper
}

// RoundTrip creates a trace.Span and inserts it into the outgoing request's headers.
// The created span can follow a parent span, if a parent is presented in
// the request's context.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	name := "Sent" + strings.Replace(req.URL.String(), req.URL.Scheme, ".", -1)
	span := trace.FromContext(req.Context()).StartSpan(name)
	req = req.WithContext(trace.WithSpan(req.Context(), span))
	resp, err := t.base().RoundTrip(req)
	// TODO(jbd): Add status and attributes.
	span.End()
	return resp, err
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

// TODO(jbd): Add Handler for incoming requests.
