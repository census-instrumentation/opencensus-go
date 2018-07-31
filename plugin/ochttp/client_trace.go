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
	"net/http/httptrace"
	"strings"

	"go.opencensus.io/trace"
)

// ClientTrace implements the annotation of all available hooks for
// httptrace.ClientTrace.
type ClientTrace struct {
	*trace.Span
}

// ClientTracer returns a httptrace.ClientTrace instance which instruments all
// emitted httptrace events on the provided Span.
func ClientTracer(s *trace.Span) *httptrace.ClientTrace {
	clientTrace := ClientTrace{s}

	return &httptrace.ClientTrace{
		GetConn:              clientTrace.GetConn,
		GotConn:              clientTrace.GotConn,
		PutIdleConn:          clientTrace.PutIdleConn,
		GotFirstResponseByte: clientTrace.GotFirstResponseByte,
		Got100Continue:       clientTrace.Got100Continue,
		DNSStart:             clientTrace.DNSStart,
		DNSDone:              clientTrace.DNSDone,
		ConnectStart:         clientTrace.ConnectStart,
		ConnectDone:          clientTrace.ConnectDone,
		TLSHandshakeStart:    clientTrace.TLSHandshakeStart,
		TLSHandshakeDone:     clientTrace.TLSHandshakeDone,
		WroteHeaders:         clientTrace.WroteHeaders,
		Wait100Continue:      clientTrace.Wait100Continue,
		WroteRequest:         clientTrace.WroteRequest,
	}
}

// GetConn implements a httptrace.ClientTrace hook
func (c ClientTrace) GetConn(hostPort string) {
	attrs := []trace.Attribute{
		trace.StringAttribute("httptrace.get_connection.host_port", hostPort),
	}
	c.Annotate(attrs, "GetConn")
}

// GotConn implements a httptrace.ClientTrace hook
func (c ClientTrace) GotConn(info httptrace.GotConnInfo) {
	attrs := []trace.Attribute{
		trace.BoolAttribute("httptrace.got_connection.reused", info.Reused),
		trace.BoolAttribute("httptrace.got_connection.was_idle", info.WasIdle),
	}
	if info.WasIdle {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.got_connection.idle_time", info.IdleTime.String()))
	}
	c.Annotate(attrs, "GotConn")
}

// PutIdleConn implements a httptrace.ClientTrace hook
func (c ClientTrace) PutIdleConn(err error) {
	var attrs []trace.Attribute
	if err != nil {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.put_idle_connection.error", err.Error()))
	}
	c.Annotate(attrs, "PutIdleConn")
}

// GotFirstResponseByte implements a httptrace.ClientTrace hook
func (c ClientTrace) GotFirstResponseByte() {
	c.Annotate(nil, "GotFirstResponseByte")
}

// Got100Continue implements a httptrace.ClientTrace hook
func (c ClientTrace) Got100Continue() {
	c.Annotate(nil, "Got100Continue")
}

// DNSStart implements a httptrace.ClientTrace hook
func (c ClientTrace) DNSStart(info httptrace.DNSStartInfo) {
	attrs := []trace.Attribute{
		trace.StringAttribute("httptrace.dns_start.host", info.Host),
	}
	c.Annotate(attrs, "DNSStart")
}

// DNSDone implements a httptrace.ClientTrace hook
func (c ClientTrace) DNSDone(info httptrace.DNSDoneInfo) {
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
	c.Annotate(attrs, "DNSDone")
}

// ConnectStart implements a httptrace.ClientTrace hook
func (c ClientTrace) ConnectStart(network, addr string) {
	attrs := []trace.Attribute{
		trace.StringAttribute("httptrace.connect_start.network", network),
		trace.StringAttribute("httptrace.connect_start.addr", addr),
	}
	c.Annotate(attrs, "ConnectStart")
}

// ConnectDone implements a httptrace.ClientTrace hook
func (c ClientTrace) ConnectDone(network, addr string, err error) {
	attrs := []trace.Attribute{
		trace.StringAttribute("httptrace.connect_done.network", network),
		trace.StringAttribute("httptrace.connect_done.addr", addr),
	}
	if err != nil {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.connect_done.error", err.Error()))
	}
	c.Annotate(attrs, "ConnectDone")
}

// TLSHandshakeStart implements a httptrace.ClientTrace hook
func (c ClientTrace) TLSHandshakeStart() {
	c.Annotate(nil, "TLSHandshakeStart")
}

// TLSHandshakeDone implements a httptrace.ClientTrace hook
func (c ClientTrace) TLSHandshakeDone(_ tls.ConnectionState, err error) {
	var attrs []trace.Attribute
	if err != nil {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.tls_handshake_done.error", err.Error()))
	}
	c.Annotate(attrs, "TLSHandshakeDone")
}

// WroteHeaders implements a httptrace.ClientTrace hook
func (c ClientTrace) WroteHeaders() {
	c.Annotate(nil, "WroteHeaders")
}

// Wait100Continue implements a httptrace.ClientTrace hook
func (c ClientTrace) Wait100Continue() {
	c.Annotate(nil, "Wait100Continue")
}

// WroteRequest implements a httptrace.ClientTrace hook
func (c ClientTrace) WroteRequest(info httptrace.WroteRequestInfo) {
	var attrs []trace.Attribute
	if info.Err != nil {
		attrs = append(attrs,
			trace.StringAttribute("httptrace.wrote_request.error", info.Err.Error()))
	}
	c.Annotate(attrs, "WroteRequest")
}
