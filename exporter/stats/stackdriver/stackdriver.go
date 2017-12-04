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
	"errors"
	"fmt"
	"log"
	"net/url"
	"path"
	"reflect"
	"sync"
	"time"

	"go.opencensus.io/tag"

	"go.opencensus.io/internal"

	"go.opencensus.io/stats"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/option"
	"google.golang.org/api/support/bundler"
	distributionpb "google.golang.org/genproto/googleapis/api/distribution"
	labelpb "google.golang.org/genproto/googleapis/api/label"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const maxTimeSeriesPerUpload = 200

// Exporter exports stats to the Stackdriver Monitoring.
type Exporter struct {
	bundler *bundler.Bundler
	o       Options

	createdViewsMu sync.Mutex
	createdViews   map[string]*metricpb.MetricDescriptor // Views already created remotely

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
		c:            client,
		o:            o,
		createdViews: make(map[string]*metricpb.MetricDescriptor),
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
	for _, req := range e.makeReq(vds, maxTimeSeriesPerUpload) {
		if err := e.c.CreateTimeSeries(ctx, req); err != nil {
			// TODO(jbd): Don't fail fast here, batch errors?
			return err
		}
	}
	return nil
}

func (e *Exporter) makeReq(vds []*stats.ViewData, limit int) []*monitoringpb.CreateTimeSeriesRequest {
	var reqs []*monitoringpb.CreateTimeSeriesRequest
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
					Type:   namespacedViewName(vd.View.Name(), false),
					Labels: newLabels(row.Tags),
				},
				Resource: &monitoredrespb.MonitoredResource{
					Type: "global",
				},
				Points: []*monitoringpb.Point{newPoint(vd.View, row, vd.Start, vd.End)},
			}
			timeSeries = append(timeSeries, ts)
			if len(timeSeries) == limit {
				reqs = append(reqs, &monitoringpb.CreateTimeSeriesRequest{
					Name:       monitoring.MetricProjectPath(e.o.ProjectID),
					TimeSeries: timeSeries,
				})
				timeSeries = []*monitoringpb.TimeSeries{}
			}
		}
	}
	if len(timeSeries) > 0 {
		reqs = append(reqs, &monitoringpb.CreateTimeSeriesRequest{
			Name:       monitoring.MetricProjectPath(e.o.ProjectID),
			TimeSeries: timeSeries,
		})
	}
	return reqs
}

