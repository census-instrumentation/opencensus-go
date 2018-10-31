// Copyright 2018, OpenCensus Authors
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

package main

import (
	"go.opencensus.io/metric"
	"log"
	"net/http"
	"time"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"

	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/stats/view"
)

const server = "http://localhost:50030"

func main() {
	// Start a metrics exporter to log metrics to os.Stderr.
	logger := metric.NewLogExporter()
	logger.ReportingPeriod = 2 * time.Second
	go logger.Run()

	// Register trace exporter to export the collected data.
	trace.RegisterExporter(&exporter.PrintExporter{})

	// Always trace for this demo. In a production application, you should
	// configure this to a trace.ProbabilitySampler set at the desired
	// probability.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// Report stats at every second.
	view.SetReportingPeriod(1 * time.Second)

	client := &http.Client{Transport: &ochttp.Transport{}}

	resp, err := client.Get(server)
	if err != nil {
		log.Printf("Failed to get response: %v", err)
	} else {
		resp.Body.Close()
	}

	time.Sleep(2 * time.Second) // Wait until stats are reported.
}
