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

package collection

import (
	"log"
	"time"

	istats "github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tagging"
)

var (
	// C is the channel where the client code can access the collected views.
	C                         chan *istats.View
	rpcBytesBucketBoundaries  = []float64{0, 1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296}
	rpcMillisBucketBoundaries = []float64{0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000}

	keyMethod                          tagging.KeyStringUTF8
	keyOpStatus                        tagging.KeyStringUTF8
	bytes                              *istats.MeasurementUnit
	milliseconds                       *istats.MeasurementUnit
	count                              *istats.MeasurementUnit
	RPCclientErrorCount                istats.MeasureDescFloat64
	RPCclientRequestBytes              istats.MeasureDescFloat64
	RPCclientResponseBytes             istats.MeasureDescFloat64
	RPCclientRoundTripLatency          istats.MeasureDescFloat64
	RPCclientServerElapsedTime         istats.MeasureDescFloat64
	RPCclientUncompressedRequestBytes  istats.MeasureDescFloat64
	RPCclientUncompressedResponseBytes istats.MeasureDescFloat64
	RPCserverErrorCount                istats.MeasureDescFloat64
	RPCserverRequestBytes              istats.MeasureDescFloat64
	RPCserverResponseBytes             istats.MeasureDescFloat64
	RPCserverServerLatency             istats.MeasureDescFloat64
	RPCserverServerElapsedTime         istats.MeasureDescFloat64
	RPCserverUncompressedRequestBytes  istats.MeasureDescFloat64
	RPCserverUncompressedResponseBytes istats.MeasureDescFloat64
	RPCserverStartedCount              istats.MeasureDescFloat64
	RPCserverFinishedCount             istats.MeasureDescFloat64

	RPCclientErrorCountDist                *istats.DistributionViewDesc
	RPCclientRequestBytesDist              *istats.DistributionViewDesc
	RPCclientResponseBytesDist             *istats.DistributionViewDesc
	RPCclientRoundTripLatencyDist          *istats.DistributionViewDesc
	RPCclientServerElapsedTimeDist         *istats.DistributionViewDesc
	RPCclientRequestUncompressedBytesDist  *istats.DistributionViewDesc
	RPCclientResponseUncompressedBytesDist *istats.DistributionViewDesc

	RPCclientErrorCountInterval                *istats.IntervalViewDesc
	RPCclientRoundTripLatencyInterval          *istats.IntervalViewDesc
	RPCclientServerElapsedTimeInterval         *istats.IntervalViewDesc
	RPCclientRequestBytesInterval              *istats.IntervalViewDesc
	RPCclientResponseBytesInterval             *istats.IntervalViewDesc
	RPCclientRequestUncompressedBytesInterval  *istats.IntervalViewDesc
	RPCclientResponseUncompressedBytesInterval *istats.IntervalViewDesc

	RPCserverErrorCountDist                *istats.DistributionViewDesc
	RPCserverRequestBytesDist              *istats.DistributionViewDesc
	RPCserverResponseBytesDist             *istats.DistributionViewDesc
	RPCserverServerLatencyDist             *istats.DistributionViewDesc
	RPCserverServerElapsedTimeDist         *istats.DistributionViewDesc
	RPCserverRequestUncompressedBytesDist  *istats.DistributionViewDesc
	RPCserverResponseUncompressedBytesDist *istats.DistributionViewDesc

	RPCserverErrorCountInterval                *istats.IntervalViewDesc
	RPCserverServerLatencyInterval             *istats.IntervalViewDesc
	RPCserverServerElapsedTimeInterval         *istats.IntervalViewDesc
	RPCserverRequestBytesInterval              *istats.IntervalViewDesc
	RPCserverResponseBytesInterval             *istats.IntervalViewDesc
	RPCserverRequestUncompressedBytesInterval  *istats.IntervalViewDesc
	RPCserverResponseUncompressedBytesInterval *istats.IntervalViewDesc
)

func createDefaultKeys() {
	// Initializing keys
	var err error
	if keyMethod, err = tagging.DefaultKeyManager().CreateKeyStringUTF8("grpc.method"); err != nil {
		log.Fatalf("init() failed to create/retrieve keyStringUTF8. %v", err)
	}
	if keyOpStatus, err = tagging.DefaultKeyManager().CreateKeyStringUTF8("grpc.opStatus"); err != nil {
		log.Fatalf("init() failed to create/retrieve keyStringUTF8. %v", err)
	}
}

func createDefaultMeasurementUnits() {
	// Initializing units
	bytes = &istats.MeasurementUnit{
		Power10:    1,
		Numerators: []istats.BasicUnit{istats.BytesUnit},
	}
	count = &istats.MeasurementUnit{
		Power10:    1,
		Numerators: []istats.BasicUnit{istats.ScalarUnit},
	}
	milliseconds = &istats.MeasurementUnit{
		Power10:    -3,
		Numerators: []istats.BasicUnit{istats.SecsUnit},
	}
}

