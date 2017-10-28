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
	"fmt"

	"github.com/census-instrumentation/opencensus-go/stats"
	"github.com/census-instrumentation/opencensus-go/tag"
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
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/error_count. %v", err))
	}
	if RPCClientRoundTripLatency, err = stats.NewMeasureFloat64("/grpc.io/client/roundtrip_latency", "RPC roundtrip latency in msecs", unitMillisecond); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/roundtrip_latency. %v", err))
	}
	if RPCClientRequestBytes, err = stats.NewMeasureInt64("/grpc.io/client/request_bytes", "Request bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/request_bytes. %v", err))
	}
	if RPCClientResponseBytes, err = stats.NewMeasureInt64("/grpc.io/client/response_bytes", "Response bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/response_bytes. %v", err))
	}
	if RPCClientStartedCount, err = stats.NewMeasureInt64("/grpc.io/client/started_count", "Number of client RPCs (streams) started", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure rpc/client/started_count. %v", err))
	}
	if RPCClientFinishedCount, err = stats.NewMeasureInt64("/grpc.io/client/finished_count", "Number of client RPCs (streams) finished", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure /grpc.io/client/finished_count. %v", err))
	}

	if RPCClientRequestCount, err = stats.NewMeasureInt64("/grpc.io/client/request_count", "Number of client RPC request messages", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure rpc/client/request_count. %v", err))
	}
	if RPCClientResponseCount, err = stats.NewMeasureInt64("/grpc.io/client/response_count", "Number of client RPC response messages", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure /grpc.io/client/response_count. %v", err))
	}
}

