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

// Package amazon contains a propagation.HTTPFormat implementation
// for Amazon services: ELB, ALB, Lambda, etc.
package amazon // import "go.opencensus.io/plugin/ochttp/propagation/amazon"

import (
	"net/http"
	"strings"

	"go.opencensus.io/exporter/xray"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

const (
	httpHeaderMaxSize = 200
	httpHeader        = `X-Amzn-Trace-Id`
	prefixRoot        = "Root="
	prefixParent      = "Parent="
	prefixSampled     = "Sampled="
)

// HTTPFormat implements propagation.HTTPFormat to propagate
// traces in HTTP headers for for Amazon services: ELB, ALB, Lambda, etc.
type HTTPFormat struct{}

var _ propagation.HTTPFormat = (*HTTPFormat)(nil)

func parseHeader(h string) (trace.SpanContext, bool) {
	var (
		amazonTraceID string
		parentSpanID  string
		traceOptions  trace.TraceOptions
	)

	if strings.HasPrefix(h, prefixRoot) {
		h = h[len(prefixRoot):]
	}

	// Parse the trace id field.
	if index := strings.Index(h, `;`); index == -1 {
		amazonTraceID, h = h, h[len(h):]
	} else {
		amazonTraceID, h = h[:index], h[index+1:]
	}

	if strings.HasPrefix(h, prefixParent) {
		h = h[len(prefixParent):]

		if index := strings.Index(h, `;`); index == -1 {
			parentSpanID, h = h, h[len(h):]
		} else {
			parentSpanID, h = h[:index], h[index+1:]
		}
	}

	if strings.HasPrefix(h, prefixSampled) {
		h = h[len(prefixSampled):]
		if strings.HasPrefix(h, "1") {
			traceOptions = 1
		}
	}

	traceID, err := xray.ParseAmazonTraceID(amazonTraceID)
	if err != nil {
		return trace.SpanContext{}, false
	}

	spanID, err := xray.ParseAmazonSpanID(parentSpanID)
	if err != nil {
		return trace.SpanContext{}, false
	}

	return trace.SpanContext{
		TraceID:      traceID,
		SpanID:       spanID,
		TraceOptions: traceOptions,
	}, true
}

// SpanContextFromRequest extracts an AWS X-Ray Trace span context from incoming requests.
func (f *HTTPFormat) SpanContextFromRequest(req *http.Request) (sc trace.SpanContext, ok bool) {
	h := req.Header.Get(httpHeader)

	// See https://docs.aws.amazon.com/xray/latest/devguide/xray-concepts.html
	// for the header format. Return if the header is empty or missing, or if
	// the header is unreasonably large, to avoid making unnecessary copies of
	// a large string.
	if h == "" || len(h) > httpHeaderMaxSize {
		return trace.SpanContext{}, false
	}

	return parseHeader(h)
}

// SpanContextToRequest modifies the given request to include a AWS X-Ray trace header.
func (f *HTTPFormat) SpanContextToRequest(sc trace.SpanContext, req *http.Request) {
	var (
		header        = make([]byte, 0, 64)
		amazonTraceID = xray.MakeAmazonTraceID(sc.TraceID)
		amazonSpanID  = xray.MakeAmazonSpanID(sc.SpanID)
	)

	header = append(header, prefixRoot...)
	header = append(header, amazonTraceID...)
	header = append(header, ";"...)
	header = append(header, prefixParent...)
	header = append(header, amazonSpanID...)
	header = append(header, ";"...)
	header = append(header, prefixSampled...)

	if sc.TraceOptions&0x1 == 1 {
		header = append(header, "1"...)
	} else {
		header = append(header, "0"...)
	}

	req.Header.Set(httpHeader, string(header))
}
