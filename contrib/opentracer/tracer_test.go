package opentracer

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

	var (
		ctx    = context.Background()
		tracer = New(nil)
	)

	opentracing.SetGlobalTracer(tracer)

	parent, ctx := opentracing.StartSpanFromContext(ctx, "parent")
	child, _ := opentracing.StartSpanFromContext(ctx, "child")
	child.Finish()
	parent.Finish()

	if expected := 2; len(exporter.spans) != expected {
		t.Fatalf("expected %v spans; got %v", expected, len(exporter.spans))
	}

	var (
		c = exporter.spans[0]
		p = exporter.spans[1]
	)

	if expected := "child"; c.Name != expected {
		t.Fatalf("expected %v; got %v", expected, c.Name)
	}
	if expected := "parent"; p.Name != expected {
		t.Fatalf("expected %v; got %v", expected, p.Name)
	}
}

func TestUnmarshal(t *testing.T) {
	var (
		tracer = New(nil)
		span   = tracer.StartSpan("span")
	)

	t.Run("TextMap", func(t *testing.T) {
		var (
			carrier = opentracing.TextMapCarrier{}
		)

		err := tracer.Inject(span.Context(), opentracing.TextMap, carrier)
		if err != nil {
			t.Fatalf("expected nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.TextMap, carrier)
		if err != nil {
			t.Fatalf("expected nil; got %v", err)
		}

		if reflect.DeepEqual(span, spanContext) {
			t.Fatalf("expected %#v; got %#v", span, spanContext)
		}
	})

	t.Run("Binary", func(t *testing.T) {
		var (
			carrier = bytes.NewBuffer(nil)
		)

		err := tracer.Inject(span.Context(), opentracing.Binary, carrier)
		if err != nil {
			t.Fatalf("expected nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.Binary, carrier)
		if err != nil {
			t.Fatalf("expected nil; got %v", err)
		}

		if reflect.DeepEqual(span, spanContext) {
			t.Fatalf("expected %#v; got %#v", span, spanContext)
		}
	})

	t.Run("HTTPHeaders", func(t *testing.T) {
		var (
			values  = url.Values{}
			carrier = opentracing.HTTPHeadersCarrier(values)
		)

		err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("expected nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("expected nil; got %v", err)
		}

		if reflect.DeepEqual(span, spanContext) {
			t.Fatalf("expected %#v; got %#v", span, spanContext)
		}
	})

	t.Run("HTTPHeaders - via http.Header", func(t *testing.T) {
		var (
			carrier = http.Header{}
		)

		err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("expected nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("expected nil; got %v", err)
		}

		if reflect.DeepEqual(span, spanContext) {
			t.Fatalf("expected %#v; got %#v", span, spanContext)
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
			t.Fatalf("expected nil; got %v", err)
		}

		spanContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
		if err != nil {
			t.Fatalf("expected nil; got %v", err)
		}

		remoteSpan := spanContext.(*Span)
		if expected := []log.Field{baggage}; !reflect.DeepEqual(expected, remoteSpan.baggage) {
			t.Fatalf("expected %#v; got %v", expected, remoteSpan.baggage)
		}
	})
}
