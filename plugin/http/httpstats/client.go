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

// Package httpstats provides OpenCensus stats support for the standard library
// HTTP client.
package httpstats

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"sync"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	// Number of client requests started.
	ClientRequest *stats.MeasureInt64
	// Size of request body if set as ContentLength.
	ClientRequestBodySize *stats.MeasureInt64
	// Size of response body, if returned in Content-Length header.
	ClientResponseBodySize *stats.MeasureInt64
	// Number of underlying transport connections opened.
	ClientConnectionsOpened *stats.MeasureInt64
	// End-to-end client latency.
	ClientLatency *stats.MeasureFloat64

	ClientRequestCount                 *stats.View
	ClientRequestBodySizeDistribution  *stats.View
	ClientResponseBodySizeDistribution *stats.View
	ClientConnectionsOpenedCount       *stats.View
	ClientLatencyDistribution          *stats.View

	ClientResponseCountByStatusCode *stats.View
	ClientRequestCountByMethod      *stats.View

	Host       tag.Key
	StatusCode tag.Key
	Path       tag.Key
	Method     tag.Key

	rpcBytesBucketBoundaries  = []float64{0, 1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296}
	rpcMillisBucketBoundaries = []float64{0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000}

	aggCount      = stats.CountAggregation{}
	aggDistBytes  = stats.DistributionAggregation(rpcBytesBucketBoundaries)
	aggDistMillis = stats.DistributionAggregation(rpcMillisBucketBoundaries)

	unitByte        = "byte"
	unitCount       = "1"
	unitMillisecond = "ms"
)

func init() {
	for _, c := range []struct {
		measure **stats.MeasureInt64
		name    string
		desc    string
		unit    string
	}{
		{measure: &ClientRequest, name: "requests", desc: "Number of HTTP requests started", unit: "1"},
		{measure: &ClientRequestBodySize, name: "request_size", desc: "Number of HTTP connections opened", unit: "1"},
		{measure: &ClientResponseBodySize, name: "response_size", desc: "Number of HTTP connections opened", unit: "1"},
		{measure: &ClientConnectionsOpened, name: "connections_opened", desc: "Number of HTTP connections opened", unit: "1"},
	} {
		fullname := qualify(c.name)
		m, err := stats.NewMeasureInt64(fullname, c.desc, c.unit)
		if err != nil {
			log.Fatalf("Failed to create measure %q: %s", fullname, err)
		}
		*c.measure = m
	}
	var err error
	ClientLatency, err = stats.NewMeasureFloat64(qualify("latency"), "End-to-end request latency", "microseconds")
	if err != nil {
		log.Fatalf("Failed to create measure: %v", err)
	}
	for _, view := range []struct {
		unit string
		agg  stats.Aggregation
		m    stats.Measure
		p    **stats.View
	}{
		{unitMillisecond, aggDistMillis, ClientLatency, &ClientLatencyDistribution},
		{unitByte, aggDistBytes, ClientRequestBodySize, &ClientRequestBodySizeDistribution},
		{unitByte, aggDistBytes, ClientResponseBodySize, &ClientResponseBodySizeDistribution},
		{unitCount, aggCount, ClientConnectionsOpened, &ClientConnectionsOpenedCount},
		{unitCount, aggCount, ClientRequest, &ClientRequestCount},
	} {
		var err error
		*view.p, err = stats.NewView(view.m.Name(), view.m.Description(), nil, view.m, view.agg, &stats.Cumulative{})
		if err != nil {
			log.Fatalf("Failed to create view: %v", err)
		}
	}
	for _, c := range []struct {
		key  *tag.Key
		name string
	}{
		{key: &Host, name: "host"},
		{key: &StatusCode, name: "status_code"},
		{key: &Path, name: "path"},
		{key: &Method, name: "method"},
	} {
		m, err := tag.NewKey(qualify(c.name))
		if err != nil {
			log.Fatalf("Failed to validate tag key: %v", err)
		}
		*c.key = m
	}
	ClientResponseCountByStatusCode, err = stats.NewView(
		qualify("response_count_by_status_code"),
		"Client response count by status code",
		[]tag.Key{StatusCode},
		ClientLatency,
		aggCount,
		&stats.Cumulative{})
	if err != nil {
		log.Fatal("Failed to create view")
	}
	ClientRequestCountByMethod, err = stats.NewView(
		qualify("request_count_by_method"),
		"Client request count by HTTP method",
		[]tag.Key{Method},
		ClientRequest,
		aggCount,
		&stats.Cumulative{})
	if err != nil {
		log.Fatal("Failed to create view")
	}
}

