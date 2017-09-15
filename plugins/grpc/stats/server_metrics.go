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
// a GRPC server. These are Go objects instances mirroring the proto
// definitions found at "github.com/google/instrumentation-proto/census.proto".
// A complete description of each can be found there.
// TODO(acetechnologist): This is temporary and will need to be replaced by a
// mechanism to load these defaults from a common repository/config shared by
// all supported languages. Likely a serialized protobuf of these defaults.
var (
	RPCServerErrorCount *istats.MeasureInt64
	// RPCServerServerLatency *istats.MeasureFloat64
	RPCServerServerElapsedTime *istats.MeasureFloat64
	RPCServerRequestBytes      *istats.MeasureInt64
	RPCServerResponseBytes     *istats.MeasureInt64
	// RPCServerUncompressedRequestBytes  *istats.MeasureInt64
	// RPCServerUncompressedResponseBytes *istats.MeasureInt64
	RPCServerStartedCount  *istats.MeasureInt64
	RPCServerFinishedCount *istats.MeasureInt64
	RPCServerRequestCount  *istats.MeasureInt64
	RPCServerResponseCount *istats.MeasureInt64

	RPCServerErrorCountView istats.View
	// RPCServerServerLatencyView istats.View
	RPCServerServerElapsedTimeView istats.View
	RPCServerRequestBytesView      istats.View
	RPCServerResponseBytesView     istats.View
	// RPCServerRequestUncompressedBytesView  istats.View
	// RPCServerResponseUncompressedBytesView istats.View
	RPCServerRequestCountView  istats.View
	RPCServerResponseCountView istats.View

	// RPCServerServerServerLatencyMinuteView istats.View
	RPCServerRequestBytesMinuteView  istats.View
	RPCServerResponseBytesMinuteView istats.View
	RPCServerErrorCountMinuteView    istats.View
	// RPCServerRequestUncompressedBytesMinuteView  istats.View
	// RPCServerResponseUncompressedBytesMinuteView istats.View
	RPCServerServerElapsedTimeMinuteView   istats.View
	RPCServerStartedCountMinuteMinuteView  istats.View
	RPCServerFinishedCountMinuteMinuteView istats.View
	RPCServerRequestCountMinuteView        istats.View
	RPCServerResponseCountMinuteView       istats.View

	// RPCServerServerServerLatencyHourView istats.View
	RPCServerRequestBytesHourView  istats.View
	RPCServerResponseBytesHourView istats.View
	RPCServerErrorCountHourView    istats.View
	// RPCServerRequestUncompressedBytesHourView  istats.View
	// RPCServerResponseUncompressedBytesHourView istats.View
	RPCServerServerElapsedTimeHourView   istats.View
	RPCServerStartedCountMinuteHourView  istats.View
	RPCServerFinishedCountMinuteHourView istats.View
	RPCServerRequestCountHourView        istats.View
	RPCServerResponseCountHourView       istats.View
)

func createDefaultMeasuresServer() {
	var err error

	// Creating server measures
	if RPCServerErrorCount, err = istats.NewMeasureInt64("/grpc.io/server/error_count", "RPC Errors", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/error_count. %v", err))
	}
	// if RPCServerServerLatency, err = istats.NewMeasureFloat64("/grpc.io/server/server_latency", "Latency in msecs", unitMillisecond); err != nil {
	// 	panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/server_latency. %v", err))
	// }
	if RPCServerServerElapsedTime, err = istats.NewMeasureFloat64("/grpc.io/server/server_elapsed_time", "Server elapsed time in msecs", unitMillisecond); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/server_elapsed_time. %v", err))
	}
	if RPCServerRequestBytes, err = istats.NewMeasureInt64("/grpc.io/server/request_bytes", "Request bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/request_bytes. %v", err))
	}
	if RPCServerResponseBytes, err = istats.NewMeasureInt64("/grpc.io/server/response_bytes", "Response bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/response_bytes. %v", err))
	}
	// if RPCServerUncompressedRequestBytes, err = istats.NewMeasureInt64("/grpc.io/server/uncompressed_request_bytes", "Uncompressed Request bytes",unitByte); err != nil {
	// 	panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/uncompressed_request_bytes. %v", err))
	// }
	// if RPCServerUncompressedResponseBytes, err = istats.NewMeasureInt64("/grpc.io/server/uncompressed_response_bytes", "Uncompressed Response bytes",unitByte); err != nil {
	// 	panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/uncompressed_response_bytes. %v", err))
	// }
	if RPCServerStartedCount, err = istats.NewMeasureInt64("/grpc.io/server/started_count", "Number of server RPCs (streams) started", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure rpc/server/started_count. %v", err))
	}
	if RPCServerFinishedCount, err = istats.NewMeasureInt64("/grpc.io/server/finished_count", "Number of server RPCs (streams) finished", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/finished_count. %v", err))
	}

	if RPCServerRequestCount, err = istats.NewMeasureInt64("/grpc.io/server/request_count", "Number of server RPC request messages", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure rpc/server/request_count. %v", err))
	}
	if RPCServerResponseCount, err = istats.NewMeasureInt64("/grpc.io/server/response_count", "Number of server RPC response messages", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/response_count. %v", err))
	}
}