func registerClientDefaultMeasures() {
	// Creating client measures
	RPCclientErrorCount = istats.NewMeasureDescFloat64("/rpc/client/error_count", "RPC Errors", count)
	RPCclientRequestBytes = istats.NewMeasureDescFloat64("/rpc/client/request_bytes", "Request bytes", bytes)
	RPCclientResponseBytes = istats.NewMeasureDescFloat64("/rpc/client/response_bytes", "Response bytes", bytes)
	RPCclientRoundTripLatency = istats.NewMeasureDescFloat64("/rpc/client/roundtrip_latency", "RPC roundtrip latency in msecs", milliseconds)
	RPCclientServerElapsedTime = istats.NewMeasureDescFloat64("/rpc/client/server_elapsed_time", "Server elapsed time in msecs", milliseconds)
	RPCclientUncompressedRequestBytes = istats.NewMeasureDescFloat64("/rpc/client/uncompressed_request_bytes", "Uncompressed Request bytes", bytes)
	RPCclientUncompressedResponseBytes = istats.NewMeasureDescFloat64("/rpc/client/uncompressed_response_bytes", "Uncompressed Response bytes", bytes)
	// Creating server measures
	RPCserverErrorCount = istats.NewMeasureDescFloat64("/rpc/server/error_count", "RPC Errors", count)
	RPCserverRequestBytes = istats.NewMeasureDescFloat64("/rpc/server/request_bytes", "Request bytes", bytes)
	RPCserverResponseBytes = istats.NewMeasureDescFloat64("/rpc/server/response_bytes", "Response bytes", bytes)
	RPCserverServerLatency = istats.NewMeasureDescFloat64("/rpc/server/server_latency", "Latency in msecs", milliseconds)
	RPCserverServerElapsedTime = istats.NewMeasureDescFloat64("/rpc/server/server_elapsed_time", "Server elapsed time in msecs", milliseconds)
	RPCserverUncompressedRequestBytes = istats.NewMeasureDescFloat64("/rpc/server/uncompressed_request_bytes", "Uncompressed Request bytes", bytes)
	RPCserverUncompressedResponseBytes = istats.NewMeasureDescFloat64("/rpc/server/uncompressed_response_bytes", "Uncompressed Response bytes", bytes)
	RPCserverStartedCount = istats.NewMeasureDescFloat64("/rpc/server/started_count", "Number of RPCs started", count)
	RPCserverFinishedCount = istats.NewMeasureDescFloat64("/rpc/server/finished_count", "Number of RPCs finished", count)

	// Registering client/server measures
	var measures []istats.MeasureDesc
	measures = append(measures, RPCclientErrorCount)
	measures = append(measures, RPCclientRequestBytes)
	measures = append(measures, RPCclientResponseBytes)
	measures = append(measures, RPCclientRoundTripLatency)
	measures = append(measures, RPCclientServerElapsedTime)
	measures = append(measures, RPCclientUncompressedRequestBytes)  //Not needed?
	measures = append(measures, RPCclientUncompressedResponseBytes) //Not needed?

	for _, m := range measures {
		if err := istats.RegisterMeasureDesc(m); err != nil {
			log.Fatalf("init() failed to register %v.\n %v", m, err)
		}
	}
}

func registerServerDefaultMeasures() {
	// Creating server measures
	RPCserverErrorCount = istats.NewMeasureDescFloat64("/rpc/server/error_count", "RPC Errors", count)
	RPCserverRequestBytes = istats.NewMeasureDescFloat64("/rpc/server/request_bytes", "Request bytes", bytes)
	RPCserverResponseBytes = istats.NewMeasureDescFloat64("/rpc/server/response_bytes", "Response bytes", bytes)
	RPCserverServerLatency = istats.NewMeasureDescFloat64("/rpc/server/server_latency", "Latency in msecs", milliseconds)
	RPCserverServerElapsedTime = istats.NewMeasureDescFloat64("/rpc/server/server_elapsed_time", "Server elapsed time in msecs", milliseconds)
	RPCserverUncompressedRequestBytes = istats.NewMeasureDescFloat64("/rpc/server/uncompressed_request_bytes", "Uncompressed Request bytes", bytes)
	RPCserverUncompressedResponseBytes = istats.NewMeasureDescFloat64("/rpc/server/uncompressed_response_bytes", "Uncompressed Response bytes", bytes)
	RPCserverStartedCount = istats.NewMeasureDescFloat64("/rpc/server/started_count", "Number of RPCs started", count)
	RPCserverFinishedCount = istats.NewMeasureDescFloat64("/rpc/server/finished_count", "Number of RPCs finished", count)

	// Registering client/server measures
	var measures []istats.MeasureDesc
	measures = append(measures, RPCserverErrorCount)
	measures = append(measures, RPCserverRequestBytes)
	measures = append(measures, RPCserverResponseBytes)
	measures = append(measures, RPCserverServerLatency)             // difference between serverLatency and ServerElapsedTime?
	measures = append(measures, RPCserverServerElapsedTime)         // difference between serverLatency and ServerElapsedTime?
	measures = append(measures, RPCserverUncompressedRequestBytes)  //Not needed?
	measures = append(measures, RPCserverUncompressedResponseBytes) //Not needed?
	measures = append(measures, RPCserverStartedCount)
	measures = append(measures, RPCserverFinishedCount)

	for _, m := range measures {
		if err := istats.RegisterMeasureDesc(m); err != nil {
			log.Fatalf("init() failed to register %v.\n %v", m, err)
		}
	}
}