func qualify(suffix string) string {
	return fmt.Sprintf("opencensus.io/http/client/%s", suffix)
}

// Transport is an http.RoundTripper that collects stats for the outgoing requests.
type Transport struct {
	// Base represents the underlying roundtripper that does the actual requests.
	// If none is given, http.DefaultTransport is used.
	//
	// If base HTTP roundtripper implements CancelRequest,
	// the returned round tripper will be cancelable.
	Base http.RoundTripper
}

func (t Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

func (t Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	tags, _ := tag.NewMap(req.Context(),
		tag.Upsert(Host, req.URL.Host),
		tag.Upsert(Path, req.URL.Path),
		tag.Upsert(Method, req.Method))
	ctx := tag.NewContext(req.Context(), tags)
	ctx = httptrace.WithClientTrace(ctx, trace(ctx))
	req = req.WithContext(ctx)
	var (
		resp    *http.Response
		err     error
		tracker tracker
	)
	tracker.start = time.Now()
	tracker.ctx = ctx
	if req.Body == nil {
		//TODO: handle cases where ContentLength is not set
		tracker.reqSize = -1
	} else if req.ContentLength > 0 {
		tracker.reqSize = req.ContentLength
	}
	stats.Record(ctx, ClientRequest.M(1))
	resp, err = t.base().RoundTrip(req)
	if err != nil {
		tracker.end()
	} else {
		tracker.resp = resp
		if resp.Body == nil {
			tracker.end()
		} else {
			resp.Body = &tracker
		}
	}
	return resp, err
}

func trace(ctx context.Context) *httptrace.ClientTrace {
	prev := httptrace.ContextClientTrace(ctx)
	var next httptrace.ClientTrace
	if prev != nil {
		next = *prev
	}
	next.ConnectDone = func(network, addr string, err error) {
		stats.Record(ctx, ClientConnectionsOpened.M(1))
		if prev != nil && prev.ConnectDone != nil {
			prev.ConnectDone(network, addr, err)
		}
	}
	return &next
}

type tracker struct {
	respBody io.ReadCloser
	respSize int64
	reqSize  int64
	ctx      context.Context
	start    time.Time
	resp     *http.Response
	endOnce  sync.Once
}

func (t *tracker) end() {
	t.endOnce.Do(func() {
		var status string
		if t.resp != nil {
			status = strconv.Itoa(t.resp.StatusCode)
		} else {
			status = "error"
		}
		tags, _ := tag.NewMap(t.ctx, tag.Upsert(StatusCode, status))
		t.ctx = tag.NewContext(t.ctx, tags)
		m := []stats.Measurement{
			ClientLatency.M(float64(time.Since(t.start)) / float64(time.Millisecond)),
			ClientResponseBodySize.M(t.respSize),
		}
		if t.reqSize >= 0 {
			m = append(m, ClientRequestBodySize.M(t.reqSize))
		}
		stats.Record(t.ctx, m...)
	})
}

var _ io.ReadCloser = (*tracker)(nil)

func (t *tracker) Read(b []byte) (int, error) {
	n, err := t.respBody.Read(b)
	switch err {
	case nil:
		t.respSize += int64(n)
		return n, nil
	case io.EOF:
		t.end()
	}
	return n, err
}

func (t *tracker) Close() error {
	// Invoking endSpan on Close will help catch the cases
	// in which a read returned a non-nil error, we set the
	// span status but didn't end the span.
	t.end()
	return t.respBody.Close()
}