// createMeasure creates a MetricDescriptor for the given view data in Stackdriver Monitoring.
// An error will be returned if there is already a metric descriptor created with the same name
// but it has a different aggregation, window or keys.
func (e *Exporter) createMeasure(ctx context.Context, vd *stats.ViewData) error {
	e.createdViewsMu.Lock()
	defer e.createdViewsMu.Unlock()

	m := vd.View.Measure()
	agg := vd.View.Aggregation()
	window := vd.View.Window()
	tagKeys := vd.View.TagKeys()
	viewName := vd.View.Name()

	if md, ok := e.createdViews[viewName]; ok {
		// Check agg, window and keys.
		if err := equalAggWindowTagKeys(md, agg, window, tagKeys); err != nil {
			return err
		}
		return nil
	}

	metricName := monitoring.MetricMetricDescriptorPath(e.o.ProjectID, namespacedViewName(viewName, true))
	md, err := e.c.GetMetricDescriptor(ctx, &monitoringpb.GetMetricDescriptorRequest{
		Name: metricName,
	})
	if err == nil {
		if err := equalAggWindowTagKeys(md, agg, window, tagKeys); err != nil {
			return err
		}
		e.createdViews[viewName] = md
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
	case stats.SumAggregation:
		valueType = metricpb.MetricDescriptor_DOUBLE
	case stats.MeanAggregation:
		valueType = metricpb.MetricDescriptor_DOUBLE
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

	md, err = e.c.CreateMetricDescriptor(ctx, &monitoringpb.CreateMetricDescriptorRequest{
		Name: monitoring.MetricProjectPath(e.o.ProjectID),
		MetricDescriptor: &metricpb.MetricDescriptor{
			DisplayName: viewName,
			Description: m.Description(),
			Unit:        m.Unit(),
			Type:        namespacedViewName(viewName, false),
			MetricKind:  metricKind,
			ValueType:   valueType,
			Labels:      newLabelDescriptors(vd.View.TagKeys()),
		},
	})
	if err != nil {
		return err
	}

	e.createdViews[viewName] = md
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
	switch v := r.Data.(type) {
	case *stats.CountData:
		return &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_Int64Value{
			Int64Value: int64(*v),
		}}
	case *stats.SumData:
		return &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DoubleValue{
			DoubleValue: float64(*v),
		}}
	case *stats.MeanData:
		return &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DoubleValue{
			DoubleValue: v.Mean,
		}}
	case *stats.DistributionData:
		bounds := view.Aggregation().(stats.DistributionAggregation)
		return &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_DistributionValue{
			DistributionValue: &distributionpb.Distribution{
				Count: v.Count,
				Mean:  v.Mean,
				SumOfSquaredDeviation: v.SumOfSquaredDev,
				// TODO(songya): uncomment this once Stackdriver supports min/max.
				// Range: &distributionpb.Distribution_Range{
				// 	Min: v.Min,
				// 	Max: v.Max,
				// },
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

func namespacedViewName(v string, escaped bool) string {
	p := path.Join("opencensus", v)
	if escaped {
		p = url.PathEscape(p)
	}
	return path.Join("custom.googleapis.com", p)
}

func newLabels(tags []tag.Tag) map[string]string {
	labels := make(map[string]string)
	for _, tag := range tags {
		labels[internal.Sanitize(tag.Key.Name())] = tag.Value
	}
	return labels
}

func newLabelDescriptors(keys []tag.Key) []*labelpb.LabelDescriptor {
	labelDescriptors := make([]*labelpb.LabelDescriptor, len(keys))
	for i, key := range keys {
		labelDescriptors[i] = &labelpb.LabelDescriptor{
			Key:       internal.Sanitize(key.Name()),
			ValueType: labelpb.LabelDescriptor_STRING, // We only use string tags
		}
	}
	return labelDescriptors
}

func equalAggWindowTagKeys(md *metricpb.MetricDescriptor, agg stats.Aggregation, window stats.Window, keys []tag.Key) error {
	var w stats.Window
	var a stats.Aggregation

	switch md.MetricKind {
	case metricpb.MetricDescriptor_DELTA:
		w = stats.Interval{}
	case metricpb.MetricDescriptor_CUMULATIVE:
		w = stats.Cumulative{}
	}

	switch md.ValueType {
	case metricpb.MetricDescriptor_INT64:
		a = stats.CountAggregation{}
	case metricpb.MetricDescriptor_DISTRIBUTION:
		a = stats.DistributionAggregation{}
	}

	aggType := reflect.TypeOf(agg)
	if aggType.Kind() == reflect.Ptr { // if pointer, find out the concrete type
		aggType = reflect.ValueOf(agg).Elem().Type()
	}
	if aggType != reflect.TypeOf(a) {
		return fmt.Errorf("stackdriver metric descriptor was not created with aggregation type %T", a)
	}

	winType := reflect.TypeOf(window)
	if winType.Kind() == reflect.Ptr { // if pointer, find out the concrete type
		winType = reflect.ValueOf(window).Elem().Type()
	}
	if winType != reflect.TypeOf(w) {
		return fmt.Errorf("stackdriver metric descriptor was not created with window type %T", w)
	}

	if len(md.Labels) != len(keys) {
		return errors.New("stackdriver metric descriptor was not created with the view labels")
	}

	labels := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		labels[internal.Sanitize(k.Name())] = struct{}{}
	}

	for _, k := range md.Labels {
		if _, ok := labels[k.Key]; !ok {
			return fmt.Errorf("stackdriver metric descriptor was not created with label %q", k)
		}
	}

	return nil
}
