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

package stats

import (
	"fmt"
	"log"

	istats "github.com/census-instrumentation/opencensus-go/stats"
	"github.com/census-instrumentation/opencensus-go/tags"
)

// The following variables define the default hard-coded metrics to collect for
// a GRPC client.
// TODO(acetechnologist): This is temporary and will need to be replaced by a
// mechanism to load these defaults from a common repository/config shared by
// all supported languages. Likely a serialized protobuf of these defaults.
var (
	// Default client measures
	RPCClientErrorCount       *istats.MeasureInt64
	RPCClientRoundTripLatency *istats.MeasureFloat64
	RPCClientRequestBytes     *istats.MeasureInt64
	RPCClientResponseBytes    *istats.MeasureInt64
	RPCClientStartedCount     *istats.MeasureInt64
	RPCClientFinishedCount    *istats.MeasureInt64
	RPCClientRequestCount     *istats.MeasureInt64
	RPCClientResponseCount    *istats.MeasureInt64

	// Default client views
	RPCClientErrorCountView       *istats.View
	RPCClientRoundTripLatencyView *istats.View
	RPCClientRequestBytesView     *istats.View
	RPCClientResponseBytesView    *istats.View
	RPCClientRequestCountView     *istats.View
	RPCClientResponseCountView    *istats.View

	RPCClientRoundTripLatencyMinuteView *istats.View
	RPCClientRequestBytesMinuteView     *istats.View
	RPCClientResponseBytesMinuteView    *istats.View
	RPCClientErrorCountMinuteView       *istats.View
	RPCClientStartedCountMinuteView     *istats.View
	RPCClientFinishedCountMinuteView    *istats.View
	RPCClientRequestCountMinuteView     *istats.View
	RPCClientResponseCountMinuteView    *istats.View

	RPCClientRoundTripLatencyHourView *istats.View
	RPCClientRequestBytesHourView     *istats.View
	RPCClientResponseBytesHourView    *istats.View
	RPCClientErrorCountHourView       *istats.View
	RPCClientStartedCountHourView     *istats.View
	RPCClientFinishedCountHourView    *istats.View
	RPCClientRequestCountHourView     *istats.View
	RPCClientResponseCountHourView    *istats.View
)

func createDefaultMeasuresClient() {
	var err error

	// Creating client measures
	if RPCClientErrorCount, err = istats.NewMeasureInt64("/grpc.io/client/error_count", "RPC Errors", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/error_count. %v", err))
	}
	if RPCClientRoundTripLatency, err = istats.NewMeasureFloat64("/grpc.io/client/roundtrip_latency", "RPC roundtrip latency in msecs", unitMillisecond); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/roundtrip_latency. %v", err))
	}
	if RPCClientRequestBytes, err = istats.NewMeasureInt64("/grpc.io/client/request_bytes", "Request bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/request_bytes. %v", err))
	}
	if RPCClientResponseBytes, err = istats.NewMeasureInt64("/grpc.io/client/response_bytes", "Response bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/response_bytes. %v", err))
	}
	if RPCClientStartedCount, err = istats.NewMeasureInt64("/grpc.io/client/started_count", "Number of client RPCs (streams) started", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure rpc/client/started_count. %v", err))
	}
	if RPCClientFinishedCount, err = istats.NewMeasureInt64("/grpc.io/client/finished_count", "Number of client RPCs (streams) finished", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure /grpc.io/client/finished_count. %v", err))
	}

	if RPCClientRequestCount, err = istats.NewMeasureInt64("/grpc.io/client/request_count", "Number of client RPC request messages", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure rpc/client/request_count. %v", err))
	}
	if RPCClientResponseCount, err = istats.NewMeasureInt64("/grpc.io/client/response_count", "Number of client RPC response messages", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure /grpc.io/client/response_count. %v", err))
	}
}