func registerClientDefaultViews() {
	// Creating client distributions views
	RPCclientErrorCountDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/error_count/distribution_cumulative",
			Description:     "RPC Errors",
			MeasureDescName: "/rpc/client/error_count",
			TagKeys:         []tagging.Key{keyMethod, keyOpStatus},
		},
		Bounds: []float64{},
	}

	RPCclientRequestBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/request_bytes/distribution_cumulative",
			Description:     "Request bytes",
			MeasureDescName: "/rpc/client/request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}

	RPCclientResponseBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/response_bytes/distribution_cumulative",
			Description:     "Response bytes",
			MeasureDescName: "/rpc/client/response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}

	RPCclientRoundTripLatencyDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/roundtrip_latency/distribution_cumulative",
			Description:     "Latency in msecs",
			MeasureDescName: "/rpc/client/roundtrip_latency",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcMillisBucketBoundaries,
	}

	RPCclientServerElapsedTimeDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/server_elapsed_time/distribution_cumulative",
			Description:     "Server elapsed time in msecs",
			MeasureDescName: "/rpc/client/server_elapsed_time",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcMillisBucketBoundaries,
	}

	RPCclientRequestUncompressedBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/uncompressed_request_bytes/distribution_cumulative",
			Description:     "Request bytes",
			MeasureDescName: "/rpc/client/uncompressed_request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}

	RPCclientResponseUncompressedBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/uncompressed_response_bytes/distribution_cumulative",
			Description:     "Response bytes",
			MeasureDescName: "/rpc/client/uncompressed_response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}

	// Creating client intervals views
	RPCclientErrorCountInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/error_count/interval",
			Description:     "Minute and Hour stats for rpc errors",
			MeasureDescName: "/rpc/client/error_count",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCclientRoundTripLatencyInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/roundtrip_latency/interval",
			Description:     "Minute and Hour stats for latency in msecs",
			MeasureDescName: "/rpc/client/roundtrip_latency",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCclientServerElapsedTimeInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/server_elapsed_time/interval",
			Description:     "Minute and Hour stats for server elapsed time in msecs",
			MeasureDescName: "/rpc/client/server_elapsed_time",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCclientRequestBytesInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/request_bytes/interval",
			Description:     "Minute and Hour stats for request size in bytes",
			MeasureDescName: "/rpc/client/request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCclientResponseBytesInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/response_bytes/interval",
			Description:     "Minute and Hour stats for response size in bytes",
			MeasureDescName: "/rpc/client/response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCclientRequestUncompressedBytesInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/uncompressed_request_bytes/interval",
			Description:     "Minute and Hour stats for uncompressed request size in bytes",
			MeasureDescName: "/rpc/client/uncompressed_request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCclientResponseUncompressedBytesInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/uncompressed_response_bytes/interval",
			Description:     "Minute and Hour stats for uncompressed response size in bytes",
			MeasureDescName: "/rpc/client/uncompressed_response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	// Registering client/server views
	var views []istats.ViewDesc
	C = make(chan *istats.View, 1024)
	views = append(views, RPCclientErrorCountDist)
	views = append(views, RPCclientRequestBytesDist)
	views = append(views, RPCclientResponseBytesDist)
	views = append(views, RPCclientRoundTripLatencyDist)
	views = append(views, RPCclientServerElapsedTimeDist)
	views = append(views, RPCclientRequestUncompressedBytesDist)
	views = append(views, RPCclientResponseUncompressedBytesDist)
	views = append(views, RPCclientErrorCountInterval)
	views = append(views, RPCclientRoundTripLatencyInterval)
	views = append(views, RPCclientServerElapsedTimeInterval)
	views = append(views, RPCclientRequestBytesInterval)
	views = append(views, RPCclientResponseBytesInterval)
	views = append(views, RPCclientRequestUncompressedBytesInterval)
	views = append(views, RPCclientResponseUncompressedBytesInterval)

	for _, v := range views {
		if err := istats.RegisterViewDesc(v); err != nil {
			log.Fatalf("init() failed to register %v.\n %v", v, err)
		}
	}
}

