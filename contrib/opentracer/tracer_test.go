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
	"bytes"
	"context"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"go.opencensus.io/trace"
)

func TestImplementsTracer(t *testing.T) {
	var tracer opentracing.Tracer = New(nil)
	if tracer == nil {
		t.Fatalf("should never occur")
	}
}

type spanCapturer struct {
	spans []*trace.SpanData
}

func (c *spanCapturer) ExportSpan(s *trace.SpanData) {
	c.spans = append(c.spans, s)
}

func TestTracer(t *testing.T) {
	var exporter spanCapturer
	trace.RegisterExporter(&exporter)
	trace.SetDefaultSampler(trace.AlwaysSample()) // Always trace for this demo.

	ctx := context.Background()
	tracer := New(nil)

	opentracing.SetGlobalTracer(tracer)

	parent, ctx := opentracing.StartSpanFromContext(ctx, "parent")
	child, _ := opentracing.StartSpanFromContext(ctx, "child")
	child.Finish()
	parent.Finish()

	if want := 2; len(exporter.spans) != want {
		t.Fatalf("want %v spans; got %v", want, len(exporter.spans))
	}

	c := exporter.spans[0]
	p := exporter.spans[1]

	if want := "child"; c.Name != want {
		t.Fatalf("want %v; got %v", want, c.Name)
	}
	if want := "parent"; p.Name != want {
		t.Fatalf("want %v; got %v", want, p.Name)
	}
}

func TestUnmarshal(t *testing.T) {
	tracer := New(nil)
	span := tracer.StartSpan("span")

	t.Run("TextMap", func(t *testing.T) {
		var carrier = opentracing.TextMapCarrier{}

		err := tracer.Inject(span.Context(), opentracing.TextMap, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.TextMap, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		if reflect.DeepEqual(span, spanContext) {
			t.Fatalf("want %#v; got %#v", span, spanContext)
		}
	})

	t.Run("Binary", func(t *testing.T) {
		var carrier = bytes.NewBuffer(nil)

		err := tracer.Inject(span.Context(), opentracing.Binary, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.Binary, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		if reflect.DeepEqual(span, spanContext) {
			t.Fatalf("want %#v; got %#v", span, spanContext)
		}
	})

	t.Run("HTTPHeaders", func(t *testing.T) {
		var (
			values  = url.Values{}
			carrier = opentracing.HTTPHeadersCarrier(values)
		)

		err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		if reflect.DeepEqual(span, spanContext) {
			t.Fatalf("want %#v; got %#v", span, spanContext)
		}
	})

	t.Run("HTTPHeaders - via http.Header", func(t *testing.T) {
		var carrier = http.Header{}

		err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		if reflect.DeepEqual(span, spanContext) {
			t.Fatalf("want %#v; got %#v", span, spanContext)
		}
	})

	t.Run("propagates baggage", func(t *testing.T) {
		opentracing.SetGlobalTracer(tracer)

		var (
			carrier = http.Header{}
			span, _ = opentracing.StartSpanFromContext(context.Background(), "span")
			baggage = log.String("key", "value")
		)

		span.SetBaggageItem(baggage.Key(), baggage.Value().(string))

		err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		remoteSpan := spanContext.(*Span)
		if want := []log.Field{baggage}; !reflect.DeepEqual(want, remoteSpan.baggage) {
			t.Fatalf("want %#v; got %v", want, remoteSpan.baggage)
		}
	})
}
