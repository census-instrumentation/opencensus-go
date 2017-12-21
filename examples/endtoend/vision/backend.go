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

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"golang.org/x/net/context"

	"cloud.google.com/go/vision/apiv1"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"

	"go.opencensus.io/exporter/stats/prometheus"
	"go.opencensus.io/exporter/stats/stackdriver"
	"go.opencensus.io/stats"
	"go.opencensus.io/trace"
)

var (
	byteCountMeasure  stats.Measure
	byteBucketMeasure stats.Measure
	errCountMeasure   stats.Measure
	imageCountMeasure stats.Measure
	urlCountMeasure   stats.Measure
)

func init() {
	port := flag.Int("port", 8899, "the port to run the server on")
	projectID := flag.String("gcp-id", "opencensus-demos", "the Google Cloud Platform projectID")
	flag.Parse()

	addr = fmt.Sprintf(":%d", *port)

	promExp, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatalf("prometheus exporter: %v", err)
	}
	stats.RegisterExporter(promExp)
	mux.Handle("/metrics", promExp)

	stackDExp, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: *projectID,
	})
	if err != nil {
		log.Fatalf("stackDriver exporter: %v", err)
	}
	stats.RegisterExporter(stackDExp)

	errCountMeasure = mustCreateCountMeasure(func() (stats.Measure, error) {
		return stats.NewMeasureInt64("vision/measures/errors_cum", "number of errors", "error")
	},
		"errors_cum",
		"number of errors over time",
	)

	imageCountMeasure = mustCreateCountMeasure(func() (stats.Measure, error) {
		return stats.NewMeasureInt64("vision/measures/images_cum", "number of images uploaded", "image")
	},
		"images_cum",
		"number of images uploaded over time",
	)

	byteCountMeasure = mustCreateDistMeasure(func() (stats.Measure, error) {
		return stats.NewMeasureInt64("vision/measures/bytes_in_cum", "the number of bytes processed", "byte")
	},
		"bytes_bucket_cum",
		"number of bytes ingested over time",
		// Place them in buckets: [0, 1KiB, 100KiB, 1MiB, 10MiB, 100MiB, 1GiB]
		stats.DistributionAggregation([]float64{0, 1 << 10, 100 << 10, 1 << 20, 10 << 20, 100 << 20, 10 << 30}),
	)

	// Now set the reporting period
	stats.SetReportingPeriod(15 * time.Second)
}

func mustCreateCountMeasure(measureFn func() (stats.Measure, error), name, desc string) stats.Measure {
	return mustCreateMeasure(measureFn, name, desc, stats.CountAggregation{})
}

func mustCreateDistMeasure(measureFn func() (stats.Measure, error), name, desc string, agg stats.Aggregation) stats.Measure {
	return mustCreateMeasure(measureFn, name, desc, agg)
}

func mustCreateMeasure(measureFn func() (stats.Measure, error), name, desc string, agg stats.Aggregation) stats.Measure {
	m, err := measureFn()
	if err != nil {
		log.Fatalf("creating measure: %v", err)
	}
	v, err := stats.NewView(
		name,
		desc,
		nil,
		m,
		agg,
		stats.Cumulative{},
	)
	if err != nil {
		log.Fatalf("creating view: %v", err)
	}
	if err := v.Subscribe(); err != nil {
		log.Fatalf("view subscription: %v", err)
	}
	// No need to unsubscribe the view since it last for the lifetime of the webapp
	return m
}

type countWriter struct {
	n int
}

func (cr *countWriter) Write(b []byte) (int, error) {
	cr.n += len(b)
	return len(b), nil
}

var _ io.Writer = (*countWriter)(nil)

func recordStatsErrorCount(ctx context.Context, n int64) {
	stats.Record(ctx, errCountMeasure.(*stats.MeasureInt64).M(n))
}

func detectFacesAndLogos(r io.Reader, ctx context.Context) (*DetectionResult, error) {
	ctx = trace.StartSpan(ctx, "/detect")
	defer trace.EndSpan(ctx)

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, err
	}

	cw := new(countWriter)
	mr := io.TeeReader(r, cw)
	img, err := vision.NewImageFromReader(mr)
	if err != nil {
		recordStatsErrorCount(ctx, 1)
		return nil, err
	}
	stats.Record(ctx, byteCountMeasure.(*stats.MeasureInt64).M(int64(cw.n)))
	go stats.Record(ctx, imageCountMeasure.(*stats.MeasureInt64).M(int64(1)))

	res := new(DetectionResult)
	res.Faces, res.FacesErr = client.DetectFaces(ctx, img, nil, 1000)
	res.Labels, res.LabelsErr = client.DetectLabels(ctx, img, nil, 1000)
	return res, nil
}

type DetectionResult struct {
	Faces     []*pb.FaceAnnotation   `json:"faces,omitempty"`
	FacesErr  error                  `json:"faces_err,omitempty"`
	Labels    []*pb.EntityAnnotation `json:"labels,omitempty"`
	LabelsErr error                  `json:"labels_err,omitempty"`
}