func registerServerDefaultViews() {
	// Creating server distributions views
	RPCserverErrorCountDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/error_count/distribution_cumulative",
			Description:     "RPC Errors",
			MeasureDescName: "/rpc/server/error_count",
			TagKeys:         []tagging.Key{keyMethod, keyOpStatus},
		},
		Bounds: []float64{},
	}

	RPCserverRequestBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/request_bytes/distribution_cumulative",
			Description:     "Request bytes",
			MeasureDescName: "/rpc/server/request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}

	RPCserverResponseBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/response_bytes/distribution_cumulative",
			Description:     "Response bytes",
			MeasureDescName: "/rpc/server/response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}

	RPCserverServerLatencyDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/server_latency/distribution_cumulative",
			Description:     "Latency in msecs",
			MeasureDescName: "/rpc/server/server_latency",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcMillisBucketBoundaries,
	}

	RPCserverServerElapsedTimeDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/server_elapsed_time/distribution_cumulative",
			Description:     "Server elapsed time in msecs",
			MeasureDescName: "/rpc/server/server_elapsed_time",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcMillisBucketBoundaries,
	}

	RPCserverRequestUncompressedBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/uncompressed_request_bytes/distribution_cumulative",
			Description:     "Request bytes",
			MeasureDescName: "/rpc/server/uncompressed_request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}

	RPCserverResponseUncompressedBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/uncompressed_response_bytes/distribution_cumulative",
			Description:     "Response bytes",
			MeasureDescName: "/rpc/server/uncompressed_response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}

	// Creating server intervals views
	RPCserverErrorCountInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/error_count/interval",
			Description:     "Minute and Hour stats for rpc errors",
			MeasureDescName: "/rpc/server/error_count",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCserverServerLatencyInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/server_latency/interval",
			Description:     "Minute and Hour stats for latency in msecs",
			MeasureDescName: "/rpc/server/server_latency",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCserverServerElapsedTimeInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/server_elapsed_time/interval",
			Description:     "Minute and Hour stats for server elapsed time in msecs",
			MeasureDescName: "/rpc/server/server_elapsed_time",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCserverRequestBytesInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/request_bytes/interval",
			Description:     "Minute and Hour stats for request size in bytes",
			MeasureDescName: "/rpc/server/request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCserverResponseBytesInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/response_bytes/interval",
			Description:     "Minute and Hour stats for response size in bytes",
			MeasureDescName: "/rpc/server/response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCserverRequestUncompressedBytesInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/uncompressed_request_bytes/interval",
			Description:     "Minute and Hour stats for uncompressed request size in bytes",
			MeasureDescName: "/rpc/server/uncompressed_request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	RPCserverResponseUncompressedBytesInterval = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/uncompressed_response_bytes/interval",
			Description:     "Minute and Hour stats for uncompressed response size in bytes",
			MeasureDescName: "/rpc/server/uncompressed_response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
	}

	// Registering server/server views
	var views []istats.ViewDesc
	C = make(chan *istats.View, 1024)
	views = append(views, RPCserverErrorCountDist)
	views = append(views, RPCserverRequestBytesDist)
	views = append(views, RPCserverResponseBytesDist)
	views = append(views, RPCserverServerLatencyDist)
	views = append(views, RPCserverServerElapsedTimeDist)
	views = append(views, RPCserverRequestUncompressedBytesDist)
	views = append(views, RPCserverResponseUncompressedBytesDist)
	views = append(views, RPCserverErrorCountInterval)
	views = append(views, RPCserverServerLatencyInterval)
	views = append(views, RPCserverServerElapsedTimeInterval)
	views = append(views, RPCserverRequestBytesInterval)
	views = append(views, RPCserverResponseBytesInterval)
	views = append(views, RPCserverRequestUncompressedBytesInterval)
	views = append(views, RPCserverResponseUncompressedBytesInterval)

	for _, v := range views {
		if err := istats.RegisterViewDesc(v); err != nil {
			log.Fatalf("init() failed to register %v.\n %v", v, err)
		}
	}
}

func RegisterClientDefaults() {
	registerClientDefaultMeasures()
	registerClientDefaultViews()
}

func RegisterServerDefaults() {
	registerServerDefaultMeasures()
	registerServerDefaultViews()
}

func init() {
	createDefaultKeys()
	createDefaultMeasurementUnits()
}
