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

var (
	RPCclientErrorCount                istats.MeasureDescFloat64
	RPCclientRequestBytes              istats.MeasureDescFloat64
	RPCclientResponseBytes             istats.MeasureDescFloat64
	RPCclientRoundTripLatency          istats.MeasureDescFloat64
	RPCclientServerElapsedTime         istats.MeasureDescFloat64
	RPCclientUncompressedRequestBytes  istats.MeasureDescFloat64
	RPCclientUncompressedResponseBytes istats.MeasureDescFloat64
	// Do not make sense.. Distributed over what.
	// RPCclientErrorCountDist                *istats.DistributionViewDesc
	RPCclientRequestBytesDist              *istats.DistributionViewDesc
	RPCclientResponseBytesDist             *istats.DistributionViewDesc
	RPCclientRoundTripLatencyDist          *istats.DistributionViewDesc
	RPCclientServerElapsedTimeDist         *istats.DistributionViewDesc
	RPCclientRequestUncompressedBytesDist  *istats.DistributionViewDesc
	RPCclientResponseUncompressedBytesDist *istats.DistributionViewDesc
	RPCclientErrorCountInterval            *istats.IntervalViewDesc
	// Do not make sense. Summming and counting elapsed_times
	// RPCclientRoundTripLatencyInterval      *istats.IntervalViewDesc
	// RPCclientServerElapsedTimeInterval         *istats.IntervalViewDesc
	RPCclientRequestBytesInterval              *istats.IntervalViewDesc
	RPCclientResponseBytesInterval             *istats.IntervalViewDesc
	RPCclientRequestUncompressedBytesInterval  *istats.IntervalViewDesc
	RPCclientResponseUncompressedBytesInterval *istats.IntervalViewDesc
)

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

func registerClientDefaultViews() {
	// Creating client distributions views
	var views []istats.ViewDesc
	C = make(chan *istats.View, 1024)
	/*
		RPCclientErrorCountDist = &istats.DistributionViewDesc{
			Vdc: &istats.ViewDescCommon{
				Name:            "/rpc/client/error_count/distribution_cumulative",
				Description:     "RPC Errors",
				MeasureDescName: "/rpc/client/error_count",
				TagKeys:         []tagging.Key{keyMethod, keyOpStatus},
			},
			Bounds: []float64{},
		}
		views = append(views, RPCclientErrorCountDist)
	*/
	RPCclientRequestBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/request_bytes/distribution_cumulative",
			Description:     "Request bytes",
			MeasureDescName: "/rpc/client/request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}
	views = append(views, RPCclientRequestBytesDist)

	RPCclientResponseBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/response_bytes/distribution_cumulative",
			Description:     "Response bytes",
			MeasureDescName: "/rpc/client/response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}
	views = append(views, RPCclientResponseBytesDist)

	RPCclientRoundTripLatencyDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/roundtrip_latency/distribution_cumulative",
			Description:     "Latency in msecs",
			MeasureDescName: "/rpc/client/roundtrip_latency",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcMillisBucketBoundaries,
	}
	views = append(views, RPCclientRoundTripLatencyDist)

	RPCclientServerElapsedTimeDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/server_elapsed_time/distribution_cumulative",
			Description:     "Server elapsed time in msecs",
			MeasureDescName: "/rpc/client/server_elapsed_time",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcMillisBucketBoundaries,
	}
	views = append(views, RPCclientServerElapsedTimeDist)

	RPCclientRequestUncompressedBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/uncompressed_request_bytes/distribution_cumulative",
			Description:     "Request bytes",
			MeasureDescName: "/rpc/client/uncompressed_request_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}
	views = append(views, RPCclientRequestUncompressedBytesDist)

	RPCclientResponseUncompressedBytesDist = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "/rpc/client/uncompressed_response_bytes/distribution_cumulative",
			Description:     "Response bytes",
			MeasureDescName: "/rpc/client/uncompressed_response_bytes",
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcBytesBucketBoundaries,
	}
	views = append(views, RPCclientResponseUncompressedBytesDist)

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
	views = append(views, RPCclientErrorCountInterval)

	/*
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
		views = append(views, RPCclientRoundTripLatencyInterval)

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
		views = append(views, RPCclientServerElapsedTimeInterval)
	*/
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
	views = append(views, RPCclientRequestBytesInterval)

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
	views = append(views, RPCclientResponseBytesInterval)

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
	views = append(views, RPCclientRequestUncompressedBytesInterval)

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
	views = append(views, RPCclientResponseUncompressedBytesInterval)

	// Registering client/server views
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
