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

package stats_test

import (
	"log"
	"time"

	"golang.org/x/net/context"

	"github.com/census-instrumentation/opencensus-go/stats"
)

func Example_record() {
	m, err := stats.NewMeasureInt64("my.org/measure/openconns", "open connections", "")
	if err != nil {
		log.Fatal(err)
	}

	stats.Record(context.TODO(), m.M(124)) // Record 124 open connections.
}

func Example_view() {
	m, err := stats.NewMeasureInt64("my.org/measure/openconns", "open connections", "")
	if err != nil {
		log.Fatal(err)
	}

	agg := stats.DistributionAggregation([]float64{0, 1000, 2000})
	window := stats.SlidingTimeWindow{Duration: time.Second}
	view := stats.NewView("my.org/views/openconns", "open connections distribution over one second time window", nil, m, agg, window)

	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}

	// Use stats.RegisterExporter to export collected
	// data or force collect.
}
