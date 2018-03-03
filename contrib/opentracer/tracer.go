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

// Package opentracer contains an OpenTracing implementation for OpenCensus.
package opentracer // import "go.opencensus.io/contrib/opentracer"

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

const (
	httpHeader = "X-Opencensus-Trace"
)

// Tracer is a simple, thin interface for Span creation and SpanContext
// propagation.
type Tracer struct {
	logger Logger
}

// New returns an opentracing.Tracer backed by OpenCensus
func New(logger Logger) *Tracer {
	if logger == nil {
		logger = Stdout
	}

	return &Tracer{
		logger: logger,
	}
}

// Create, start, and return a new Span with the given `operationName` and
// incorporate the given StartSpanOption `opts`. (Note that `opts` borrows
// from the "functional options" pattern, per
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
//
// A Span with no SpanReference options (e.g., opentracing.ChildOf() or
// opentracing.FollowsFrom()) becomes the root of its own trace.
//
// Examples:
//
//     var tracer opentracing.Tracer = ...
//
//     // The root-span case:
//     sp := tracer.StartSpan("GetFeed")
//
//     // The vanilla child span case:
//     sp := tracer.StartSpan(
//         "GetFeed",
//         opentracing.ChildOf(parentSpan.Context()))
//
//     // All the bells and whistles:
//     sp := tracer.StartSpan(
//         "GetFeed",
//         opentracing.ChildOf(parentSpan.Context()),
//         opentracing.Tag{"user_agent", loggedReq.UserAgent},
//         opentracing.StartTime(loggedReq.Timestamp),
//     )
//
func (t *Tracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	var options opentracing.StartSpanOptions
	for _, opt := range opts {
		opt.Apply(&options)
	}

	var (
		parentSpan   *Span
		ocParentSpan *trace.Span
	)

	for _, ref := range options.References {
		if ref.Type == opentracing.ChildOfRef {
			if v, ok := ref.ReferencedContext.(*Span); ok {
				parentSpan = v
				ocParentSpan = v.ocSpan
				break
			}

			// implementation assumes same opentracing.Tracer implementation used by all
		}
	}

	var (
		ocSpan  = trace.NewSpan(operationName, ocParentSpan, trace.StartOptions{})
		baggage = makeBaggage(parentSpan)
		tags    = makeTags(options.Tags)
		span    = Span{
			tracer:  t,
			ocSpan:  ocSpan,
			baggage: baggage,
			tags:    tags,
		}
	)

	return &span
}

// Inject() takes the `sm` SpanContext instance and injects it for
// propagation within `carrier`. The actual type of `carrier` depends on
// the value of `format`.
//
// OpenTracing defines a common set of `format` values (see BuiltinFormat),
// and each has an expected carrier type.
//
// Other packages may declare their own `format` values, much like the keys
// used by `context.Context` (see
// https://godoc.org/golang.org/x/net/context#WithValue).
//
// Example usage (sans error handling):
//
//     carrier := opentracing.HTTPHeadersCarrier(httpReq.Header)
//     err := tracer.Inject(
//         span.Context(),
//         opentracing.HTTPHeaders,
//         carrier)
//
// NOTE: All opentracing.Tracer implementations MUST support all
// BuiltinFormats.
//
// Implementations may return opentracing.ErrUnsupportedFormat if `format`
// is not supported by (or not known by) the implementation.
//
// Implementations may return opentracing.ErrInvalidCarrier or any other
// implementation-specific error if the format is supported but injection
// fails anyway.
//
// See Tracer.Extract().
func (t *Tracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	span, ok := sm.(*Span)
	if !ok {
		return opentracing.ErrInvalidSpanContext
	}

	if carrier == nil {
		return opentracing.ErrInvalidCarrier
	}

	header := marshal(span)

	if format == opentracing.Binary {
		w, ok := carrier.(io.Writer)
		if !ok {
			return opentracing.ErrInvalidCarrier
		}
		io.WriteString(w, header)

	} else if format == opentracing.TextMap || format == opentracing.HTTPHeaders {
		m, ok := carrier.(opentracing.TextMapWriter)
		if !ok {
			return opentracing.ErrInvalidCarrier
		}
		m.Set(httpHeader, header)

	} else {
		return opentracing.ErrUnsupportedFormat
	}

	return nil
}

// Extract() returns a SpanContext instance given `format` and `carrier`.
//
// OpenTracing defines a common set of `format` values (see BuiltinFormat),
// and each has an expected carrier type.
//
// Other packages may declare their own `format` values, much like the keys
// used by `context.Context` (see
// https://godoc.org/golang.org/x/net/context#WithValue).
//
// Example usage (with StartSpan):
//
//
//     carrier := opentracing.HTTPHeadersCarrier(httpReq.Header)
//     clientContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
//
//     // ... assuming the ultimate goal here is to resume the trace with a
//     // server-side Span:
//     var serverSpan opentracing.Span
//     if err == nil {
//         span = tracer.StartSpan(
//             rpcMethodName, ext.RPCServerOption(clientContext))
//     } else {
//         span = tracer.StartSpan(rpcMethodName)
//     }
//
//
// NOTE: All opentracing.Tracer implementations MUST support all
// BuiltinFormats.
//
// Return values:
//  - A successful Extract returns a SpanContext instance and a nil error
//  - If there was simply no SpanContext to extract in `carrier`, Extract()
//    returns (nil, opentracing.ErrSpanContextNotFound)
//  - If `format` is unsupported or unrecognized, Extract() returns (nil,
//    opentracing.ErrUnsupportedFormat)
//  - If there are more fundamental problems with the `carrier` object,
//    Extract() may return opentracing.ErrInvalidCarrier,
//    opentracing.ErrSpanContextCorrupted, or implementation-specific
//    errors.
//
// See Tracer.Inject().
func (t *Tracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	var spanContext opentracing.SpanContext
	var err error

	if format == opentracing.Binary {
		spanContext, err = extractBinary(t, carrier)
		if err != nil {
			return nil, err
		}

	} else if format == opentracing.TextMap {
		spanContext, err = extractTextMap(t, carrier)
		if err != nil {
			return nil, err
		}

	} else if format == opentracing.HTTPHeaders {
		spanContext, err = extractHTTPHeaders(t, carrier)
		if err != nil {
			return nil, err
		}

	} else {
		return nil, fmt.Errorf("unhandled format, %v", format)
	}

	return spanContext, nil
}

