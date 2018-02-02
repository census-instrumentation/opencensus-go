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

// Package httpstats provides OpenCensus stats instrumentation for the standard library http package.
package httpstats

import (
	"fmt"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	// See: http://unitsofmeasure.org/ucum.html
	unitByte          = "By"
	unitDimensionless = "1"
	unitMillisecond   = "ms"

	bytesBucketBoundaries  = []float64{0, 1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296}
	millisBucketBoundaries = []float64{0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000}

	aggCount      = stats.CountAggregation{}
	aggDistBytes  = stats.DistributionAggregation(bytesBucketBoundaries)
	aggDistMillis = stats.DistributionAggregation(millisBucketBoundaries)
)

var (
	// ClientRequest is the number of client requests started.
	ClientRequest = int64Measure("requests", "Number of HTTP requests started", unitDimensionless)
	// ClientRequestBodySize is the size of request body if set as ContentLength (uncompressed bytes).
	ClientRequestBodySize = int64Measure("request_size", "HTTP request body size (uncompressed)", unitByte)
	// ClientResponseBodySize is the size of response body (uncompressed bytes).
	ClientResponseBodySize = int64Measure("response_size", "HTTP response body size (uncompressed)", unitByte)
	// ClientLatency is the end-to-end client latency (in milliseconds).
	ClientLatency = floatMeasure("latency", "End-to-end latency", unitMillisecond)

	// ClientRequestCount is a count of all instrumented HTTP requests.
	ClientRequestCount = defaultView(ClientRequest, aggCount)
	// ClientRequestBodySizeDistribution is a view of the size distribution of all instrumented request bodies.
	ClientRequestBodySizeDistribution = defaultView(ClientRequestBodySize, aggDistBytes)
	// ClientResponseBodySizeDistribution is a view of the size distribution of instrumented response bodies.
	ClientResponseBodySizeDistribution = defaultView(ClientResponseBodySize, aggDistBytes)
	// ClientLatencyDistribution is a view of the latency distribution of all instrumented requests.
	ClientLatencyDistribution = defaultView(ClientLatency, aggDistMillis)
	// ClientResponseCountByStatusCode is a view of response counts by StatusCode.
	ClientRequestCountByMethod = mustView(stats.NewView(
		qualify("request_count_by_method"),
		"Client request count by HTTP method",
		[]tag.Key{Method},
		ClientRequest,
		aggCount,
		&stats.Cumulative{}))
	// ClientRequestCountByMethod is a count of all instrumented HTTP requests by Method.
	ClientResponseCountByStatusCode = mustView(stats.NewView(
		qualify("response_count_by_status_code"),
		"Client response count by status code",
		[]tag.Key{StatusCode},
		ClientLatency,
		aggCount,
		&stats.Cumulative{}))

	// Host is the value of the HTTP Host header.
	Host = key("host")
	// StatusCode is the numeric HTTP response status code, or "error" if a transport error occurred and no status code
	// was read.
	StatusCode = key("status_code")
	// Path is the URL path (not including query string) in the request.
	Path = key("path")
	// Method is the HTTP method of the request, capitalized (GET, POST, etc.).
	Method = key("method")
)

func defaultView(m stats.Measure, agg stats.Aggregation) *stats.View {
	v, err := stats.NewView(m.Name(), m.Description(), nil, m, agg, stats.Cumulative{})
	if err != nil {
		panic(err)
	}
	if err := stats.RegisterView(v); err != nil {
		panic(err)
	}
	return v
}

func key(name string) tag.Key {
	k, err := tag.NewKey(qualify(name))
	if err != nil {
		panic(err)
	}
	return k
}

func int64Measure(name, desc, unit string) *stats.MeasureInt64 {
	m, err := stats.NewMeasureInt64(qualify(name), desc, unit)
	if err != nil {
		panic(err)
	}
	return m
}

func floatMeasure(name, desc, unit string) *stats.MeasureFloat64 {
	m, err := stats.NewMeasureFloat64(qualify(name), desc, unit)
	if err != nil {
		panic(err)
	}
	return m
}

func mustView(v *stats.View, err error) *stats.View {
	if err != nil {
		panic(err)
	}
	return v
}

func qualify(suffix string) string {
	return fmt.Sprintf("opencensus.io/http/client/%s", suffix)
}
