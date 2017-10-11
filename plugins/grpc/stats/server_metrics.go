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

// These variables define the default hard-coded metrics to collect for a GRPC
// server.
// TODO(acetechnologist): This is temporary and will need to be replaced by a
// mechanism to load these defaults from a common repository/config shared by
// all supported languages. Likely a serialized protobuf of these defaults.
var (
	// Default server measures
	RPCServerErrorCount        *istats.MeasureInt64
	RPCServerServerElapsedTime *istats.MeasureFloat64
	RPCServerRequestBytes      *istats.MeasureInt64
	RPCServerResponseBytes     *istats.MeasureInt64
	RPCServerStartedCount      *istats.MeasureInt64
	RPCServerFinishedCount     *istats.MeasureInt64
	RPCServerRequestCount      *istats.MeasureInt64
	RPCServerResponseCount     *istats.MeasureInt64

	// Default server views
	RPCServerErrorCountView        *istats.View
	RPCServerServerElapsedTimeView *istats.View
	RPCServerRequestBytesView      *istats.View
	RPCServerResponseBytesView     *istats.View
	RPCServerRequestCountView      *istats.View
	RPCServerResponseCountView     *istats.View

	RPCServerServerElapsedTimeMinuteView *istats.View
	RPCServerRequestBytesMinuteView      *istats.View
	RPCServerResponseBytesMinuteView     *istats.View
	RPCServerErrorCountMinuteView        *istats.View
	RPCServerStartedCountMinuteView      *istats.View
	RPCServerFinishedCountMinuteView     *istats.View
	RPCServerRequestCountMinuteView      *istats.View
	RPCServerResponseCountMinuteView     *istats.View

	RPCServerServerElapsedTimeHourView *istats.View
	RPCServerRequestBytesHourView      *istats.View
	RPCServerResponseBytesHourView     *istats.View
	RPCServerErrorCountHourView        *istats.View
	RPCServerStartedCountHourView      *istats.View
	RPCServerFinishedCountHourView     *istats.View
	RPCServerRequestCountHourView      *istats.View
	RPCServerResponseCountHourView     *istats.View
)

