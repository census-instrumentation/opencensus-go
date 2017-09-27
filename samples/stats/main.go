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
	"log"
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
	key1, err := tags.CreateKeyString("/widget.com/key/deviceTypeID")
	if err != nil {
		log.Fatalf("Key '/widget.com/key/deviceTypeID' not created %v", err)
	}
	key2, err := tags.CreateKeyString("/widget.com/key/osVersionID")
	if err != nil {
		log.Fatalf("Key '/widget.com/key/osVersionID' not created %v", err)
	}

	// Create measures
	mf, err := stats.NewMeasureFloat64("/widget.com/measure/float64/video_size", "size of video document processed in megabytes (10**6)", "MBy")
	if err != nil {
		log.Fatalf("Measure '/widget.com/measure/float64/video_size' not created %v", err)
	}
	mi, err := stats.NewMeasureInt64("/widget.com/measure/int64/video_spam_count", "count of videos marked as spam/inappropriate", "1")
	if err != nil {
		log.Fatalf("Measure '/widget.com/measure/int64/video_spam_count' not created %v", err)
	}

	// Create aggregations
	histogramBounds := []float64{-10, 0, 10, 20}
	agg1 := stats.NewAggregationDistribution(histogramBounds)
	agg2 := stats.NewAggregationCount()

	duration := 10 * time.Second
	precisionIntervals := 10
	wnd1 := stats.NewWindowSlidingTime(duration, precisionIntervals)

	// Create views
	myView1 := stats.NewView("/widget.com/view/video_size/distribution", "a distribution of video sizes processed tagged by device and os", []tags.Key{key1, key2}, mf, agg1, wnd1)
	myView2 := stats.NewView("/widget.com/view/video_spam_count/count", "a count of video marked as spam tagged by device", []tags.Key{key1}, mi, agg2, wnd1)

	// Register views
	if err = stats.RegisterView(myView1); err != nil {
		log.Fatalf("View %v not registered. %v", myView1, err)
	}
	if err = stats.RegisterView(myView2); err != nil {
		log.Fatalf("View %v not registered. %v", myView2, err)
	}

	// set the reporting period to 1 second instead of the 10 seconds default
	reporitngDuration := 1 * time.Second
	stats.SetReportingPeriod(reporitngDuration)

	// Subscribe to view
	c1 := make(chan *stats.ViewData, 4)

	// Process collected data asynchronously
	go func(c chan *stats.ViewData) {
		for vd := range c1 {
			log.Printf("ViewData collected for view %v received after default duration elapsed. %v row(s) received", vd.V.Name(), len(vd.Rows))
			for _, r := range vd.Rows {
				log.Printf("row received with len(tags): %v", len(r.Tags))
			}
		}
	}(c1)

	if err = stats.SubscribeToView(myView1, c1); err != nil {
		log.Fatalf("Subscription to view %v failed. %v", myView1, err)
	}

	// Explicitly instruct the library to collect the view data for on-demand
	// retrieval
	if err := stats.ForceCollection(myView2); err != nil {
		log.Fatalf("Forced collection of view %v failed. %v", myView2, err)
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
	stats.Record(ctx, mi.Is(2), mf.Is(100.0))

	// Wait for a duration longer than reporting duration to ensure the census
	// library reports the collected data
	fmt.Printf("\nWait %v for default reporting duration to kick in\n", reporitngDuration+100*time.Millisecond)
	time.Sleep(reporitngDuration + 100*time.Millisecond)

	fmt.Print("\nRetrieve data on demand\n")
	// Pull collected data synchronously from the library
	rows, err := stats.RetrieveData(myView2)
	if err != nil {
		log.Fatalf("Retrieving data from view %v failed. %v", myView2, err)
	}

	// Process collected data on-demand
	log.Printf("ViewData collected for view %v received on demand. %v row(s) received", myView2.Name(), len(rows))
	for _, r := range rows {
		log.Printf("row received with len(tags): %v", len(r.Tags))
	}

	fmt.Println()
}