// makeBaggage create a baggage from the span provided; if no span is provided,
// an empty map wll be returned
func makeBaggage(span *Span) []log.Field {
	var baggage []log.Field
	if span != nil {
		baggage = append(baggage, span.baggage...)
	}
	return baggage
}

func makeTags(tags map[string]interface{}) []log.Field {
	var fields []log.Field
	for k, v := range tags {
		fields = append(fields, makeLogFields(k, v)...)
	}
	return fields
}

const (
	version   = "1"
	separator = "-"
)

// marshal encodes to 1-{identifier}-{baggage}
func marshal(span *Span) string {
	values := url.Values{}
	for _, field := range span.baggage {
		values.Set(field.Key(), field.Value().(string))
	}

	var (
		spanContext = span.ocSpan.SpanContext()
		binary      = propagation.Binary(spanContext)
		identifier  = hex.EncodeToString(binary)
	)
	return version + separator +
		identifier + separator +
		values.Encode()
}

// unmarshal decodes to 1-{identifier}-{baggage}
func unmarshal(tracer *Tracer, value string) (opentracing.SpanContext, error) {
	segments := strings.SplitN(value, separator, 3)
	if len(segments) != 3 {
		return nil, opentracing.ErrSpanContextCorrupted
	}
	if segments[0] != version {
		return nil, opentracing.ErrSpanContextCorrupted
	}

	binary, err := hex.DecodeString(segments[1])
	if err != nil {
		return nil, opentracing.ErrSpanContextCorrupted
	}

	parentContext, ok := propagation.FromBinary(binary)
	if !ok {
		return nil, opentracing.ErrSpanContextCorrupted
	}

	var baggage []log.Field
	if encoded := segments[2]; len(encoded) > 0 {
		values, err := url.ParseQuery(encoded)
		if err != nil {
			return nil, opentracing.ErrSpanContextCorrupted
		}

		for key := range values {
			value := values.Get(key)
			baggage = upsert(baggage, log.String(key, value))
		}
	}

	var (
		ocSpan = trace.NewSpanWithRemoteParent("remote", parentContext, trace.StartOptions{})
		span   = Span{
			tracer:  tracer,
			ocSpan:  ocSpan,
			baggage: baggage,
		}
	)

	return &span, nil
}

func extractBinary(tracer *Tracer, carrier interface{}) (opentracing.SpanContext, error) {
	r, ok := carrier.(io.Reader)
	if !ok {
		return nil, opentracing.ErrInvalidCarrier
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, opentracing.ErrSpanContextCorrupted
	}
	return unmarshal(tracer, string(data))
}

func extractTextMap(tracer *Tracer, carrier interface{}) (opentracing.SpanContext, error) {
	var spanContext opentracing.SpanContext

	m, ok := carrier.(opentracing.TextMapReader)
	if !ok {
		return nil, opentracing.ErrInvalidCarrier
	}

	fn := func(key, value string) error {
		if key == httpHeader {
			v, err := unmarshal(tracer, value)
			if err != nil {
				return opentracing.ErrSpanContextCorrupted
			}
			spanContext = v
		}
		return nil
	}

	if err := m.ForeachKey(fn); err != nil {
		return nil, err
	}

	if spanContext == nil {
		return nil, opentracing.ErrSpanContextCorrupted
	}

	return spanContext, nil
}

type mapCarrier interface {
	ForeachKey(handler func(key, val string) error) error
	Set(key, val string)
}

func extractHTTPHeaders(tracer *Tracer, carrier interface{}) (opentracing.SpanContext, error) {
	var spanContext opentracing.SpanContext
	var mc mapCarrier

	if m, ok := carrier.(mapCarrier); !ok {
		if v, ok := carrier.(http.Header); ok {
			mc = opentracing.HTTPHeadersCarrier(v)
		} else {
			return nil, opentracing.ErrInvalidCarrier
		}
	} else {
		mc = m
	}

	fn := func(key, value string) error {
		if key == httpHeader {
			v, err := unmarshal(tracer, value)
			if err != nil {
				return opentracing.ErrSpanContextCorrupted
			}
			spanContext = v
		}
		return nil
	}

	if err := mc.ForeachKey(fn); err != nil {
		return nil, err
	}

	if spanContext == nil {
		return nil, opentracing.ErrSpanContextCorrupted
	}

	return spanContext, nil
}
