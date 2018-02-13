// Copyright 2017, OpenCensus Authors
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
//

package ocgrpc

import (
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

// The following variables are measures and views made available for gRPC clients.
// Client connection needs to use a clientStatsHandler in order to enable collection.
var (
	ClientErrors, _           = stats.Int64("grpc.io/client/error_count", "RPC Errors", stats.UnitNone)
	ClientRequestBytes, _     = stats.Int64("grpc.io/client/request_bytes", "Request bytes", stats.UnitBytes)
	ClientResponseBytes, _    = stats.Int64("grpc.io/client/response_bytes", "Response bytes", stats.UnitBytes)
	ClientStartedCount, _     = stats.Int64("grpc.io/client/started_count", "Number of client RPCs (streams) started", stats.UnitNone)
	ClientFinishedCount, _    = stats.Int64("grpc.io/client/finished_count", "Number of client RPCs (streams) finished", stats.UnitNone)
	ClientRequestCount, _     = stats.Int64("grpc.io/client/request_count", "Number of client RPC request messages", stats.UnitNone)
	ClientResponseCount, _    = stats.Int64("grpc.io/client/response_count", "Number of client RPC response messages", stats.UnitNone)
	ClientRoundTripLatency, _ = stats.Float64("grpc.io/client/roundtrip_latency", "RPC roundtrip latency in msecs", stats.UnitMilliseconds)
)

// Predefined client views
var (
	ClientErrorCountView, _ = view.New(
		"grpc.io/client/error_count",
		"RPC Errors",
		[]tag.Key{KeyStatus, KeyMethod},
		ClientErrors,
		view.MeanAggregation{})
	ClientRoundTripLatencyView, _ = view.New(
		"grpc.io/client/roundtrip_latency",
		"Latency in msecs",
		[]tag.Key{KeyMethod},
		ClientRoundTripLatency,
		DefaultMillisecondsDistribution)
	ClientRequestBytesView, _ = view.New(
		"grpc.io/client/request_bytes",
		"Request bytes",
		[]tag.Key{KeyMethod},
		ClientRequestBytes,
		DefaultBytesDistribution)
	ClientResponseBytesView, _ = view.New(
		"grpc.io/client/response_bytes",
		"Response bytes",
		[]tag.Key{KeyMethod},
		ClientResponseBytes,
		DefaultBytesDistribution)
	ClientRequestCountView, _ = view.New(
		"grpc.io/client/request_count",
		"Count of request messages per client RPC",
		[]tag.Key{KeyMethod},
		ClientRequestCount,
		DefaultMessageCountDistribution)
	ClientResponseCountView, _ = view.New(
		"grpc.io/client/response_count",
		"Count of response messages per client RPC",
		[]tag.Key{KeyMethod},
		ClientResponseCount,
		DefaultMessageCountDistribution)

	DefaultClientViews = []*view.View{
		ClientErrorCountView,
		ClientRoundTripLatencyView,
		ClientRequestBytesView,
		ClientResponseBytesView,
		ClientRequestCountView,
		ClientResponseCountView,
	}
)

// TODO(jbd): Add roundtrip_latency, uncompressed_request_bytes, uncompressed_response_bytes, request_count, response_count.
// TODO(acetechnologist): This is temporary and will need to be replaced by a
// mechanism to load these defaults from a common repository/config shared by
// all supported languages. Likely a serialized protobuf of these defaults.
