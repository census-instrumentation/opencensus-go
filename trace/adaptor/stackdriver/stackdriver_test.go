package stackdriver

import (
	"context"
	"testing"
	"time"

	"github.com/census-instrumentation/opencensus-go/trace"
)

func TestBundling(t *testing.T) {
	exporter, err := NewExporter(Options{
		ProjectID:            "fakeProjectID",
		BundleDelayThreshold: time.Second / 10,
		BundleCountThreshold: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	ch := make(chan []*trace.SpanData)
	exporter.uploadFn = func(spans []*trace.SpanData) {
		ch <- spans
	}

	trace.SetDefaultSampler(trace.AlwaysSample())
	for i := 0; i < 35; i++ {
		ctx := trace.StartSpan(context.Background(), "span")
		trace.EndSpan(ctx)
	}

	// Read the first three bundles.
	<-ch
	<-ch
	<-ch

	// Test that the fourth bundle isn't sent early.
	select {
	case <-ch:
		t.Errorf("bundle sent too early")
	case <-time.After(time.Second / 20):
		<-ch
	}

	// Test that there aren't extra bundles.
	select {
	case <-ch:
		t.Errorf("too many bundles sent")
	case <-time.After(time.Second / 5):
	}
}
