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

// Package stats contains an example program that collects data for
// video size and spam video count over a time window. Collected data is
// tagged with operating system and device ID.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/census-instrumentation/opencensus-go/stats"
	"github.com/census-instrumentation/opencensus-go/tag"
)

func main() {
	// Creates keys. There will be two dimensions, device ID and OS version,
	// that will be collected with the metrics we collect in the program.
	deviceIDKey, err := tag.NewStringKey("/mycompany.com/key/deviceID")
	if err != nil {
		log.Fatalf("Device ID key not created: %v\n", err)
	}
	osVersionKey, err := tag.NewStringKey("/mycompany.com/key/osVersionKey")
	if err != nil {
		log.Fatalf("OS version key not created: %v\n", err)
	}

	// Create measures. The program will record measures for the size of
	// processed videos and the nubmer of videos marked as spam.
	videoSize, err := stats.NewMeasureFloat64("/mycompany.com/measure/video_size", "size of processed videos", "MBy")
	if err != nil {
		log.Fatalf("Video size measure not created: %v\n", err)
	}
	videoSpamCount, err := stats.NewMeasureInt64("/mycompany.com/measure/video_spam_count", "number of videos marked as spam", "1")
	if err != nil {
		log.Fatalf("Video spam count measure not created %v\n", err)
	}

	// Create aggregations.
	agg1 := stats.DistributionAggregation([]float64{-10, 0, 10, 20})
	agg2 := stats.CountAggregation{}

	window := stats.SlidingTimeWindow{Duration: 10 * time.Second, Intervals: 10}

	// Create views.
	const (
		videoSizeName = "/mycompany.com/view/video_size/distribution"
		videoSizeDesc = "a distribution of video sizes processed tagged by device ID and OS"
		videoSpamName = "/mycompany.com/view/video_spam_count/count"
		videoSpamDesc = "count of videos marked as spam tagged by device ID"
	)
	// Create view to see video size over 10 seconds with device ID and OS version tags.
	videoSizeView := stats.NewView(videoSizeName, videoSizeDesc, []tag.Key{deviceIDKey, osVersionKey}, videoSize, agg1, window)
	// Create view to see the count of spam videos over 10 sconds with device ID.
	videoSpamCountView := stats.NewView(videoSpamName, videoSpamDesc, []tag.Key{deviceIDKey}, videoSpamCount, agg2, window)

	// Register views in order to collect data.
	if err := stats.RegisterView(videoSizeView); err != nil {
		log.Fatalf("View %v cannot be registered: %v\n", videoSizeView, err)
	}
	if err = stats.RegisterView(videoSpamCountView); err != nil {
		log.Fatalf("View %v cannot be registered: %v\n", videoSpamCountView, err)
	}

	// Set reporting period to report data at every second.
	stats.SetReportingPeriod(1 * time.Second)

	// ForceCollect explicitly instructs the library to collect the
	// view data for on-demand retrieval.
	if err := videoSpamCountView.ForceCollect(); err != nil {
		log.Fatalf("Cannot force collect from the video spam count view: %v\n", err)
	}

	// Subscribe will allow view data to be exported.
	// Once no longer need, you can unsubscribe from the view.
	if err := videoSizeView.Subscribe(); err != nil {
		log.Fatalf("Cannot subscribe to the view: %v\n", err)
	}

	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})

	// Record usage. This section demonstrates how user code
	// can record measures for video size and spam count with
	// device ID and OS version dimensions.

	// Adding tags to context to record each datapoint with
	// the following device ID and OS version.
	tm := tag.NewMap(nil,
		tag.UpsertString(deviceIDKey, "device-id-768dfd76"),
		tag.UpsertString(osVersionKey, "mac-osx-10.12.6"),
	)
	ctx := tag.NewContext(context.Background(), tm)

	// Recording datapoints.
	stats.Record(ctx, videoSpamCount.M(2), videoSize.M(100.0))

	// Wait for a duration longer than reporting duration to ensure the stats
	// library reports the collected data.
	fmt.Println("Wait longer than the reporting duration...")
	time.Sleep(2 * time.Second)

	fmt.Println("Retrieving data on demand...")
	rows, err := videoSpamCountView.RetrieveData()
	if err != nil {
		log.Fatalf("Cannot retrieve spam stats data: %v", err)
	}

	// Process the collected data.
	log.Printf("Data collected from view %v; %v row(s) received\n", videoSpamCountView.Name(), len(rows))
	for _, r := range rows {
		log.Println(r)
	}
}

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}
