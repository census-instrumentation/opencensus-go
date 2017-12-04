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

	RPCClientRoundTripLatencyMinuteView *stats.View
	RPCClientRequestBytesMinuteView     *stats.View
	RPCClientResponseBytesMinuteView    *stats.View
	RPCClientErrorCountMinuteView       *stats.View
	RPCClientStartedCountMinuteView     *stats.View
	RPCClientFinishedCountMinuteView    *stats.View
	RPCClientRequestCountMinuteView     *stats.View
	RPCClientResponseCountMinuteView    *stats.View

	RPCClientRoundTripLatencyHourView *stats.View
	RPCClientRequestBytesHourView     *stats.View
	RPCClientResponseBytesHourView    *stats.View
	RPCClientErrorCountHourView       *stats.View
	RPCClientStartedCountHourView     *stats.View
	RPCClientFinishedCountHourView    *stats.View
	RPCClientRequestCountHourView     *stats.View
	RPCClientResponseCountHourView    *stats.View
)

// TODO(acetechnologist): This is temporary and will need to be replaced by a
// mechanism to load these defaults from a common repository/config shared by
// all supported languages. Likely a serialized protobuf of these defaults.

func defaultClientMeasures() {
	var err error

	// Creating client measures
	if RPCClientErrorCount, err = stats.NewMeasureInt64("/grpc.io/client/error_count", "RPC Errors", unitCount); err != nil {
		log.Fatalf("Cannot create measure /grpc.io/client/error_count: %v", err)
	}
	if RPCClientRoundTripLatency, err = stats.NewMeasureFloat64("/grpc.io/client/roundtrip_latency", "RPC roundtrip latency in msecs", unitMillisecond); err != nil {
		log.Fatalf("Cannot create measure /grpc.io/client/roundtrip_latency: %v", err)
	}
	if RPCClientRequestBytes, err = stats.NewMeasureInt64("/grpc.io/client/request_bytes", "Request bytes", unitByte); err != nil {
		log.Fatalf("Cannot create measure /grpc.io/client/request_bytes: %v", err)
	}
	if RPCClientResponseBytes, err = stats.NewMeasureInt64("/grpc.io/client/response_bytes", "Response bytes", unitByte); err != nil {
		log.Fatalf("Cannot create measure /grpc.io/client/response_bytes: %v", err)
	}
	if RPCClientStartedCount, err = stats.NewMeasureInt64("/grpc.io/client/started_count", "Number of client RPCs (streams) started", unitCount); err != nil {
		log.Fatalf("Cannot create measure /grpc.io/client/started_count: %v", err)
	}
	if RPCClientFinishedCount, err = stats.NewMeasureInt64("/grpc.io/client/finished_count", "Number of client RPCs (streams) finished", unitCount); err != nil {
		log.Fatalf("Cannot create measure /grpc.io/client/finished_count: %v", err)
	}
	if RPCClientRequestCount, err = stats.NewMeasureInt64("/grpc.io/client/request_count", "Number of client RPC request messages", unitCount); err != nil {
		log.Fatalf("Cannot create measure /grpc.io/client/request_count: %v", err)
	}
	if RPCClientResponseCount, err = stats.NewMeasureInt64("/grpc.io/client/response_count", "Number of client RPC response messages", unitCount); err != nil {
		log.Fatalf("Cannot create measure /grpc.io/client/response_count: %v", err)
	}
}

