package propagation_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/plugin/ochttp/propagation/google"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

func TestRoundTripAllFormats(t *testing.T) {
	// TODO: test combinations of different formats for chains of calls
	formats := []propagation.HTTPFormat{
		&b3.HTTPFormat{},
		&google.HTTPFormat{},
	}

	ctx := context.Background()
	trace.SetDefaultSampler(trace.AlwaysSample())
	ctx, span := trace.StartSpan(ctx, "test")
	sc := span.SpanContext()
	wantStr := fmt.Sprintf("trace_id=%x, span_id=%x, options=%d", sc.TraceID, sc.SpanID, sc.TraceOptions)
	defer span.End()

	for _, format := range formats {
		srv := httptest.NewServer(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			sc, ok := format.FromRequest(req)
			if !ok {
				resp.WriteHeader(http.StatusBadRequest)
			}
			fmt.Fprintf(resp, "trace_id=%x, span_id=%x, options=%d", sc.TraceID, sc.SpanID, sc.TraceOptions)
		}))
		req, err := http.NewRequest("GET", srv.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		format.ToRequest(span.SpanContext(), req)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Fatal(resp.Status)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if got, want := string(body), wantStr; got != want {
			t.Errorf("%s; want %s", got, want)
		}
		srv.Close()
	}
}