func createDefaultMeasuresServer() {
	var err error

	// Creating server measures
	if RPCServerErrorCount, err = istats.NewMeasureInt64("/grpc.io/server/error_count", "RPC Errors", unitCount); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/error_count. %v", err))
	}
	if RPCServerServerElapsedTime, err = istats.NewMeasureFloat64("/grpc.io/server/server_elapsed_time", "Server elapsed time in msecs", unitMillisecond); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/server_elapsed_time. %v", err))
	}
	if RPCServerRequestBytes, err = istats.NewMeasureInt64("/grpc.io/server/request_bytes", "Request bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/request_bytes. %v", err))
	}
	if RPCServerResponseBytes, err = istats.NewMeasureInt64("/grpc.io/server/response_bytes", "Response bytes", unitByte); err != nil {
		panic(fmt.Sprintf("createDefaultMeasuresServer failed for measure /grpc.io/server/response_bytes. %v", err))
	}
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
	var views []*istats.View

	RPCServerErrorCountView = istats.NewView("grpc.io/server/error_count/distribution_cumulative", "RPC Errors", []tags.Key{keyMethod, keyOpStatus, keyService}, RPCServerErrorCount, aggCount, windowCumulative)
	views = append(views, RPCServerErrorCountView)
	RPCServerServerElapsedTimeView = istats.NewView("grpc.io/server/server_elapsed_time/distribution_cumulative", "Server elapsed time in msecs", []tags.Key{keyService, keyMethod}, RPCServerServerElapsedTime, aggDistMillis, windowCumulative)
	views = append(views, RPCServerServerElapsedTimeView)
	RPCServerRequestBytesView = istats.NewView("grpc.io/server/request_bytes/distribution_cumulative", "Request bytes", []tags.Key{keyService, keyMethod}, RPCServerRequestBytes, aggDistBytes, windowCumulative)
	views = append(views, RPCServerRequestBytesView)
	RPCServerResponseBytesView = istats.NewView("grpc.io/server/response_bytes/distribution_cumulative", "Response bytes", []tags.Key{keyService, keyMethod}, RPCServerResponseBytes, aggDistBytes, windowCumulative)
	views = append(views, RPCServerResponseBytesView)
	RPCServerRequestCountView = istats.NewView("grpc.io/server/request_count/distribution_cumulative", "Count of request messages per server RPC", []tags.Key{keyService, keyMethod}, RPCServerRequestCount, aggDistCounts, windowCumulative)
	views = append(views, RPCServerRequestCountView)
	RPCServerResponseCountView = istats.NewView("grpc.io/server/response_count/distribution_cumulative", "Count of response messages per server RPC", []tags.Key{keyService, keyMethod}, RPCServerResponseCount, aggDistCounts, windowCumulative)
	views = append(views, RPCServerResponseCountView)

	RPCServerServerElapsedTimeMinuteView = istats.NewView("grpc.io/server/server_elapsed_time/minute_interval", "Minute stats for server elapsed time in msecs", []tags.Key{keyService, keyMethod}, RPCServerServerElapsedTime, aggDistMillis, windowSlidingMinute)
	views = append(views, RPCServerServerElapsedTimeMinuteView)
	RPCServerRequestBytesMinuteView = istats.NewView("grpc.io/server/request_bytes/minute_interval", "Minute stats for request size in bytes", []tags.Key{keyService, keyMethod}, RPCServerRequestBytes, aggCount, windowSlidingMinute)
	views = append(views, RPCServerRequestBytesMinuteView)
	RPCServerResponseBytesMinuteView = istats.NewView("grpc.io/server/response_bytes/minute_interval", "Minute stats for response size in bytes", []tags.Key{keyService, keyMethod}, RPCServerResponseBytes, aggCount, windowSlidingMinute)
	views = append(views, RPCServerResponseBytesMinuteView)
	RPCServerErrorCountMinuteView = istats.NewView("grpc.io/server/error_count/minute_interval", "Minute stats for rpc errors", []tags.Key{keyService, keyMethod}, RPCServerErrorCount, aggCount, windowSlidingMinute)
	views = append(views, RPCServerErrorCountMinuteView)
	RPCServerStartedCountMinuteView = istats.NewView("grpc.io/server/started_count/minute_interval", "Minute stats on the number of server RPCs started", []tags.Key{keyService, keyMethod}, RPCServerStartedCount, aggCount, windowSlidingMinute)
	views = append(views, RPCServerStartedCountMinuteView)
	RPCServerFinishedCountMinuteView = istats.NewView("grpc.io/server/finished_count/minute_interval", "Minute stats on the number of server RPCs finished", []tags.Key{keyService, keyMethod}, RPCServerFinishedCount, aggCount, windowSlidingMinute)
	views = append(views, RPCServerFinishedCountMinuteView)
	RPCServerRequestCountMinuteView = istats.NewView("grpc.io/server/request_count/minute_interval", "Minute stats on the count of request messages per server RPC", []tags.Key{keyService, keyMethod}, RPCServerRequestCount, aggCount, windowSlidingMinute)
	views = append(views, RPCServerRequestCountMinuteView)
	RPCServerResponseCountMinuteView = istats.NewView("grpc.io/server/response_count/minute_interval", "Minute stats on the count of response messages per server RPC", []tags.Key{keyService, keyMethod}, RPCServerResponseCount, aggCount, windowSlidingMinute)
	views = append(views, RPCServerResponseCountMinuteView)

	RPCServerServerElapsedTimeHourView = istats.NewView("grpc.io/server/server_elapsed_time/hour_interval", "Hour stats for server elapsed time in msecs", []tags.Key{keyService, keyMethod}, RPCServerServerElapsedTime, aggDistMillis, windowSlidingHour)
	views = append(views, RPCServerServerElapsedTimeHourView)
	RPCServerRequestBytesHourView = istats.NewView("grpc.io/server/request_bytes/hour_interval", "Hour stats for request size in bytes", []tags.Key{keyService, keyMethod}, RPCServerRequestBytes, aggCount, windowSlidingHour)
	views = append(views, RPCServerRequestBytesHourView)
	RPCServerResponseBytesHourView = istats.NewView("grpc.io/server/response_bytes/hour_interval", "Hour stats for response size in bytes", []tags.Key{keyService, keyMethod}, RPCServerResponseBytes, aggCount, windowSlidingHour)
	views = append(views, RPCServerResponseBytesHourView)
	RPCServerErrorCountHourView = istats.NewView("grpc.io/server/error_count/hour_interval", "Hour stats for rpc errors", []tags.Key{keyService, keyMethod}, RPCServerErrorCount, aggCount, windowSlidingHour)
	views = append(views, RPCServerErrorCountHourView)
	RPCServerStartedCountHourView = istats.NewView("grpc.io/server/started_count/hour_interval", "Hour stats on the number of server RPCs started", []tags.Key{keyService, keyMethod}, RPCServerStartedCount, aggCount, windowSlidingHour)
	views = append(views, RPCServerStartedCountHourView)
	RPCServerFinishedCountHourView = istats.NewView("grpc.io/server/finished_count/hour_interval", "Hour stats on the number of server RPCs finished", []tags.Key{keyService, keyMethod}, RPCServerFinishedCount, aggCount, windowSlidingHour)
	views = append(views, RPCServerFinishedCountHourView)
	RPCServerRequestCountHourView = istats.NewView("grpc.io/server/request_count/hour_interval", "Hour stats on the count of request messages per server RPC", []tags.Key{keyService, keyMethod}, RPCServerRequestCount, aggCount, windowSlidingHour)
	views = append(views, RPCServerRequestCountHourView)
	RPCServerResponseCountHourView = istats.NewView("grpc.io/server/response_count/hour_interval", "Hour stats on the count of response messages per server RPC", []tags.Key{keyService, keyMethod}, RPCServerResponseCount, aggCount, windowSlidingHour)
	views = append(views, RPCServerResponseCountHourView)

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

// registerDefaultsServer registers the default metrics (measures and views)
// for a GRPC server.
func registerDefaultsServer() {
	grpcServerConnKey = &grpcInstrumentationKey{}
	grpcServerRPCKey = &grpcInstrumentationKey{}

	createDefaultKeys()

	createDefaultMeasuresServer()

	registerDefaultViewsServer()
}