func defaultClientViews() {
	// Use the Java implementation as a reference at
	// https://github.com/census-instrumentation/opencensus-java/blob/2b464864e3dd3f80e8e4c9dc72fccc225444a939/contrib/grpc_metrics/src/main/java/io/opencensus/contrib/grpc/metrics/RpcViewConstants.java#L113-L658
	RPCClientErrorCountView, _ = stats.NewView("grpc.io/client/error_count/cumulative", "RPC Errors", []tag.Key{keyOpStatus, keyService, keyMethod}, RPCClientErrorCount, aggMean, windowCumulative)
	clientViews = append(clientViews, RPCClientErrorCountView)
	RPCClientRoundTripLatencyView, _ = stats.NewView("grpc.io/client/roundtrip_latency/cumulative", "Latency in msecs", []tag.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggDistMillis, windowCumulative)
	clientViews = append(clientViews, RPCClientRoundTripLatencyView)
	RPCClientRequestBytesView, _ = stats.NewView("grpc.io/client/request_bytes/cumulative", "Request bytes", []tag.Key{keyService, keyMethod}, RPCClientRequestBytes, aggDistBytes, windowCumulative)
	clientViews = append(clientViews, RPCClientRequestBytesView)
	RPCClientResponseBytesView, _ = stats.NewView("grpc.io/client/response_bytes/cumulative", "Response bytes", []tag.Key{keyService, keyMethod}, RPCClientResponseBytes, aggDistBytes, windowCumulative)
	clientViews = append(clientViews, RPCClientResponseBytesView)
	RPCClientRequestCountView, _ = stats.NewView("grpc.io/client/request_count/cumulative", "Count of request messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientRequestCount, aggDistCounts, windowCumulative)
	clientViews = append(clientViews, RPCClientRequestCountView)
	RPCClientResponseCountView, _ = stats.NewView("grpc.io/client/response_count/cumulative", "Count of response messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientResponseCount, aggDistCounts, windowCumulative)
	clientViews = append(clientViews, RPCClientResponseCountView)

	RPCClientRoundTripLatencyMinuteView, _ = stats.NewView("grpc.io/client/roundtrip_latency/minute", "Minute stats for latency in msecs", []tag.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggMean, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientRoundTripLatencyMinuteView)
	RPCClientRequestBytesMinuteView, _ = stats.NewView("grpc.io/client/request_bytes/minute", "Minute stats for request size in bytes", []tag.Key{keyService, keyMethod}, RPCClientRequestBytes, aggMean, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientRequestBytesMinuteView)
	RPCClientResponseBytesMinuteView, _ = stats.NewView("grpc.io/client/response_bytes/minute", "Minute stats for response size in bytes", []tag.Key{keyService, keyMethod}, RPCClientResponseBytes, aggMean, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientResponseBytesMinuteView)
	RPCClientErrorCountMinuteView, _ = stats.NewView("grpc.io/client/error_count/minute", "Minute stats for rpc errors", []tag.Key{keyService, keyMethod}, RPCClientErrorCount, aggMean, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientErrorCountMinuteView)
	RPCClientStartedCountMinuteView, _ = stats.NewView("grpc.io/client/started_count/minute", "Minute stats on the number of client RPCs started", []tag.Key{keyService, keyMethod}, RPCClientStartedCount, aggMean, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientStartedCountMinuteView)
	RPCClientFinishedCountMinuteView, _ = stats.NewView("grpc.io/client/finished_count/minute", "Minute stats on the number of client RPCs finished", []tag.Key{keyService, keyMethod}, RPCClientFinishedCount, aggMean, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientFinishedCountMinuteView)
	RPCClientRequestCountMinuteView, _ = stats.NewView("grpc.io/client/request_count/minute", "Minute stats on the count of request messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientRequestCount, aggMean, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientRequestCountMinuteView)
	RPCClientResponseCountMinuteView, _ = stats.NewView("grpc.io/client/response_count/minute", "Minute stats on the count of response messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientResponseCount, aggMean, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientResponseCountMinuteView)

	RPCClientRoundTripLatencyHourView, _ = stats.NewView("grpc.io/client/roundtrip_latency/hour", "Hour stats for latency in msecs", []tag.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggMean, windowSlidingHour)
	clientViews = append(clientViews, RPCClientRoundTripLatencyHourView)
	RPCClientRequestBytesHourView, _ = stats.NewView("grpc.io/client/request_bytes/hour", "Hour stats for request size in bytes", []tag.Key{keyService, keyMethod}, RPCClientRequestBytes, aggMean, windowSlidingHour)
	clientViews = append(clientViews, RPCClientRequestBytesHourView)
	RPCClientResponseBytesHourView, _ = stats.NewView("grpc.io/client/response_bytes/hour", "Hour stats for response size in bytes", []tag.Key{keyService, keyMethod}, RPCClientResponseBytes, aggMean, windowSlidingHour)
	clientViews = append(clientViews, RPCClientResponseBytesHourView)
	RPCClientErrorCountHourView, _ = stats.NewView("grpc.io/client/error_count/hour", "Hour stats for rpc errors", []tag.Key{keyService, keyMethod}, RPCClientErrorCount, aggMean, windowSlidingHour)
	clientViews = append(clientViews, RPCClientErrorCountHourView)
	RPCClientStartedCountHourView, _ = stats.NewView("grpc.io/client/started_count/hour", "Hour stats on the number of client RPCs started", []tag.Key{keyService, keyMethod}, RPCClientStartedCount, aggMean, windowSlidingHour)
	clientViews = append(clientViews, RPCClientStartedCountHourView)
	RPCClientFinishedCountHourView, _ = stats.NewView("grpc.io/client/finished_count/hour", "Hour stats on the number of client RPCs finished", []tag.Key{keyService, keyMethod}, RPCClientFinishedCount, aggMean, windowSlidingHour)
	clientViews = append(clientViews, RPCClientFinishedCountHourView)
	RPCClientRequestCountHourView, _ = stats.NewView("grpc.io/client/request_count/hour", "Hour stats on the count of request messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientRequestCount, aggMean, windowSlidingHour)
	clientViews = append(clientViews, RPCClientRequestCountHourView)
	RPCClientResponseCountHourView, _ = stats.NewView("grpc.io/client/response_count/hour", "Hour stats on the count of response messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientResponseCount, aggMean, windowSlidingHour)
	clientViews = append(clientViews, RPCClientResponseCountHourView)
	// TODO(jbd): Register it when constructing the stats handler.
}

// initClient registers the default metrics (measures and views)
// for a GRPC client.
func initClient() {
	defaultClientMeasures()
	defaultClientViews()
}

var clientViews []*stats.View