func registerDefaultViewsClient() {
	var views []*istats.View

	RPCClientErrorCountView = istats.NewView("grpc.io/client/error_count/distribution_cumulative", "RPC Errors", []tags.Key{keyOpStatus, keyService, keyMethod}, RPCClientErrorCount, aggCount, windowCumulative)
	views = append(views, RPCClientErrorCountView)
	RPCClientRoundTripLatencyView = istats.NewView("grpc.io/client/roundtrip_latency/distribution_cumulative", "Latency in msecs", []tags.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggDistMillis, windowCumulative)
	views = append(views, RPCClientRoundTripLatencyView)
	RPCClientRequestBytesView = istats.NewView("grpc.io/client/request_bytes/distribution_cumulative", "Request bytes", []tags.Key{keyService, keyMethod}, RPCClientRequestBytes, aggDistBytes, windowCumulative)
	views = append(views, RPCClientRequestBytesView)
	RPCClientResponseBytesView = istats.NewView("grpc.io/client/response_bytes/distribution_cumulative", "Response bytes", []tags.Key{keyService, keyMethod}, RPCClientResponseBytes, aggDistBytes, windowCumulative)
	views = append(views, RPCClientResponseBytesView)
	RPCClientRequestCountView = istats.NewView("grpc.io/client/request_count/distribution_cumulative", "Count of request messages per client RPC", []tags.Key{keyService, keyMethod}, RPCClientRequestCount, aggDistCounts, windowCumulative)
	views = append(views, RPCClientRequestCountView)
	RPCClientResponseCountView = istats.NewView("grpc.io/client/response_count/distribution_cumulative", "Count of response messages per client RPC", []tags.Key{keyService, keyMethod}, RPCClientResponseCount, aggDistCounts, windowCumulative)
	views = append(views, RPCClientResponseCountView)

	RPCClientRoundTripLatencyMinuteView = istats.NewView("grpc.io/client/roundtrip_latency/minute_interval", "Minute stats for latency in msecs", []tags.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggDistMillis, windowSlidingMinute)
	views = append(views, RPCClientRoundTripLatencyMinuteView)
	RPCClientRequestBytesMinuteView = istats.NewView("grpc.io/client/request_bytes/minute_interval", "Minute stats for request size in bytes", []tags.Key{keyService, keyMethod}, RPCClientRequestBytes, aggCount, windowSlidingMinute)
	views = append(views, RPCClientRequestBytesMinuteView)
	RPCClientResponseBytesMinuteView = istats.NewView("grpc.io/client/response_bytes/minute_interval", "Minute stats for response size in bytes", []tags.Key{keyService, keyMethod}, RPCClientResponseBytes, aggCount, windowSlidingMinute)
	views = append(views, RPCClientResponseBytesMinuteView)
	RPCClientErrorCountMinuteView = istats.NewView("grpc.io/client/error_count/minute_interval", "Minute stats for rpc errors", []tags.Key{keyService, keyMethod}, RPCClientErrorCount, aggCount, windowSlidingMinute)
	views = append(views, RPCClientErrorCountMinuteView)
	RPCClientStartedCountMinuteView = istats.NewView("grpc.io/client/started_count/minute_interval", "Minute stats on the number of client RPCs started", []tags.Key{keyService, keyMethod}, RPCClientStartedCount, aggCount, windowSlidingMinute)
	views = append(views, RPCClientStartedCountMinuteView)
	RPCClientFinishedCountMinuteView = istats.NewView("grpc.io/client/finished_count/minute_interval", "Minute stats on the number of client RPCs finished", []tags.Key{keyService, keyMethod}, RPCClientFinishedCount, aggCount, windowSlidingMinute)
	views = append(views, RPCClientFinishedCountMinuteView)
	RPCClientRequestCountMinuteView = istats.NewView("grpc.io/client/request_count/minute_interval", "Minute stats on the count of request messages per client RPC", []tags.Key{keyService, keyMethod}, RPCClientRequestCount, aggCount, windowSlidingMinute)
	views = append(views, RPCClientRequestCountMinuteView)
	RPCClientResponseCountMinuteView = istats.NewView("grpc.io/client/response_count/minute_interval", "Minute stats on the count of response messages per client RPC", []tags.Key{keyService, keyMethod}, RPCClientResponseCount, aggCount, windowSlidingMinute)
	views = append(views, RPCClientResponseCountMinuteView)

	RPCClientRoundTripLatencyHourView = istats.NewView("grpc.io/client/roundtrip_latency/hour_interval", "Hour stats for latency in msecs", []tags.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggDistMillis, windowSlidingHour)
	views = append(views, RPCClientRoundTripLatencyHourView)
	RPCClientRequestBytesHourView = istats.NewView("grpc.io/client/request_bytes/hour_interval", "Hour stats for request size in bytes", []tags.Key{keyService, keyMethod}, RPCClientRequestBytes, aggCount, windowSlidingHour)
	views = append(views, RPCClientRequestBytesHourView)
	RPCClientResponseBytesHourView = istats.NewView("grpc.io/client/response_bytes/hour_interval", "Hour stats for response size in bytes", []tags.Key{keyService, keyMethod}, RPCClientResponseBytes, aggCount, windowSlidingHour)
	views = append(views, RPCClientResponseBytesHourView)
	RPCClientErrorCountHourView = istats.NewView("grpc.io/client/error_count/hour_interval", "Hour stats for rpc errors", []tags.Key{keyService, keyMethod}, RPCClientErrorCount, aggCount, windowSlidingHour)
	views = append(views, RPCClientErrorCountHourView)
	RPCClientStartedCountHourView = istats.NewView("grpc.io/client/started_count/hour_interval", "Hour stats on the number of client RPCs started", []tags.Key{keyService, keyMethod}, RPCClientStartedCount, aggCount, windowSlidingHour)
	views = append(views, RPCClientStartedCountHourView)
	RPCClientFinishedCountHourView = istats.NewView("grpc.io/client/finished_count/hour_interval", "Hour stats on the number of client RPCs finished", []tags.Key{keyService, keyMethod}, RPCClientFinishedCount, aggCount, windowSlidingHour)
	views = append(views, RPCClientFinishedCountHourView)
	RPCClientRequestCountHourView = istats.NewView("grpc.io/client/request_count/hour_interval", "Hour stats on the count of request messages per client RPC", []tags.Key{keyService, keyMethod}, RPCClientRequestCount, aggCount, windowSlidingHour)
	views = append(views, RPCClientRequestCountHourView)
	RPCClientResponseCountHourView = istats.NewView("grpc.io/client/response_count/hour_interval", "Hour stats on the count of response messages per client RPC", []tags.Key{keyService, keyMethod}, RPCClientResponseCount, aggCount, windowSlidingHour)
	views = append(views, RPCClientResponseCountHourView)

	// Registering views
	for _, v := range views {
		if err := istats.RegisterView(v); err != nil {
			log.Fatalf("init() failed to register %v: %v.\n", v, err)
		}
		if err := v.ForceCollect(); err != nil {
			log.Fatalf("init() failed to ForceCollect %v: %v.\n", v, err)
		}
	}
}

// registerDefaultsClient registers the default metrics (measures and views)
// for a GRPC client.
func registerDefaultsClient() {
	grpcClientConnKey = &grpcInstrumentationKey{}
	grpcClientRPCKey = &grpcInstrumentationKey{}

	createDefaultKeys()

	createDefaultMeasuresClient()

	registerDefaultViewsClient()
}
