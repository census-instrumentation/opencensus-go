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

// Package stackdriver contains the OpenCensus exporters for
// Stackdriver Monitoring.
//
// Please note that the Stackdriver exporter is currently experimental.
package stackdriver

import (
	"context"
	"fmt"
	"log"
	"path"
	"sync"
	"time"

	"go.opencensus.io/internal"

	"go.opencensus.io/stats"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/option"
	"google.golang.org/api/support/bundler"
	distributionpb "google.golang.org/genproto/googleapis/api/distribution"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Exporter exports stats to the Stackdriver Monitoring.
type Exporter struct {
	bundler *bundler.Bundler
	o       Options

	measuresMu sync.Mutex
	measures   map[stats.Measure]struct{} // measures already created remotely

	c *monitoring.MetricClient
}

// Options contains options for configuring the exporter.
type Options struct {
	// ProjectID is the identifier of the Stackdriver
	// project the user is uploading the stats data to.
	ProjectID string

	// OnError is the hook to be called when there is
	// an error occured when uploading the stats data.
	// If no custom hook is set, errors are logged.
	// Optional.
	OnError func(err error)

	// ClientOptions are additional options to be passed
	// to the underlying Stackdriver Monitoring API client.
	// Optional.
	ClientOptions []option.ClientOption

	// BundleDelayThreshold determines the max amount of time
	// the exporter can wait before uploading view data to
	// the backend.
	// Optional.
	BundleDelayThreshold time.Duration

	// BundleCountThreshold determines how many view data events
	// can be buffered before batch uploading them to the backend.
	// Optional.
	BundleCountThreshold int
}

// NewExporter returns an exporter that uploads stats data to Stackdriver Monitoring.
func NewExporter(o Options) (*Exporter, error) {
	opts := append(o.ClientOptions, option.WithUserAgent(internal.UserAgent))
	client, err := monitoring.NewMetricClient(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	e := &Exporter{
		c:        client,
		o:        o,
		measures: make(map[stats.Measure]struct{}),
	}
	e.bundler = bundler.NewBundler((*stats.ViewData)(nil), func(bundle interface{}) {
		vds := bundle.([]*stats.ViewData)
		if err := e.upload(vds); err != nil {
			e.onError(err)
		}
	})
	e.bundler.DelayThreshold = e.o.BundleDelayThreshold
	e.bundler.BundleCountThreshold = e.o.BundleCountThreshold
	return e, nil
}

// Export exports to the Stackdriver Monitoring if view data
// has one or more rows.
func (e *Exporter) Export(vd *stats.ViewData) {
	if len(vd.Rows) == 0 {
		return
	}
	e.bundler.Add(vd, 1)
}

// Flush waits for exported view data to be uploaded.
//
// This is useful if your program is ending and you do not
// want to lose recent spans.
func (e *Exporter) Flush() {
	e.bundler.Flush()
}

func (e *Exporter) onError(err error) {
	if e.o.OnError != nil {
		e.o.OnError(err)
		return
	}
	log.Printf("Failed to export to Stackdriver Monitoring: %v", err)
}

func (e *Exporter) upload(vds []*stats.ViewData) error {
	ctx := context.Background()

	for _, vd := range vds {
		if _, ok := vd.View.Window().(stats.Cumulative); !ok {
			// TODO(jbd): Only Cumulative window will be exported to Stackdriver in this version.
			// Support others when custom delta metrics are supported.
			continue
		}
		if err := e.createMeasure(ctx, vd); err != nil {
			return err
		}
	}
	req := e.makeReq(vds)
	if req == nil {
		return nil
	}
	if err := e.c.CreateTimeSeries(ctx, req); err != nil {
		return err
	}
	return nil
}

func (e *Exporter) makeReq(vds []*stats.ViewData) *monitoringpb.CreateTimeSeriesRequest {
	var timeSeries []*monitoringpb.TimeSeries
	for _, vd := range vds {
		if _, ok := vd.View.Window().(stats.Cumulative); !ok {
			// TODO(jbd): Only Cumulative window will be exported to Stackdriver in this version.
			// Support others when custom delta metrics are supported.
			continue
		}
		for _, row := range vd.Rows {
			ts := &monitoringpb.TimeSeries{
				Metric: &metricpb.Metric{
					Type: path.Join("custom.googleapis.com", vd.View.Measure().Name()),
					// TODO(jbd): Add labels.
				},
				Resource: &monitoredrespb.MonitoredResource{
					Type:   "global",
					Labels: map[string]string{"project_id": e.o.ProjectID},
				},
				Points: []*monitoringpb.Point{newPoint(vd.View, row, vd.Start, vd.End)},
			}
			timeSeries = append(timeSeries, ts)
		}
	}
	if len(timeSeries) == 0 {
		return nil
	}
	return &monitoringpb.CreateTimeSeriesRequest{
		Name:       monitoring.MetricProjectPath(e.o.ProjectID),
		TimeSeries: timeSeries,
	}
}

func (e *Exporter) createMeasure(ctx context.Context, vd *stats.ViewData) error {
	e.measuresMu.Lock()
	defer e.measuresMu.Unlock()

	m := vd.View.Measure()
	agg := vd.View.Aggregation()
	window := vd.View.Window()

	_, ok := e.measures[m]
	if ok {
		return nil
	}

	name := monitoring.MetricMetricDescriptorPath(e.o.ProjectID, m.Name())
	_, err := e.c.GetMetricDescriptor(ctx, &monitoringpb.GetMetricDescriptorRequest{
		Name: name,
	})
	if err == nil {
		e.measures[m] = struct{}{}
		return nil
	}
	if grpc.Code(err) != codes.NotFound {
		return err
	}

	var metricKind metricpb.MetricDescriptor_MetricKind
	var valueType metricpb.MetricDescriptor_ValueType

	switch agg.(type) {
	case stats.CountAggregation:
		valueType = metricpb.MetricDescriptor_INT64
	case stats.DistributionAggregation:
		valueType = metricpb.MetricDescriptor_DISTRIBUTION
	default:
		return fmt.Errorf("unsupported aggregation type: %T", agg)
	}

	switch window.(type) {
	case stats.Cumulative:
		metricKind = metricpb.MetricDescriptor_CUMULATIVE
	case stats.Interval:
		metricKind = metricpb.MetricDescriptor_DELTA
	default:
		return fmt.Errorf("unsupported window type: %T", window)
	}

	if _, err := e.c.CreateMetricDescriptor(ctx, &monitoringpb.CreateMetricDescriptorRequest{
		Name: monitoring.MetricProjectPath(e.o.ProjectID),
		MetricDescriptor: &metricpb.MetricDescriptor{
			DisplayName: vd.View.Name(),
			Description: m.Description(),
			Unit:        m.Unit(),
			Type:        path.Join("custom.googleapis.com", m.Name()),
			MetricKind:  metricKind,
			ValueType:   valueType,
			Labels:      nil, // TODO(jbd): Add labels.
		},
	}); err != nil {
		return err
	}

	e.measures[m] = struct{}{}
	return nil
}

func newPoint(v *stats.View, row *stats.Row, start, end time.Time) *monitoringpb.Point {
	return &monitoringpb.Point{
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{
				Seconds: start.Unix(),
				Nanos:   int32(start.Nanosecond()),
			},
			EndTime: &timestamp.Timestamp{
				Seconds: end.Unix(),
				Nanos:   int32(end.Nanosecond()),
			},
		},
		Value: newTypedValue(v, row),
	}
}

func newTypedValue(view *stats.View, r *stats.Row) *monitoringpb.TypedValue {
	switch r.Data.(type) {
	case *stats.CountData:
		v := r.Data.(*stats.CountData)
		return &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_Int64Value{
			Int64Value: int64(*v),
		}}
	case *stats.DistributionData:
		v := r.Data.(*stats.DistributionData)
		bounds := view.Aggregation().(stats.DistributionAggregation)
		return &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DistributionValue{
			DistributionValue: &distributionpb.Distribution{
				Count: v.Count,
				Mean:  v.Mean,
				SumOfSquaredDeviation: v.SumOfSquaredDev,
				Range: &distributionpb.Distribution_Range{
					Min: v.Min,
					Max: v.Max,
				},
				BucketOptions: &distributionpb.Distribution_BucketOptions{
					Options: &distributionpb.Distribution_BucketOptions_ExplicitBuckets{
						ExplicitBuckets: &distributionpb.Distribution_BucketOptions_Explicit{
							Bounds: []float64(bounds),
						},
					},
				},
				BucketCounts: v.CountPerBucket,
			},
		}}
	}
	return nil
}
