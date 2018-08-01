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

package ochttp

import (
	"crypto/tls"
	"net/http"
	"net/http/httptrace"
	"strings"

	"go.opencensus.io/trace"
)

// SpanAnnotator implements the annotation of all available hooks for
// httptrace.ClientTrace.
type SpanAnnotator struct {
	sp *trace.Span
}

// NewSpanAnnotator returns a httptrace.ClientTrace which annotates all emitted
// httptrace events on the provided Span.
func NewSpanAnnotator(req *http.Request, s *trace.Span) *httptrace.ClientTrace {
	spanAnnotator := SpanAnnotator{s}

	return &httptrace.ClientTrace{
		GetConn:              spanAnnotator.GetConn,
		GotConn:              spanAnnotator.GotConn,
		PutIdleConn:          spanAnnotator.PutIdleConn,
		GotFirstResponseByte: spanAnnotator.GotFirstResponseByte,
		Got100Continue:       spanAnnotator.Got100Continue,
		DNSStart:             spanAnnotator.DNSStart,
		DNSDone:              spanAnnotator.DNSDone,
		ConnectStart:         spanAnnotator.ConnectStart,
		ConnectDone:          spanAnnotator.ConnectDone,
		TLSHandshakeStart:    spanAnnotator.TLSHandshakeStart,
		TLSHandshakeDone:     spanAnnotator.TLSHandshakeDone,
		WroteHeaders:         spanAnnotator.WroteHeaders,
		Wait100Continue:      spanAnnotator.Wait100Continue,
		WroteRequest:         spanAnnotator.WroteRequest,
	}
}

// GetConn implements a httptrace.ClientTrace hook
func (s SpanAnnotator) GetConn(hostPort string) {
	attrs := []trace.Attribute{
		trace.StringAttribute("httptrace.get_connection.host_port", hostPort),
	}
	s.sp.Annotate(attrs, "GetConn")
}

// GotConn implements a httptrace.ClientTrace hook
func (s SpanAnnotator) GotConn(info httptrace.GotConnInfo) {
	attrs := []trace.Attribute{
		trace.BoolAttribute("httptrace.got_connection.reused", info.Reused),
		trace.BoolAttribute("httptrace.got_connection.was_idle", info.WasIdle),
	}
	if info.WasIdle {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.got_connection.idle_time", info.IdleTime.String()))
	}
	s.sp.Annotate(attrs, "GotConn")
}

// PutIdleConn implements a httptrace.ClientTrace hook
func (s SpanAnnotator) PutIdleConn(err error) {
	var attrs []trace.Attribute
	if err != nil {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.put_idle_connection.error", err.Error()))
	}
	s.sp.Annotate(attrs, "PutIdleConn")
}

// GotFirstResponseByte implements a httptrace.ClientTrace hook
func (s SpanAnnotator) GotFirstResponseByte() {
	s.sp.Annotate(nil, "GotFirstResponseByte")
}

// Got100Continue implements a httptrace.ClientTrace hook
func (s SpanAnnotator) Got100Continue() {
	s.sp.Annotate(nil, "Got100Continue")
}

// DNSStart implements a httptrace.ClientTrace hook
func (s SpanAnnotator) DNSStart(info httptrace.DNSStartInfo) {
	attrs := []trace.Attribute{
		trace.StringAttribute("httptrace.dns_start.host", info.Host),
	}
	s.sp.Annotate(attrs, "DNSStart")
}

// DNSDone implements a httptrace.ClientTrace hook
func (s SpanAnnotator) DNSDone(info httptrace.DNSDoneInfo) {
	var addrs []string
	for _, addr := range info.Addrs {
		addrs = append(addrs, addr.String())
	}
	attrs := []trace.Attribute{
		trace.StringAttribute("httptrace.dns_done.addrs", strings.Join(addrs, " , ")),
	}
	if info.Err != nil {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.dns_done.error", info.Err.Error()))
	}
	s.sp.Annotate(attrs, "DNSDone")
}

// ConnectStart implements a httptrace.ClientTrace hook
func (s SpanAnnotator) ConnectStart(network, addr string) {
	attrs := []trace.Attribute{
		trace.StringAttribute("httptrace.connect_start.network", network),
		trace.StringAttribute("httptrace.connect_start.addr", addr),
	}
	s.sp.Annotate(attrs, "ConnectStart")
}

// ConnectDone implements a httptrace.ClientTrace hook
func (s SpanAnnotator) ConnectDone(network, addr string, err error) {
	attrs := []trace.Attribute{
		trace.StringAttribute("httptrace.connect_done.network", network),
		trace.StringAttribute("httptrace.connect_done.addr", addr),
	}
	if err != nil {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.connect_done.error", err.Error()))
	}
	s.sp.Annotate(attrs, "ConnectDone")
}

// TLSHandshakeStart implements a httptrace.ClientTrace hook
func (s SpanAnnotator) TLSHandshakeStart() {
	s.sp.Annotate(nil, "TLSHandshakeStart")
}

// TLSHandshakeDone implements a httptrace.ClientTrace hook
func (s SpanAnnotator) TLSHandshakeDone(_ tls.ConnectionState, err error) {
	var attrs []trace.Attribute
	if err != nil {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.tls_handshake_done.error", err.Error()))
	}
	s.sp.Annotate(attrs, "TLSHandshakeDone")
}

// WroteHeaders implements a httptrace.ClientTrace hook
func (s SpanAnnotator) WroteHeaders() {
	s.sp.Annotate(nil, "WroteHeaders")
}

// Wait100Continue implements a httptrace.ClientTrace hook
func (s SpanAnnotator) Wait100Continue() {
	s.sp.Annotate(nil, "Wait100Continue")
}

// WroteRequest implements a httptrace.ClientTrace hook
func (s SpanAnnotator) WroteRequest(info httptrace.WroteRequestInfo) {
	var attrs []trace.Attribute
	if info.Err != nil {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.wrote_request.error", info.Err.Error()))
	}
	s.sp.Annotate(attrs, "WroteRequest")
}
