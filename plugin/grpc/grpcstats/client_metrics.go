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

package grpcstats

import (
	"log"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

// The following variables are measures and views made available for gRPC clients.
// Client connection needs to use a ClientStatsHandler in order to enable collection.
var (
	// Available client measures
	RPCClientErrorCount       *stats.MeasureInt64
	RPCClientRoundTripLatency *stats.MeasureFloat64
	RPCClientRequestBytes     *stats.MeasureInt64
	RPCClientResponseBytes    *stats.MeasureInt64
	RPCClientStartedCount     *stats.MeasureInt64
	RPCClientFinishedCount    *stats.MeasureInt64
	RPCClientRequestCount     *stats.MeasureInt64
	RPCClientResponseCount    *stats.MeasureInt64

	// Predefined client views
	RPCClientErrorCountView       *stats.View
	RPCClientRoundTripLatencyView *stats.View
	RPCClientRequestBytesView     *stats.View
	RPCClientResponseBytesView    *stats.View
	RPCClientRequestCountView     *stats.View
	RPCClientResponseCountView    *stats.View
)

// TODO(acetechnologist): This is temporary and will need to be replaced by a
// mechanism to load these defaults from a common repository/config shared by
// all supported languages. Likely a serialized protobuf of these defaults.

func defaultClientMeasures() {
	var err error

	// Creating client measures
	if RPCClientErrorCount, err = stats.NewMeasureInt64("grpc.io/client/error_count", "RPC Errors", unitCount); err != nil {
		log.Fatalf("Cannot create measure grpc.io/client/error_count: %v", err)
	}
	if RPCClientRoundTripLatency, err = stats.NewMeasureFloat64("grpc.io/client/roundtrip_latency", "RPC roundtrip latency in msecs", unitMillisecond); err != nil {
		log.Fatalf("Cannot create measure grpc.io/client/roundtrip_latency: %v", err)
	}
	if RPCClientRequestBytes, err = stats.NewMeasureInt64("grpc.io/client/request_bytes", "Request bytes", unitByte); err != nil {
		log.Fatalf("Cannot create measure grpc.io/client/request_bytes: %v", err)
	}
	if RPCClientResponseBytes, err = stats.NewMeasureInt64("grpc.io/client/response_bytes", "Response bytes", unitByte); err != nil {
		log.Fatalf("Cannot create measure grpc.io/client/response_bytes: %v", err)
	}
	if RPCClientStartedCount, err = stats.NewMeasureInt64("grpc.io/client/started_count", "Number of client RPCs (streams) started", unitCount); err != nil {
		log.Fatalf("Cannot create measure grpc.io/client/started_count: %v", err)
	}
	if RPCClientFinishedCount, err = stats.NewMeasureInt64("grpc.io/client/finished_count", "Number of client RPCs (streams) finished", unitCount); err != nil {
		log.Fatalf("Cannot create measure grpc.io/client/finished_count: %v", err)
	}
	if RPCClientRequestCount, err = stats.NewMeasureInt64("grpc.io/client/request_count", "Number of client RPC request messages", unitCount); err != nil {
		log.Fatalf("Cannot create measure grpc.io/client/request_count: %v", err)
	}
	if RPCClientResponseCount, err = stats.NewMeasureInt64("grpc.io/client/response_count", "Number of client RPC response messages", unitCount); err != nil {
		log.Fatalf("Cannot create measure grpc.io/client/response_count: %v", err)
	}
}

func defaultClientViews() {
	RPCClientErrorCountView, _ = stats.NewView(
		"grpc.io/client/error_count/cumulative",
		"RPC Errors",
		[]tag.Key{keyStatus, keyMethod},
		RPCClientErrorCount,
		aggMean)
	RPCClientRoundTripLatencyView, _ = stats.NewView(
		"grpc.io/client/roundtrip_latency/cumulative",
		"Latency in msecs",
		[]tag.Key{keyMethod},
		RPCClientRoundTripLatency,
		aggDistMillis)
	RPCClientRequestBytesView, _ = stats.NewView(
		"grpc.io/client/request_bytes/cumulative",
		"Request bytes",
		[]tag.Key{keyMethod},
		RPCClientRequestBytes,
		aggDistBytes)
	RPCClientResponseBytesView, _ = stats.NewView(
		"grpc.io/client/response_bytes/cumulative",
		"Response bytes",
		[]tag.Key{keyMethod},
		RPCClientResponseBytes,
		aggDistBytes)
	RPCClientRequestCountView, _ = stats.NewView(
		"grpc.io/client/request_count/cumulative",
		"Count of request messages per client RPC",
		[]tag.Key{keyMethod},
		RPCClientRequestCount,
		aggDistCounts)
	RPCClientResponseCountView, _ = stats.NewView(
		"grpc.io/client/response_count/cumulative",
		"Count of response messages per client RPC",
		[]tag.Key{keyMethod},
		RPCClientResponseCount,
		aggDistCounts)

	clientViews = append(clientViews,
		RPCClientErrorCountView,
		RPCClientRoundTripLatencyView,
		RPCClientRequestBytesView,
		RPCClientResponseBytesView,
		RPCClientRequestCountView,
		RPCClientResponseCountView,
	)
	// TODO(jbd): Add roundtrip_latency, uncompressed_request_bytes, uncompressed_response_bytes, request_count, response_count.
}

// initClient registers the default metrics (measures and views)
// for a GRPC client.
func initClient() {
	defaultClientMeasures()
	defaultClientViews()
}

var clientViews []*stats.View
