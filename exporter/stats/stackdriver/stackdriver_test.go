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

package stackdriver

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/api/label"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func TestExporter_makeReq(t *testing.T) {
	m, err := stats.NewMeasureFloat64("test-measure", "measure desc", "unit")
	if err != nil {
		t.Fatal(err)
	}
	defer stats.DeleteMeasure(m)

	key, err := tag.NewKey("test_key")
	if err != nil {
		t.Fatal(err)
	}

	cumView, err := stats.NewView("cumview", "desc", []tag.Key{key}, m, stats.CountAggregation{}, stats.Cumulative{})
	if err != nil {
		t.Fatal(err)
	}
	if err := stats.RegisterView(cumView); err != nil {
		t.Fatal(err)
	}

	distView, err := stats.NewView("distview", "desc", nil, m, stats.DistributionAggregation([]float64{2, 4, 7}), stats.Interval{})
	if err != nil {
		t.Fatal(err)
	}
	if err := stats.RegisterView(distView); err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	end := start.Add(time.Minute)

	tests := []struct {
		name   string
		projID string
		vd     *stats.ViewData
		want   *monitoringpb.CreateTimeSeriesRequest
	}{
		{
			name:   "count agg + cum timeline",
			projID: "proj-id",
			vd:     newTestCumViewData(cumView, start, end),
			want: &monitoringpb.CreateTimeSeriesRequest{
				Name: monitoring.MetricProjectPath("proj-id"),
				TimeSeries: []*monitoringpb.TimeSeries{
					{
						Metric: &metricpb.Metric{
							Type:   "custom.googleapis.com/opencensus/cumview",
							Labels: map[string]string{"test_key": "test-value-1"},
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "global",
						},
						Points: []*monitoringpb.Point{
							{
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
								Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_Int64Value{
									Int64Value: 10,
								}},
							},
						},
					},
					{
						Metric: &metricpb.Metric{
							Type:   "custom.googleapis.com/opencensus/cumview",
							Labels: map[string]string{"test_key": "test-value-2"},
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type: "global",
						},
						Points: []*monitoringpb.Point{
							{
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
								Value: &monitoringpb.TypedValue{Value: &monitoringpb.TypedValue_Int64Value{
									Int64Value: 16,
								}},
							},
						},
					},
				},
			},
		},
		{
			name:   "dist agg + time window",
			projID: "proj-id",
			vd:     newTestDistViewData(distView, start, end),
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Exporter{o: Options{ProjectID: tt.projID}}
			if got := e.makeReq([]*stats.ViewData{tt.vd}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%v: Exporter.makeReq() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestEqualAggWindowTagKeys(t *testing.T) {
	key1, _ := tag.NewKey("test_key_one")
	key2, _ := tag.NewKey("test_key_two")
	tests := []struct {
		name    string
		md      *metricpb.MetricDescriptor
		agg     stats.Aggregation
		keys    []tag.Key
		window  stats.Window
		wantErr bool
	}{
		{
			name: "count agg + cum",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_INT64,
			},
			agg:     stats.CountAggregation{},
			window:  stats.Cumulative{},
			wantErr: false,
		},
		{
			name: "distribution agg + cum - mismatch",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
			},
			agg:     stats.CountAggregation{},
			window:  stats.Cumulative{},
			wantErr: true,
		},
		{
			name: "distribution agg + delta",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_DELTA,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
			},
			agg:     stats.DistributionAggregation{},
			window:  stats.Interval{},
			wantErr: false,
		},
		{
			name: "distribution agg + cum",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
			},
			agg:     stats.DistributionAggregation{},
			window:  stats.Interval{},
			wantErr: true,
		},
		{
			name: "distribution agg + cum with keys",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
				Labels: []*label.LabelDescriptor{
					{Key: "test_key_one"},
					{Key: "test_key_two"},
				},
			},
			agg:     stats.DistributionAggregation{},
			window:  stats.Cumulative{},
			keys:    []tag.Key{key1, key2},
			wantErr: false,
		},
		{
			name: "distribution agg + cum with keys -- mismatch",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_DISTRIBUTION,
			},
			agg:     stats.DistributionAggregation{},
			window:  stats.Cumulative{},
			keys:    []tag.Key{key1, key2},
			wantErr: true,
		},
		{
			name: "count agg + cum with pointers",
			md: &metricpb.MetricDescriptor{
				MetricKind: metricpb.MetricDescriptor_CUMULATIVE,
				ValueType:  metricpb.MetricDescriptor_INT64,
			},
			agg:     &stats.CountAggregation{},
			window:  &stats.Cumulative{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := equalAggWindowTagKeys(tt.md, tt.agg, tt.window, tt.keys)
			if err != nil && !tt.wantErr {
				t.Errorf("equalAggWindowTagKeys() = %q; want no error", err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("equalAggWindowTagKeys() = %q; want error", err)
			}

		})
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trunacate long string",
			input:    strings.Repeat("a", 101),
			expected: strings.Repeat("a", 100),
		},
		{
			name:     "replace character",
			input:    "test/key-1",
			expected: "test_key_1",
		},
		{
			name:     "don't modify alphanumeric string",
			input:    "testkey1",
			expected: "testkey1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := sanitize(tt.input)
			if actual != tt.expected {
				t.Errorf("sanitize() = %s; want %s", actual, tt.expected)
			}
		})
	}
}

func newTestCumViewData(v *stats.View, start, end time.Time) *stats.ViewData {
	count1 := stats.CountData(10)
	count2 := stats.CountData(16)
	key, _ := tag.NewKey("test-key")
	tag1 := tag.Tag{Key: key, Value: "test-value-1"}
	tag2 := tag.Tag{Key: key, Value: "test-value-2"}
	return &stats.ViewData{
		View: v,
		Rows: []*stats.Row{
			{
				Tags: []tag.Tag{tag1},
				Data: &count1,
			},
			{
				Tags: []tag.Tag{tag2},
				Data: &count2,
			},
		},
		Start: start,
		End:   end,
	}
}

func newTestDistViewData(v *stats.View, start, end time.Time) *stats.ViewData {
	return &stats.ViewData{
		View: v,
		Rows: []*stats.Row{
			{Data: &stats.DistributionData{
				Count:           5,
				Min:             1,
				Max:             7,
				Mean:            3,
				SumOfSquaredDev: 1.5,
				CountPerBucket:  []int64{2, 2, 1},
			}},
		},
		Start: start,
		End:   end,
	}
}
