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

	istats "github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tagging"
)

var (
	// C is the channel where the client code can access the collected views.
	C                         chan *istats.View
	rpcBytesBucketBoundaries  = []float64{0, 1024, 2048, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296}
	rpcMillisBucketBoundaries = []float64{0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000}

	keyMethod    tagging.KeyStringUTF8
	keyOpStatus  tagging.KeyStringUTF8
	bytes        *istats.MeasurementUnit
	milliseconds *istats.MeasurementUnit
	count        *istats.MeasurementUnit
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

func init() {
	createDefaultKeys()
	createDefaultMeasurementUnits()
}