func defaultClientViews() {
	RPCClientErrorCountView = stats.NewView("grpc.io/client/error_count/distribution_cumulative", "RPC Errors", []tag.Key{keyOpStatus, keyService, keyMethod}, RPCClientErrorCount, aggCount, windowCumulative)
	clientViews = append(clientViews, RPCClientErrorCountView)
	RPCClientRoundTripLatencyView = stats.NewView("grpc.io/client/roundtrip_latency/distribution_cumulative", "Latency in msecs", []tag.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggDistMillis, windowCumulative)
	clientViews = append(clientViews, RPCClientRoundTripLatencyView)
	RPCClientRequestBytesView = stats.NewView("grpc.io/client/request_bytes/distribution_cumulative", "Request bytes", []tag.Key{keyService, keyMethod}, RPCClientRequestBytes, aggDistBytes, windowCumulative)
	clientViews = append(clientViews, RPCClientRequestBytesView)
	RPCClientResponseBytesView = stats.NewView("grpc.io/client/response_bytes/distribution_cumulative", "Response bytes", []tag.Key{keyService, keyMethod}, RPCClientResponseBytes, aggDistBytes, windowCumulative)
	clientViews = append(clientViews, RPCClientResponseBytesView)
	RPCClientRequestCountView = stats.NewView("grpc.io/client/request_count/distribution_cumulative", "Count of request messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientRequestCount, aggDistCounts, windowCumulative)
	clientViews = append(clientViews, RPCClientRequestCountView)
	RPCClientResponseCountView = stats.NewView("grpc.io/client/response_count/distribution_cumulative", "Count of response messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientResponseCount, aggDistCounts, windowCumulative)
	clientViews = append(clientViews, RPCClientResponseCountView)

	RPCClientRoundTripLatencyMinuteView = stats.NewView("grpc.io/client/roundtrip_latency/minute_interval", "Minute stats for latency in msecs", []tag.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggDistMillis, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientRoundTripLatencyMinuteView)
	RPCClientRequestBytesMinuteView = stats.NewView("grpc.io/client/request_bytes/minute_interval", "Minute stats for request size in bytes", []tag.Key{keyService, keyMethod}, RPCClientRequestBytes, aggCount, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientRequestBytesMinuteView)
	RPCClientResponseBytesMinuteView = stats.NewView("grpc.io/client/response_bytes/minute_interval", "Minute stats for response size in bytes", []tag.Key{keyService, keyMethod}, RPCClientResponseBytes, aggCount, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientResponseBytesMinuteView)
	RPCClientErrorCountMinuteView = stats.NewView("grpc.io/client/error_count/minute_interval", "Minute stats for rpc errors", []tag.Key{keyService, keyMethod}, RPCClientErrorCount, aggCount, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientErrorCountMinuteView)
	RPCClientStartedCountMinuteView = stats.NewView("grpc.io/client/started_count/minute_interval", "Minute stats on the number of client RPCs started", []tag.Key{keyService, keyMethod}, RPCClientStartedCount, aggCount, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientStartedCountMinuteView)
	RPCClientFinishedCountMinuteView = stats.NewView("grpc.io/client/finished_count/minute_interval", "Minute stats on the number of client RPCs finished", []tag.Key{keyService, keyMethod}, RPCClientFinishedCount, aggCount, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientFinishedCountMinuteView)
	RPCClientRequestCountMinuteView = stats.NewView("grpc.io/client/request_count/minute_interval", "Minute stats on the count of request messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientRequestCount, aggCount, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientRequestCountMinuteView)
	RPCClientResponseCountMinuteView = stats.NewView("grpc.io/client/response_count/minute_interval", "Minute stats on the count of response messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientResponseCount, aggCount, windowSlidingMinute)
	clientViews = append(clientViews, RPCClientResponseCountMinuteView)

	RPCClientRoundTripLatencyHourView = stats.NewView("grpc.io/client/roundtrip_latency/hour_interval", "Hour stats for latency in msecs", []tag.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggDistMillis, windowSlidingHour)
	clientViews = append(clientViews, RPCClientRoundTripLatencyHourView)
	RPCClientRequestBytesHourView = stats.NewView("grpc.io/client/request_bytes/hour_interval", "Hour stats for request size in bytes", []tag.Key{keyService, keyMethod}, RPCClientRequestBytes, aggCount, windowSlidingHour)
	clientViews = append(clientViews, RPCClientRequestBytesHourView)
	RPCClientResponseBytesHourView = stats.NewView("grpc.io/client/response_bytes/hour_interval", "Hour stats for response size in bytes", []tag.Key{keyService, keyMethod}, RPCClientResponseBytes, aggCount, windowSlidingHour)
	clientViews = append(clientViews, RPCClientResponseBytesHourView)
	RPCClientErrorCountHourView = stats.NewView("grpc.io/client/error_count/hour_interval", "Hour stats for rpc errors", []tag.Key{keyService, keyMethod}, RPCClientErrorCount, aggCount, windowSlidingHour)
	clientViews = append(clientViews, RPCClientErrorCountHourView)
	RPCClientStartedCountHourView = stats.NewView("grpc.io/client/started_count/hour_interval", "Hour stats on the number of client RPCs started", []tag.Key{keyService, keyMethod}, RPCClientStartedCount, aggCount, windowSlidingHour)
	clientViews = append(clientViews, RPCClientStartedCountHourView)
	RPCClientFinishedCountHourView = stats.NewView("grpc.io/client/finished_count/hour_interval", "Hour stats on the number of client RPCs finished", []tag.Key{keyService, keyMethod}, RPCClientFinishedCount, aggCount, windowSlidingHour)
	clientViews = append(clientViews, RPCClientFinishedCountHourView)
	RPCClientRequestCountHourView = stats.NewView("grpc.io/client/request_count/hour_interval", "Hour stats on the count of request messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientRequestCount, aggCount, windowSlidingHour)
	clientViews = append(clientViews, RPCClientRequestCountHourView)
	RPCClientResponseCountHourView = stats.NewView("grpc.io/client/response_count/hour_interval", "Hour stats on the count of response messages per client RPC", []tag.Key{keyService, keyMethod}, RPCClientResponseCount, aggCount, windowSlidingHour)
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
