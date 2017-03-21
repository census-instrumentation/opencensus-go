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
	"log"
	"time"

	istats "github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tagging"
)

// The following variables define the default hard-coded metrics to collect for
// a GRPC server. These are Go objects instances mirroring the proto
// definitions found at "github.com/google/instrumentation-proto/census.proto".
// A complete description of each can be found there.
// TODO(acetechnologist): This is temporary and will need to be replaced by a
// mechanism to load these defaults from a common repository/config shared by
// all supported languages. Likely a serialized protobuf of these defaults.
var (
	RPCserverErrorCount                istats.MeasureDescFloat64
	RPCserverRequestBytes              istats.MeasureDescFloat64
	RPCserverResponseBytes             istats.MeasureDescFloat64
	RPCserverServerLatency             istats.MeasureDescFloat64
	RPCserverServerElapsedTime         istats.MeasureDescFloat64
	RPCserverUncompressedRequestBytes  istats.MeasureDescFloat64
	RPCserverUncompressedResponseBytes istats.MeasureDescFloat64
	RPCserverStartedCount              istats.MeasureDescFloat64
	RPCserverFinishedCount             istats.MeasureDescFloat64
	// Do not make sense.. Distributed over what.
	// RPCserverErrorCountDist                *istats.DistributionViewDesc
	RPCserverRequestBytesDist              *istats.DistributionViewDesc
	RPCserverResponseBytesDist             *istats.DistributionViewDesc
	RPCserverServerLatencyDist             *istats.DistributionViewDesc
	RPCserverServerElapsedTimeDist         *istats.DistributionViewDesc
	RPCserverRequestUncompressedBytesDist  *istats.DistributionViewDesc
	RPCserverResponseUncompressedBytesDist *istats.DistributionViewDesc
	RPCserverErrorCountInterval            *istats.IntervalViewDesc
	// Do not make sense. Summming and counting elapsed_times
	// RPCserverServerLatencyInterval         *istats.IntervalViewDesc
	// RPCserverServerElapsedTimeInterval         *istats.IntervalViewDesc
	RPCserverRequestBytesInterval              *istats.IntervalViewDesc
	RPCserverResponseBytesInterval             *istats.IntervalViewDesc
	RPCserverRequestUncompressedBytesInterval  *istats.IntervalViewDesc
	RPCserverResponseUncompressedBytesInterval *istats.IntervalViewDesc
)

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

	// Registering server measures
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

func registerServerDefaultViews() {
	// Creating server distributions views
	C = make(chan *istats.View, 1024)
	var views []istats.ViewDesc
	/*
		RPCserverErrorCountDist = &istats.DistributionViewDesc{
			Vdc: &istats.ViewDescCommon{
				Name:            "/rpc/server/error_count/distribution_cumulative",
				Description:     "RPC Errors",
				MeasureDescName: "/rpc/server/error_count",
				TagKeys:         []tagging.Key{keyMethod, keyOpStatus},
			},
			Bounds: []float64{},
		}
		views = append(views, RPCserverErrorCountDist)
	*/
	RPCserverRequestBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/request_bytes/distribution_cumulative",
			Description:     "Request bytes",
			MeasureDescName: "/rpc/server/request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}
	views = append(views, RPCserverRequestBytesDist)

	RPCserverResponseBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/response_bytes/distribution_cumulative",
			Description:     "Response bytes",
			MeasureDescName: "/rpc/server/response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}
	views = append(views, RPCserverResponseBytesDist)

	RPCserverServerLatencyDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/server_latency/distribution_cumulative",
			Description:     "Latency in msecs",
			MeasureDescName: RPCserverServerLatency.Meta().Name(),
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcMillisBucketBoundaries,
	}
	views = append(views, RPCserverServerLatencyDist)

	RPCserverServerElapsedTimeDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/server_elapsed_time/distribution_cumulative",
			Description:     "Server elapsed time in msecs",
			MeasureDescName: RPCserverServerElapsedTime.Meta().Name(),
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcMillisBucketBoundaries,
	}
	views = append(views, RPCserverServerElapsedTimeDist)

	RPCserverRequestUncompressedBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/uncompressed_request_bytes/distribution_cumulative",
			Description:     "Request bytes",
			MeasureDescName: "/rpc/server/uncompressed_request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}
	views = append(views, RPCserverRequestUncompressedBytesDist)

	RPCserverResponseUncompressedBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/server/uncompressed_response_bytes/distribution_cumulative",
			Description:     "Response bytes",
			MeasureDescName: "/rpc/server/uncompressed_response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}
	views = append(views, RPCserverResponseUncompressedBytesDist)

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
	views = append(views, RPCserverErrorCountInterval)

	/*
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
		views = append(views, RPCserverServerLatencyInterval)
		RPCserverServerElapsedTimeInterval = &istats.IntervalViewDesc{
			Vdc: &istats.ViewDescCommon{
				Name:            "/rpc/server/server_elapsed_time/interval",
				Description:     "Minute and Hour stats for server elapsed time in msecs",
				MeasureDescName: RPCserverServerElapsedTime.Meta().Name(),
				TagKeys:         []tagging.Key{keyMethod},
			},
			SubIntervals: 5,
			Intervals:    []time.Duration{time.Minute * 1, time.Hour * 1},
		}
		views = append(views, RPCserverServerElapsedTimeInterval)
	*/
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
	views = append(views, RPCserverRequestBytesInterval)

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
	views = append(views, RPCserverResponseBytesInterval)

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
	views = append(views, RPCserverRequestUncompressedBytesInterval)

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
	views = append(views, RPCserverResponseUncompressedBytesInterval)

	// Registering server/server views
	for _, v := range views {
		if err := istats.RegisterViewDesc(v); err != nil {
			log.Fatalf("init() failed to register %v.\n %v", v, err)
		}
	}
}

// RegisterServerDefaults registers the default metrics (measures and views)
// for a GRPC server.
func RegisterServerDefaults() {
	registerServerDefaultMeasures()
	registerServerDefaultViews()
}