func registerDefaultViewsServer() {
	var views []istats.View

	RPCServerErrorCountView = istats.NewView("grpc.io/server/error_count/distribution_cumulative", "RPC Errors", []tags.Key{keyMethod, keyOpStatus, keyService}, RPCServerErrorCount, aggCount, windowCumulative)
	views = append(views, RPCServerErrorCountView)
	// RPCServerServerLatencyView = istats.NewView("grpc.io/server/serverserver_latency/distribution_cumulative", "Latency in msecs", []tags.Key{keyService, keyMethod}, RPCServerRoundTripLatency, aggDistMillis, windowCumulative)
	// views = append(views, RPCServerServerLatencyView)
	RPCServerServerElapsedTimeView = istats.NewView("grpc.io/server/server_elapsed_time/distribution_cumulative", "Server elapsed time in msecs", []tags.Key{keyService, keyMethod}, RPCServerServerElapsedTime, aggDistMillis, windowCumulative)
	views = append(views, RPCServerServerElapsedTimeView)
	RPCServerRequestBytesView = istats.NewView("grpc.io/server/request_bytes/distribution_cumulative", "Request bytes", []tags.Key{keyService, keyMethod}, RPCServerRequestBytes, aggDistBytes, windowCumulative)
	views = append(views, RPCServerRequestBytesView)
	RPCServerResponseBytesView = istats.NewView("grpc.io/server/response_bytes/distribution_cumulative", "Response bytes", []tags.Key{keyService, keyMethod}, RPCServerResponseBytes, aggDistBytes, windowCumulative)
	views = append(views, RPCServerResponseBytesView)
	// RPCServerRequestUncompressedBytesView
	// views = append(views, RPCServerRequestUncompressedBytesView)
	// RPCServerResponseUncompressedBytesView
	// views = append(views, RPCServerResponseUncompressedBytesView)
	RPCServerRequestCountView = istats.NewView("grpc.io/server/request_count/distribution_cumulative", "Count of request messages per server RPC", []tags.Key{keyService, keyMethod}, RPCServerRequestCount, aggDistCounts, windowCumulative)
	views = append(views, RPCServerRequestCountView)
	RPCServerResponseCountView = istats.NewView("grpc.io/server/response_count/distribution_cumulative", "Count of response messages per server RPC", []tags.Key{keyService, keyMethod}, RPCServerResponseCount, aggDistCounts, windowCumulative)
	views = append(views, RPCServerResponseCountView)

	// RPCServerRoundTripLatencyMinuteView
	// views = append(views, RPCServerRoundTripLatencyMinuteView)
	// RPCServerRequestBytesMinuteView
	// views = append(views, RPCServerRequestBytesMinuteView)
	// RPCServerResponseBytesMinuteView
	// views = append(views, RPCServerResponseBytesMinuteView)
	// RPCServerErrorCountMinuteView
	// views = append(views, RPCServerErrorCountMinuteView)
	// // RPCServerRequestUncompressedBytesMinuteView
	// // views = append(views, RPCServerRequestUncompressedBytesMinuteView)
	// // RPCServerResponseUncompressedBytesMinuteView
	// // views = append(views, RPCServerResponseUncompressedBytesMinuteView)
	// RPCServerServerElapsedTimeMinuteView
	// views = append(views, RPCServerServerElapsedTimeMinuteView)
	// RPCServerStartedCountMinuteView
	// views = append(views, RPCServerStartedCountMinuteView)
	// RPCServerFinishedCountMinuteView
	// views = append(views, RPCServerFinishedCountMinuteView)
	// RPCServerRequestCountMinuteView
	// views = append(views, RPCServerRequestCountMinuteView)
	// RPCServerResponseCountMinuteView
	// views = append(views, RPCServerResponseCountMinuteView)

	// RPCServerRoundTripLatencyHourView
	// views = append(views, RPCServerRoundTripLatencyHourView)
	// RPCServerRequestBytesHourView
	// views = append(views, RPCServerRequestBytesHourView)
	// RPCServerResponseBytesHourView
	// views = append(views, RPCServerResponseBytesHourView)
	// RPCServerErrorCountHourView
	// views = append(views, RPCServerErrorCountHourView)
	// // RPCServerRequestUncompressedBytesHourView
	// // views = append(views, RPCServerRequestUncompressedBytesHourView)
	// // RPCServerResponseUncompressedBytesHourView
	// // views = append(views, RPCServerResponseUncompressedBytesHourView)
	// RPCServerServerElapsedTimeHourView
	// views = append(views, RPCServerServerElapsedTimeHourView)
	// RPCServerStartedCountHourView
	// views = append(views, RPCServerStartedCountHourView)
	// RPCServerFinishedCountHourView
	// views = append(views, RPCServerFinishedCountHourView)
	// RPCServerRequestCountHourView
	// views = append(views, RPCServerRequestCountHourView)
	// RPCServerResponseCountHourView
	// views = append(views, RPCServerResponseCountHourView)

	// Registering views
	for _, v := range views {
		if err := istats.RegisterView(v); err != nil {
			log.Fatalf("init() failed to register %v.%v\n", v, err)
		}
		if err := istats.ForceCollection(v); err != nil {
			log.Fatalf("init() failed to ForceCollection %v.%v\n", v, err)
		}
	}
	//c = make(chan *istats.View, 1024)
}

// registerDefaultsServer registers the default metrics (measures and views)
// for a GRPC server.
func registerDefaultsServer() {
	grpcServerConnKey = &grpcInstrumentationKey{}
	grpcServerRPCKey = &grpcInstrumentationKey{}

	createDefaultKeys()

	createDefaultMeasuresServer()

	registerDefaultViewsServer()
}
