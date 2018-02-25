package amazon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-xray-sdk-go/header"
	"go.opencensus.io/exporter/xray"
	"go.opencensus.io/trace"
)

var (
	zeroSpanID = trace.SpanID{}
)

func TestSpanContextFromRequest(t *testing.T) {
	var (
		format  = &HTTPFormat{}
		traceID = trace.TraceID{0x5a, 0x96, 0x12, 0xa2, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10}
		spanID  = trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
	)

	t.Run("no header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
		_, ok := format.SpanContextFromRequest(req)
		if ok {
			t.Errorf("expected false; got true")
		}
	})

	t.Run("traceID only", func(t *testing.T) {
		var (
			req           = httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
			amazonTraceID = xray.MakeAmazonTraceID(traceID)
		)
		req.Header.Set(httpHeader, amazonTraceID)

		sc, ok := format.SpanContextFromRequest(req)
		if !ok {
			t.Errorf("expected true; got false")
		}
		if traceID != sc.TraceID {
			t.Errorf("expected %v; got %v", traceID, sc.TraceID)
		}
		if zeroSpanID != sc.SpanID {
			t.Errorf("expected true; got false")
		}
		if 0 != sc.TraceOptions {
			t.Errorf("expected 1; got %v", sc.TraceOptions)
		}
	})

	t.Run("traceID only with root prefix", func(t *testing.T) {
		var (
			req           = httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
			amazonTraceID = xray.MakeAmazonTraceID(traceID)
		)
		req.Header.Set(httpHeader, header.RootPrefix+amazonTraceID)

		sc, ok := format.SpanContextFromRequest(req)
		if !ok {
			t.Errorf("expected true; got false")
		}
		if traceID != sc.TraceID {
			t.Errorf("expected %v; got %v", traceID, sc.TraceID)
		}
		if zeroSpanID != sc.SpanID {
			t.Errorf("expected true; got false")
		}
		if 0 != sc.TraceOptions {
			t.Errorf("expected 1; got %v", sc.TraceOptions)
		}
	})

	t.Run("traceID with parentSpanID", func(t *testing.T) {
		var (
			req           = httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
			amazonTraceID = xray.MakeAmazonTraceID(traceID)
			amazonSpanID  = xray.MakeAmazonSpanID(spanID)
		)
		req.Header.Set(httpHeader, prefixRoot+amazonTraceID+";"+prefixParent+amazonSpanID)

		sc, ok := format.SpanContextFromRequest(req)
		if !ok {
			t.Errorf("expected true; got false")
		}
		if traceID != sc.TraceID {
			t.Errorf("expected %v; got %v", traceID, sc.TraceID)
		}
		if spanID != sc.SpanID {
			t.Errorf("expected %v; got %v", spanID, sc.SpanID)
		}
		if 0 != sc.TraceOptions {
			t.Errorf("expected 1; got %v", sc.TraceOptions)
		}
	})

	t.Run("traceID with parentSpanID and sampled", func(t *testing.T) {
		var (
			req           = httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
			amazonTraceID = xray.MakeAmazonTraceID(traceID)
			amazonSpanID  = xray.MakeAmazonSpanID(spanID)
		)
		req.Header.Set(httpHeader, prefixRoot+amazonTraceID+";"+prefixParent+amazonSpanID+";"+prefixSampled+"1")

		sc, ok := format.SpanContextFromRequest(req)
		if !ok {
			t.Errorf("expected true; got false")
		}
		if traceID != sc.TraceID {
			t.Errorf("expected %v; got %v", traceID, sc.TraceID)
		}
		if spanID != sc.SpanID {
			t.Errorf("expected %v; got %v", spanID, sc.SpanID)
		}
		if 1 != sc.TraceOptions {
			t.Errorf("expected 1; got %v", sc.TraceOptions)
		}
	})

	t.Run("bad traceID", func(t *testing.T) {
		var (
			req = httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
		)
		req.Header.Set(httpHeader, "1-bad-junk")

		_, ok := format.SpanContextFromRequest(req)
		if ok {
			t.Errorf("expected false; got true")
		}
	})

	t.Run("bad spanID", func(t *testing.T) {
		var (
			req           = httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
			amazonTraceID = xray.MakeAmazonTraceID(traceID)
		)
		req.Header.Set(httpHeader, prefixRoot+amazonTraceID+";"+prefixParent+"junk-span")

		_, ok := format.SpanContextFromRequest(req)
		if ok {
			t.Errorf("expected false; got true")
		}
	})
}

func TestSpanContextToRequest(t *testing.T) {
	var (
		format  = &HTTPFormat{}
		traceID = trace.TraceID{0x5a, 0x96, 0x12, 0xa2, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10}
		spanID  = trace.SpanID{1, 2, 3, 4, 5, 6, 7, 8}
		req, _  = http.NewRequest(http.MethodGet, "http://localhost/", nil)
	)

	t.Run("trace on", func(t *testing.T) {
		var sc = trace.SpanContext{
			TraceID:      traceID,
			SpanID:       spanID,
			TraceOptions: 1,
		}
		format.SpanContextToRequest(sc, req)
		v := req.Header.Get(httpHeader)
		if expected := "Root=1-5a9612a2-05060708090a0b0c0d0e0f10;Parent=0102030405060708;Sampled=1"; expected != v {
			t.Errorf("got %v; expected %v", expected, v)
		}
	})

	t.Run("trace off", func(t *testing.T) {
		var sc = trace.SpanContext{
			TraceID:      traceID,
			SpanID:       spanID,
			TraceOptions: 0,
		}
		format.SpanContextToRequest(sc, req)
		v := req.Header.Get(httpHeader)
		if expected := "Root=1-5a9612a2-05060708090a0b0c0d0e0f10;Parent=0102030405060708;Sampled=0"; expected != v {
			t.Errorf("got %v; expected %v", expected, v)
		}
	})
}
