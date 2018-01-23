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
package httpstats

import (
	"net/http"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"log"
	"fmt"
	"context"
	"time"
	"strconv"
	"net/http/httptrace"
)

var (
	Started           *stats.MeasureInt64
	ConnectionsOpened *stats.MeasureInt64
	Latency           *stats.MeasureFloat64

	StartedCount           *stats.View
	ConnectionsOpenedCount *stats.View
	LatencyDistribution    *stats.View

	Host       tag.Key
	StatusCode tag.Key
	Path       tag.Key
	Method     tag.Key
)

func init() {
	for _, c := range []struct {
		measure **stats.MeasureInt64
		name    string
		desc    string
		unit    string

		view **stats.View
	}{
		{measure: &Started, name: "started", desc: "Number of HTTP requests started", unit: "1", view: &StartedCount},
		{measure: &ConnectionsOpened, name: "connections_opened", desc: "Number of HTTP connections opened", unit: "1", view: &ConnectionsOpenedCount},
	} {
		fullname := qualify(c.name)
		m, err := stats.NewMeasureInt64(fullname, c.desc, c.unit)
		if err != nil {
			log.Fatalf("failed to create measure %q: %s", fullname, err)
		}
		*c.measure = m
		v, err := stats.NewView(fullname, c.desc, nil, m, &stats.CountAggregation{}, &stats.Cumulative{})
		if err != nil {
			log.Fatalf("failed to create view %q: %s", fullname, err)
		}
		*c.view = v
	}

	var err error
	fullname := qualify("latency")
	Latency, err = stats.NewMeasureFloat64(fullname, "End-to-end request latency", "microseconds")
	if err != nil {
		log.Fatalf("failed to create measure %q: %#v", fullname, err)
	}
	aggregation := &stats.DistributionAggregation{1, 100, 1000, 5000, 10000, 50000, 100000, 200000, 500000, 1000000}
	LatencyDistribution, err = stats.NewView(fullname, "End-to-end request latency", nil, Latency, aggregation, &stats.Cumulative{})

	for _, c := range []struct {
		key  *tag.Key
		name string
	}{
		{key: &Host, name: "host"},
		{key: &StatusCode, name: "status_code"},
		{key: &Path, name: "path"},
		{key: &Method, name: "method"},
	} {
		fullname := qualify(c.name)
		m, err := tag.NewKey(fullname)
		if err != nil {
			log.Fatalf("failed to validate tag key %q: %s", fullname, err)
		}
		*c.key = m
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
		resp *http.Response
		err  error
	)
	start := time.Now()
	defer func() {
		finishAndRecordStats(ctx, start, resp, err)
	}()
	stats.Record(ctx, Started.M(1))
	resp, err = t.base().RoundTrip(req)
	return resp, err
}

func finishAndRecordStats(ctx context.Context, start time.Time, resp *http.Response, err error) {
	if resp != nil {
		tags, _ := tag.NewMap(ctx, tag.Upsert(StatusCode, strconv.Itoa(resp.StatusCode)))
		ctx = tag.NewContext(ctx, tags)
	}
	stats.Record(ctx, Latency.M(float64(time.Since(start))/float64(time.Microsecond)))
}

func trace(ctx context.Context) *httptrace.ClientTrace {
	prev := httptrace.ContextClientTrace(ctx)
	var next httptrace.ClientTrace
	if prev != nil {
		next = *prev
	}
	next.ConnectDone = func(network, addr string, err error) {
		stats.Record(ctx, ConnectionsOpened.M(1))
		if prev != nil && prev.ConnectDone != nil {
			prev.ConnectDone(network, addr, err)
		}
	}
	return &next
}
