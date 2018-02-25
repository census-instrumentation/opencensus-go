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

// Command xray is an example program that creates spans
// and uploads to AWS X-Ray.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.opencensus.io/exporter/xray"
	"go.opencensus.io/trace"
)

func main() {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		log.Fatalln("AWS_ACCESS_KEY_ID must be set")
	}
	if os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Fatalln("AWS_SECRET_ACCESS_KEY must be set")
	}
	if os.Getenv("AWS_DEFAULT_REGION") == "" {
		log.Fatalln("AWS_DEFAULT_REGION must be set")
	}

	ctx := context.Background()

	// Register the AWS X-Ray exporter to be able to retrieve
	// the collected spans.
	exporter, err := xray.NewExporter(
		xray.WithVersion("latest"),
		xray.WithOnExport(func(in xray.OnExport) {
			fmt.Println("publishing trace,", in.TraceID)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	// For demoing purposes, always sample.
	trace.SetDefaultSampler(trace.AlwaysSample())

	ctx, span := trace.StartSpan(ctx, "/foo")
	bar(ctx)
	span.End()
}

func bar(ctx context.Context) {
	ctx, span := trace.StartSpan(ctx, "/bar")
	defer span.End()

	// Do bar...
}
