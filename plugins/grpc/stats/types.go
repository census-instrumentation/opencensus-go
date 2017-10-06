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

// Package stats defines the handlers to collect grpc statistics using the
// opencensus library.
package stats

import (
	"log"
	"time"

	istats "github.com/census-instrumentation/opencensus-go/stats"
	"github.com/census-instrumentation/opencensus-go/tags"
)

type grpcInstrumentationKey struct{}

// rpcData holds the instrumentation RPC data that is needed between the start
// and end of an call. It holds the info that this package needs to keep track
// of between the various GRPC events.
type rpcData struct {
	// startTime represents the time at which TagRPC was invoked at the
	// beginning of an RPC. It is an appoximation of the time when the
	// application code invoked GRPC code.
	startTime           time.Time
	reqCount, respCount uint64
}

// The following variables define the default hard-coded auxiliary data used by
// both the default GRPC client and GRPC server metrics.
// These are Go objects instances mirroring the some of the proto definitions
// found at "github.com/google/instrumentation-proto/census.proto".
// A complete description of each can be found there.
// TODO(acetechnologist): This is temporary and will need to be replaced by a
// mechanism to load these defaults from a common repository/config shared by
// all supported languages. Likely a serialized protobuf of these defaults.
var (
	// C is the channel where the client code can access the collected views.
	C chan *istats.ViewData

	unitByte             = "By"
	unitCount            = "1"
	unitMillisecond      = "ms"
	slidingTimeSubuckets = 6

	rpcBytesBucketBoundaries  = []float64{0, 1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296}
	rpcMillisBucketBoundaries = []float64{0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000}
	rpcCountBucketBoundaries  = []float64{0, 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536}

	aggCount      = istats.NewAggregationCount()
	aggDistBytes  = istats.NewAggregationDistribution(rpcBytesBucketBoundaries)
	aggDistMillis = istats.NewAggregationDistribution(rpcMillisBucketBoundaries)
	aggDistCounts = istats.NewAggregationDistribution(rpcCountBucketBoundaries)

	windowCumulative    = istats.NewWindowCumulative()
	windowSlidingHour   = istats.NewWindowSlidingTime(1*time.Hour, 6)
	windowSlidingMinute = istats.NewWindowSlidingTime(1*time.Minute, 6)

	keyService  *tags.KeyString
	keyMethod   *tags.KeyString
	keyOpStatus *tags.KeyString
)

func createDefaultKeys() {
	// Initializing keys
	var err error
	if keyService, err = tags.KeyStringByName("grpc.service"); err != nil {
		log.Fatalf("tags.KeyStringByName(\"grpc.service\") failed to create/retrieve keyService. %v", err)
	}

	if keyMethod, err = tags.KeyStringByName("grpc.method"); err != nil {
		log.Fatalf("tags.KeyStringByName(\"grpc.method\") failed to create/retrieve keyMethod. %v", err)
	}

	if keyOpStatus, err = tags.KeyStringByName("grpc.opstatus"); err != nil {
		log.Fatalf("tags.KeyStringByName(\"grpc.opstatus\") failed to create/retrieve keyOpStatus. %v", err)
	}
}

func init() {
	registerDefaultsServer()
	registerDefaultsClient()
}
