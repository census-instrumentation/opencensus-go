// Copyright 2017 Google Inc.
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
// a GRPC client. These are Go objects instances mirroring the proto
// definitions found at "github.com/google/instrumentation-proto/census.proto".
// A complete description of each can be found there.
// TODO(acetechnologist): This is temporary and will need to be replaced by a
// mechanism to load these defaults from a common repository/config shared by
// all supported languages. Likely a serialized protobuf of these defaults.
var (
	RPCClientErrorCount       *istats.MeasureInt64
	RPCClientRoundTripLatency *istats.MeasureFloat64
	// RPCClientServerElapsedTime *istats.MeasureFloat64
	RPCClientRequestBytes  *istats.MeasureInt64
	RPCClientResponseBytes *istats.MeasureInt64
	// RPCClientUncompressedRequestBytes  *istats.MeasureInt64
	// RPCClientUncompressedResponseBytes *istats.MeasureInt64
	RPCClientStartedCount  *istats.MeasureInt64
	RPCClientFinishedCount *istats.MeasureInt64
	RPCClientRequestCount  *istats.MeasureInt64
	RPCClientResponseCount *istats.MeasureInt64

	RPCClientErrorCountView       istats.View
	RPCClientRoundTripLatencyView istats.View
	// RPCClientServerElapsedTimeView istats.View
	RPCClientRequestBytesView  istats.View
	RPCClientResponseBytesView istats.View
	// RPCClientRequestUncompressedBytesView  istats.View
	// RPCClientResponseUncompressedBytesView istats.View
	RPCClientRequestCountView  istats.View
	RPCClientResponseCountView istats.View

	RPCClientRoundTripLatencyMinuteView istats.View
	RPCClientRequestBytesMinuteView     istats.View
	RPCClientResponseBytesMinuteView    istats.View
	RPCClientErrorCountMinuteView       istats.View
	// RPCClientRequestUncompressedBytesMinuteView  istats.View
	// RPCClientResponseUncompressedBytesMinuteView istats.View
	// RPCClientServerElapsedTimeMinuteView istats.View
	RPCClientStartedCountMinuteView  istats.View
	RPCClientFinishedCountMinuteView istats.View
	RPCClientRequestCountMinuteView  istats.View
	RPCClientResponseCountMinuteView istats.View

	RPCClientRoundTripLatencyHourView istats.View
	RPCClientRequestBytesHourView     istats.View
	RPCClientResponseBytesHourView    istats.View
	RPCClientErrorCountHourView       istats.View
	// RPCClientRequestUncompressedBytesHourView  istats.View
	// RPCClientResponseUncompressedBytesHourView istats.View
	// RPCClientServerElapsedTimeHourView istats.View
	RPCClientStartedCountHourView  istats.View
	RPCClientFinishedCountHourView istats.View
	RPCClientRequestCountHourView  istats.View
	RPCClientResponseCountHourView istats.View
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
	// if RPCClientServerElapsedTime, err = istats.NewMeasureFloat64("/grpc.io/client/server_elapsed_time", "Server elapsed time in msecs", unitMillisecond); err != nil {
	// 	panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/server_elapsed_time. %v", err))
	// }
	if RPCClientRequestBytes, err = istats.NewMeasureInt64("/grpc.io/client/request_bytes", "Request bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/request_bytes. %v", err))
	}
	if RPCClientResponseBytes, err = istats.NewMeasureInt64("/grpc.io/client/response_bytes", "Response bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/response_bytes. %v", err))
	}
	// if RPCClientUncompressedRequestBytes, err = istats.NewMeasureInt64("/grpc.io/client/uncompressed_request_bytes", "Uncompressed Request bytes", unitByte); err != nil {
	// 	panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/uncompressed_request_bytes. %v", err))
	// }
	// if RPCClientUncompressedResponseBytes, err = istats.NewMeasureInt64("/grpc.io/client/uncompressed_response_bytes", "Uncompressed Response bytes", unitByte); err != nil {
	// 	panic(fmt.Sprintf("createDefaultMeasuresClient failed for measure grpc.io/client/uncompressed_response_bytes. %v", err))
	// }
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
	var views []istats.View

	RPCClientErrorCountView = istats.NewView("grpc.io/client/error_count/distribution_cumulative", "RPC Errors", []tags.Key{keyOpStatus, keyService, keyMethod}, RPCClientErrorCount, aggCount, windowCumulative)
	views = append(views, RPCClientErrorCountView)
	RPCClientRoundTripLatencyView = istats.NewView("grpc.io/client/roundtrip_latency/distribution_cumulative", "Latency in msecs", []tags.Key{keyService, keyMethod}, RPCClientRoundTripLatency, aggDistMillis, windowCumulative)
	views = append(views, RPCClientRoundTripLatencyView)
	// RPCClientServerElapsedTimeView = istats.NewView("grpc.io/client/server_elapsed_time/distribution_cumulative", "Server elapsed time in msecs", []tags.Key{keyService, keyMethod}, RPCClientServerElapsedTime, aggDistMillis, windowCumulative)
	// views = append(views, RPCClientServerElapsedTimeView)
	RPCClientRequestBytesView = istats.NewView("grpc.io/client/request_bytes/distribution_cumulative", "Request bytes", []tags.Key{keyService, keyMethod}, RPCClientRequestBytes, aggDistBytes, windowCumulative)
	views = append(views, RPCClientRequestBytesView)
	RPCClientResponseBytesView = istats.NewView("grpc.io/client/response_bytes/distribution_cumulative", "Response bytes", []tags.Key{keyService, keyMethod}, RPCClientResponseBytes, aggDistBytes, windowCumulative)
	views = append(views, RPCClientResponseBytesView)
	// RPCClientRequestUncompressedBytesView
	// views = append(views, RPCClientRequestUncompressedBytesView)
	// RPCClientResponseUncompressedBytesView
	// views = append(views, RPCClientResponseUncompressedBytesView)
	RPCClientRequestCountView = istats.NewView("grpc.io/client/request_count/distribution_cumulative", "Count of request messages per client RPC", []tags.Key{keyService, keyMethod}, RPCClientRequestCount, aggDistCounts, windowCumulative)
	views = append(views, RPCClientRequestCountView)
	RPCClientResponseCountView = istats.NewView("grpc.io/client/response_count/distribution_cumulative", "Count of response messages per client RPC", []tags.Key{keyService, keyMethod}, RPCClientResponseCount, aggDistCounts, windowCumulative)
	views = append(views, RPCClientResponseCountView)

	// RPCClientRoundTripLatencyMinuteView
	// views = append(views, RPCClientRoundTripLatencyMinuteView)
	// RPCClientRequestBytesMinuteView
	// views = append(views, RPCClientRequestBytesMinuteView)
	// RPCClientResponseBytesMinuteView
	// views = append(views, RPCClientResponseBytesMinuteView)
	// RPCClientErrorCountMinuteView
	// views = append(views, RPCClientErrorCountMinuteView)
	// // RPCClientRequestUncompressedBytesMinuteView
	// // views = append(views, RPCClientRequestUncompressedBytesMinuteView)
	// // RPCClientResponseUncompressedBytesMinuteView
	// // views = append(views, RPCClientResponseUncompressedBytesMinuteView)
	// RPCClientServerElapsedTimeMinuteView
	// views = append(views, RPCClientServerElapsedTimeMinuteView)
	// RPCClientStartedCountMinuteView
	// views = append(views, RPCClientStartedCountMinuteView)
	// RPCClientFinishedCountMinuteView
	// views = append(views, RPCClientFinishedCountMinuteView)
	// RPCClientRequestCountMinuteView
	// views = append(views, RPCClientRequestCountMinuteView)
	// RPCClientResponseCountMinuteView
	// views = append(views, RPCClientResponseCountMinuteView)

	// RPCClientRoundTripLatencyHourView
	// views = append(views, RPCClientRoundTripLatencyHourView)
	// RPCClientRequestBytesHourView
	// views = append(views, RPCClientRequestBytesHourView)
	// RPCClientResponseBytesHourView
	// views = append(views, RPCClientResponseBytesHourView)
	// RPCClientErrorCountHourView
	// views = append(views, RPCClientErrorCountHourView)
	// // RPCClientRequestUncompressedBytesHourView
	// // views = append(views, RPCClientRequestUncompressedBytesHourView)
	// // RPCClientResponseUncompressedBytesHourView
	// // views = append(views, RPCClientResponseUncompressedBytesHourView)
	// RPCClientServerElapsedTimeHourView
	// views = append(views, RPCClientServerElapsedTimeHourView)
	// RPCClientStartedCountHourView
	// views = append(views, RPCClientStartedCountHourView)
	// RPCClientFinishedCountHourView
	// views = append(views, RPCClientFinishedCountHourView)
	// RPCClientRequestCountHourView
	// views = append(views, RPCClientRequestCountHourView)
	// RPCClientResponseCountHourView
	// views = append(views, RPCClientResponseCountHourView)

	// Registering views
	for _, v := range views {
		if err := istats.RegisterView(v); err != nil {
			log.Fatalf("init() failed to register %v.%v\n", v, err)
		}
		if err := istats.ForceCollection(v); err != nil {
			log.Fatalf("init() failed to ForceCollection %v.%v\n", v, err)
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
