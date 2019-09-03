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

package ochttp

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// The following client HTTP measures are supported for use in custom views.
var (
	ClientSentBytes = stats.Int64(
		"opencensus.io/http/client/sent_bytes",
		"Total bytes sent in request body (not including headers). This is uncompressed bytes",
		stats.UnitBytes,
	)
	ClientReceivedBytes = stats.Int64(
		"opencensus.io/http/client/received_bytes",
		"Total bytes received in response bodies (not including headers but including error responses with bodies). Should be measured from actual bytes received and read, not the value of the Content-Length header. This is uncompressed bytes. Responses with no body should record 0 for this value",
		stats.UnitBytes,
	)
	ClientRoundtripLatency = stats.Float64(
		"opencensus.io/http/client/roundtrip_latency",
		"Time between first byte of request headers sent to last byte of response received, or terminal error",
		stats.UnitMilliseconds,
	)
)

// The following server HTTP measures are supported for use in custom views:
var (
	ServerReceivedBytes = stats.Int64(
		"opencensus.io/http/server/received_bytes",
		"Total bytes received in request body (not including headers). This is uncompressed bytes",
		stats.UnitBytes,
	)
	ServerSentBytes = stats.Int64(
		"opencensus.io/http/server/sent_bytes",
		"Total bytes sent in response bodies (not including headers but including error responses with bodies). Should be measured from actual bytes received and read, not the value of the Content-Length header. This is uncompressed bytes. Responses with no body should record 0 for this value",
		stats.UnitBytes)
	ServerLatency = stats.Float64(
		"opencensus.io/http/server/server_latency",
		"Time between first byte of request headers read to last byte of response sent, or terminal error",
		stats.UnitMilliseconds)
)

// The following tags are applied to stats recorded by this package. Host, Path
// and Method are applied to all measures. StatusCode is not applied to
// ClientRequestCount or ServerRequestCount, since it is recorded before the status is known.
var (
	// KeyServerHost is the value of the HTTP KeyServerHost header.
	//
	// The value of this tag can be controlled by the HTTP client, so you need
	// to watch out for potentially generating high-cardinality labels in your
	// metrics backend if you use this tag in views.
	KeyServerHost = tag.MustNewKey("http_server_host")

	// KeyServerStatus is the HTTP server status code returned, as an integer e.g. 200, 404, 500,
	// or "error" if a transport error occurred and no status code was read.
	KeyServerStatus = tag.MustNewKey("http_server_status")

	// KeyServerPath is the URL path (not including query string) in the request.
	//
	// The value of this tag can be controlled by the HTTP client, so you need
	// to watch out for potentially generating high-cardinality labels in your
	// metrics backend if you use this tag in views.
	KeyServerPath = tag.MustNewKey("http_server_path")

	// KeyServerMethod is the HTTP method of the request, capitalized (GET, POST, etc.).
	KeyServerMethod = tag.MustNewKey("http_server_method")

	// KeyServerRoute is a low cardinality string representing the logical
	// handler of the request. This is usually the pattern registered on the a
	// ServeMux (or similar string).
	KeyServerRoute = tag.MustNewKey("http_server_route")
)

// Client tag keys.
var (
	// KeyClientMethod is the HTTP method, capitalized (i.e. GET, POST, PUT, DELETE, etc.).
	KeyClientMethod = tag.MustNewKey("http_client_method")
	// KeyClientPath is the URL path (not including query string).
	KeyClientPath = tag.MustNewKey("http_client_path")
	// KeyClientStatus is the HTTP status code as an integer (e.g. 200, 404, 500.), or "error" if no response status line was received.
	KeyClientStatus = tag.MustNewKey("http_client_status")
	// KeyClientHost is the value of the request Host header.
	KeyClientHost = tag.MustNewKey("http_client_host")
)

// Default distributions used by views in this package.
var (
	DefaultSizeDistribution    = view.Distribution(1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296)
	DefaultLatencyDistribution = view.Distribution(1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000)
)

// Package ochttp provides some convenience views for client measures.
// You still need to register these views for data to actually be collected.
var (
	ClientSentBytesView = &view.View{
		Name:        "opencensus.io/http/client/sent_bytes",
		Measure:     ClientSentBytes,
		Aggregation: DefaultSizeDistribution,
		Description: "Total bytes sent in request body (not including headers), by HTTP method and response status",
		TagKeys:     []tag.Key{KeyClientMethod, KeyClientStatus},
	}

	ClientReceivedBytesView = &view.View{
		Name:        "opencensus.io/http/client/received_bytes",
		Measure:     ClientReceivedBytes,
		Aggregation: DefaultSizeDistribution,
		Description: "Total bytes received in response bodies (not including headers but including error responses with bodies), by HTTP method and response status",
		TagKeys:     []tag.Key{KeyClientMethod, KeyClientStatus},
	}

	ClientRoundtripLatencyView = &view.View{
		Name:        "opencensus.io/http/client/roundtrip_latency",
		Measure:     ClientRoundtripLatency,
		Aggregation: DefaultLatencyDistribution,
		Description: "End-to-end latency, by HTTP method and response status",
		TagKeys:     []tag.Key{KeyClientMethod, KeyClientStatus},
	}

	ClientCompletedCountView = &view.View{
		Name:        "opencensus.io/http/client/completed_count",
		Measure:     ClientRoundtripLatency,
		Aggregation: view.Count(),
		Description: "Count of completed requests, by HTTP method and response status",
		TagKeys:     []tag.Key{KeyClientMethod, KeyClientStatus},
	}
)

// Package ochttp provides some convenience views for server measures.
// You still need to register these views for data to actually be collected.
var (
	ServerReceivedBytesView = &view.View{
		Name:        "opencensus.io/http/server/received_bytes",
		Description: "Size distribution of HTTP request body",
		Measure:     ServerReceivedBytes,
		Aggregation: DefaultSizeDistribution,
		TagKeys:     []tag.Key{KeyServerMethod, KeyServerRoute, KeyServerStatus},
	}

	ServerSentBytesView = &view.View{
		Name:        "opencensus.io/http/server/sent_bytes",
		Description: "Size distribution of HTTP response body",
		Measure:     ServerSentBytes,
		Aggregation: DefaultSizeDistribution,
		TagKeys:     []tag.Key{KeyServerMethod, KeyServerRoute, KeyServerStatus},
	}

	ServerLatencyView = &view.View{
		Name:        "opencensus.io/http/server/server_latency",
		Description: "Latency distribution of HTTP requests",
		Measure:     ServerLatency,
		Aggregation: DefaultLatencyDistribution,
		TagKeys:     []tag.Key{KeyServerMethod, KeyServerRoute, KeyServerStatus},
	}

	ServerCompletedCountView = &view.View{
		Name:        "opencensus.io/http/server/completed_count",
		Description: "Server request count by HTTP method",
		Measure:     ServerLatency,
		Aggregation: view.Count(),
		TagKeys:     []tag.Key{KeyServerMethod, KeyServerRoute, KeyServerStatus},
	}
)

// DefaultClientViews are the default client views provided by this package.
var DefaultClientViews = []*view.View{
	ClientSentBytesView,
	ClientReceivedBytesView,
	ClientRoundtripLatencyView,
	ClientCompletedCountView,
}

// DefaultServerViews are the default server views provided by this package.
var DefaultServerViews = []*view.View{
	ServerReceivedBytesView,
	ServerSentBytesView,
	ServerLatencyView,
	ServerCompletedCountView,
}
