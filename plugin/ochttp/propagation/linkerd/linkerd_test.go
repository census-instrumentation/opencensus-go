package linkerd

import (
	"net/http"
	"testing"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

func TestSatisfiesHTTPFormat(t *testing.T) {
	var _ propagation.HTTPFormat = (*HTTPFormat)(nil)
}

func TestShouldSample(t *testing.T) {
	cases := []struct {
		name         string
		flags        byte // Only the lowest byte of flags - the rest are unused.
		shouldSample bool
	}{
		{
			name:         "DebugEnabled",
			flags:        1,
			shouldSample: true,
		},
		{
			name:         "SamplingKnownAndEnabled",
			flags:        6,
			shouldSample: true,
		},
		{
			name:         "DebugEnabledAndSamplingKnownAndEnabled",
			flags:        7,
			shouldSample: true,
		},
		{
			// Debug mode forces sampling.
			name:         "DebugEnabledAndSamplingKnownAndDisabled",
			flags:        3,
			shouldSample: true,
		},
		{
			name:         "SamplingKnownAndDisabled",
			flags:        2,
			shouldSample: false,
		},
		{
			// This flag is undefined.
			name:         "HereBeDragons",
			flags:        8,
			shouldSample: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if shouldSample(tc.flags) != tc.shouldSample {
				t.Errorf("shouldSample(%08b): want %v, got %v", tc.flags, tc.shouldSample, !tc.shouldSample)
			}
		})
	}
}

func requestWithHeader(h string) *http.Request {
	r, _ := http.NewRequest("GET", "http://example.org", nil)
	r.Header.Set(l5dHeaderTrace, h)
	return r
}

func TestSpanContextFromRequest(t *testing.T) {
	cases := []struct {
		name string
		r    *http.Request
		ok   bool
		sc   trace.SpanContext
	}{
		{
			name: "ValidHeaderWithSamplingEnabled",
			r:    requestWithHeader("9BQdXcDJNdD9O0IEyfZCbzKk2yD11ZLnAAAAAAAAAAY="),
			ok:   true,
			sc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 50, 164, 219, 32, 245, 213, 146, 231},
				SpanID:       trace.SpanID{244, 20, 29, 93, 192, 201, 53, 208},
				TraceOptions: ocShouldSample,
			},
		},
		{
			name: "ValidHeaderWithoutParentID",
			r:    requestWithHeader("9BQdXcDJNdAAAAAAAAAAADKk2yD11ZLnAAAAAAAAAAYAAAAAAAAAAA=="),
			ok:   true,
			sc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 50, 164, 219, 32, 245, 213, 146, 231},
				SpanID:       trace.SpanID{244, 20, 29, 93, 192, 201, 53, 208},
				TraceOptions: ocShouldSample,
			},
		},
		{
			name: "ValidHeaderWith128BitTraceID",
			r:    requestWithHeader("9BQdXcDJNdAAAAAAAAAAADKk2yD11ZLnAAAAAAAAAAYAAAAAAAAAAQ=="),
			ok:   true,
			sc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 1, 50, 164, 219, 32, 245, 213, 146, 231},
				SpanID:       trace.SpanID{244, 20, 29, 93, 192, 201, 53, 208},
				TraceOptions: ocShouldSample,
			},
		},
		{
			name: "ValidHeaderWithSamplingDisabled",
			r:    requestWithHeader("laEAbScFR/gDfE/j8FV/8P8jOugI0dtmAAAAAAAAAAA="),
			ok:   true,
			sc: trace.SpanContext{
				TraceID: trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 255, 35, 58, 232, 8, 209, 219, 102},
				SpanID:  trace.SpanID{149, 161, 0, 109, 39, 5, 71, 248},
			},
		},
		{
			name: "InvalidHeaderEncoding",
			r:    requestWithHeader("PROBABLYNOTBASE64"),
			ok:   false,
		},
		{
			name: "InvalidHeaderLength",
			r:    requestWithHeader("bmVlZWVyZA=="),
			ok:   false,
		},
	}

	for _, tc := range cases {
		f := &HTTPFormat{}
		t.Run(tc.name, func(t *testing.T) {
			got, ok := f.SpanContextFromRequest(tc.r)
			if ok != tc.ok {
				t.Errorf("t.SpanContextFromRequest(): want ok %v, got %v", tc.ok, ok)
			}
			if got != tc.sc {
				t.Errorf("f.SpanContextFromRequest():\ngot:  %+v\nwant: %+v\n", got, tc.sc)
			}
		})
	}
}

func TestSpanContextToRequest(t *testing.T) {
	cases := []struct {
		name   string
		header string
		sc     trace.SpanContext
	}{
		{
			name:   "ValidHeaderWithSamplingEnabled",
			header: "9BQdXcDJNdAAAAAAAAAAADKk2yD11ZLnAAAAAAAAAAYAAAAAAAAAAA==",
			sc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 50, 164, 219, 32, 245, 213, 146, 231},
				SpanID:       trace.SpanID{244, 20, 29, 93, 192, 201, 53, 208},
				TraceOptions: ocShouldSample,
			},
		},
		{
			name:   "ValidHeaderWithSamplingDisabled",
			header: "laEAbScFR/gAAAAAAAAAAP8jOugI0dtmAAAAAAAAAAAAAAAAAAAAAA==",
			sc: trace.SpanContext{
				TraceID: trace.TraceID{0, 0, 0, 0, 0, 0, 0, 0, 255, 35, 58, 232, 8, 209, 219, 102},
				SpanID:  trace.SpanID{149, 161, 0, 109, 39, 5, 71, 248},
			},
		},
		{
			name:   "ValidHeaderWith128BitTraceID",
			header: "9BQdXcDJNdAAAAAAAAAAADKk2yD11ZLnAAAAAAAAAAYAAAAAAAAAAQ==",
			sc: trace.SpanContext{
				TraceID:      trace.TraceID{0, 0, 0, 0, 0, 0, 0, 1, 50, 164, 219, 32, 245, 213, 146, 231},
				SpanID:       trace.SpanID{244, 20, 29, 93, 192, 201, 53, 208},
				TraceOptions: ocShouldSample,
			},
		},
	}

	for _, tc := range cases {
		f := &HTTPFormat{}
		t.Run(tc.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "http://example.org", nil)
			f.SpanContextToRequest(tc.sc, r)
			got := r.Header.Get(l5dHeaderTrace)
			if got != tc.header {
				t.Errorf("f.SpanContextToRequest():\ngot header: %+v\nwant:       %+v\n", got, tc.header)
			}
		})
	}
}
