package zipkin_test

import (
	"context"
	"os"
	"testing"
	"time"

	zgo "github.com/openzipkin/zipkin-go"
	zhttp "github.com/openzipkin/zipkin-go/reporter/http"
	oczipkin "go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/trace"
)

func TestEndToEnd(t *testing.T) {
	url := os.Getenv("ZIPKIN_URL")
	if url == "" {
		t.Skip("No ZIPKIN_URL set")
	}
	r := zhttp.NewReporter(url, zhttp.BatchSize(1))
	endpoint, _ := zgo.NewEndpoint("default", ":0")
	exporter := oczipkin.NewExporter(r, endpoint)
	trace.RegisterExporter(exporter)

	ctx := context.Background()
	ctx, span := trace.StartSpan(ctx, "/a/b/c", trace.WithSampler(trace.AlwaysSample()), trace.WithSpanKind(trace.SpanKindServer))
	time.Sleep(150 * time.Millisecond)
	span.End()

	r.Close()
}
