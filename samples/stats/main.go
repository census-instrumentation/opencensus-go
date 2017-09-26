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

package main

import (
	"fmt"
	"time"

	"github.com/census-instrumentation/opencensus-go/stats"
	"github.com/census-instrumentation/opencensus-go/tags"
	"golang.org/x/net/context"
)

func main() {
	// ---------------------------------------------
	// CREATING/REGISTERING KEYS AND VIEWS
	// ---------------------------------------------

	// Creates keys
	key1, err := tags.CreateKeyString("keyNameID1")
	if err != nil {
		panic(fmt.Sprintf("Key 'keyNameID1' not created %v", err))
	}
	key2, err := tags.CreateKeyString("keyNameID2")
	if err != nil {
		panic(fmt.Sprintf("Key 'keyNameID2' not created %v", err))
	}

	// Create measures
	mf, err := stats.NewMeasureFloat64("/my/float64/measureName", "some measure", "ms")
	if err != nil {
		panic(fmt.Sprintf("Measure '/my/float64/measureName' not created %v", err))
	}
	mi, err := stats.NewMeasureInt64("/my/int64/otherName", "some other measure", "By")
	if err != nil {
		panic(fmt.Sprintf("Measure '/my/int64/otherName' not created %v", err))
	}

	// Create aggregations
	histogramBounds := []float64{-10, 0, 10, 20}
	agg1 := stats.NewAggregationDistribution(histogramBounds)
	agg2 := stats.NewAggregationCount()

	duration := 10 * time.Second
	precisionIntervals := 10
	wnd1 := stats.NewWindowSlidingTime(duration, precisionIntervals)

	// Create views
	myView1 := stats.NewView("/my/int64/viewName", "some description", []tags.Key{key1, key2}, mf, agg1, wnd1)
	myView2 := stats.NewView("/my/float64/viewName", "some other description", []tags.Key{key1}, mi, agg2, wnd1)

	// Register views
	err = stats.RegisterView(myView1)
	if err != nil {
		panic(fmt.Sprintf("View %v not registered. %v", myView1, err))
	}
	err = stats.RegisterView(myView2)
	if err != nil {
		panic(fmt.Sprintf("View %v not registered. %v", myView2, err))
	}

	// set the reporting period to 1 second instead of the 10 seconds default
	reporitngDuration := 1 * time.Second
	stats.SetReportingPeriod(reporitngDuration)

	// Subscribe to view
	c1 := make(chan *stats.ViewData, 4)

	// Process collected data asynchronously
	go func(c1 chan *stats.ViewData) {
		for vd := range c1 {
			fmt.Printf("\nViewData collected for view %v received after default duration elapsed. %v row(s) received\n", vd.V.Name(), len(vd.Rows))
			for _, r := range vd.Rows {
				fmt.Printf("row received with len(tags): %v\n", len(r.Tags))
			}
		}
	}(c1)

	err = stats.SubscribeToView(myView1, c1)
	if err != nil {
		panic(fmt.Sprintf("Subscription to view %v failed. %v", myView1, err))
	}

	// Explicitly instruct the library to collect the view data for on-demand
	// retrieval
	if err := stats.ForceCollection(myView2); err != nil {
		panic(fmt.Sprintf("Forced collection of view %v failed. %v", myView2, err))
	}

	// ---------------
	// RECORDING USAGE
	// ---------------
	// Adding tags to context
	newTagSet := tags.NewTagSetBuilder(nil).UpsertString(key1, "foo1").
		UpsertString(key2, "foo2").
		Build()
	ctx := tags.NewContext(context.Background(), newTagSet)

	// Recording single datapoint at a time
	stats.RecordInt64(ctx, mi, 1)
	stats.RecordFloat64(ctx, mf, 10.0)

	// Recording multiple datapoints at once
	stats.Record(ctx, mi.Is(4), mf.Is(10.0))

	// Wait for a duration longer than reporting duration to ensure the census
	// library reports the collected data
	fmt.Printf("\nWait %v for default reporting duration to kick in\n", reporitngDuration+1*time.Second)
	time.Sleep(reporitngDuration + 1*time.Second)

	// Pull collected data synchronously from the library
	rows, err := stats.RetrieveData(myView2)
	if err != nil {
		panic(fmt.Sprintf("Retrieving data from view %v failed. %v", myView2, err))
	}

	// Process collected data on-demand
	fmt.Printf("\nViewData collected for view %v received on demand. %v row(s) received\n", myView2.Name(), len(rows))
	for _, r := range rows {
		fmt.Printf("row received with len(tags): %v\n", len(r.Tags))
	}

	fmt.Printf("\n")
}
