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
	"testing"
	"time"

	"go.opencensus.io/stats"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

func TestExporter_makeReq(t *testing.T) {
	m, _ := stats.NewMeasureFloat64("test-measure", "measure desc", "unit")
	defer stats.DeleteMeasure(m)

	cumView := stats.NewView("cumview", "desc", nil, m, stats.CountAggregation{}, stats.Cumulative{})
	if err := stats.RegisterView(cumView); err != nil {
		t.Fatal(err)
	}

	distView := stats.NewView("distview", "desc", nil, m, stats.DistributionAggregation([]float64{2, 4, 7}), stats.Interval{})
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
							Type: "custom.googleapis.com/opencensus/cumview",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type:   "global",
							Labels: map[string]string{"project_id": "proj-id"},
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
							Type: "custom.googleapis.com/opencensus/cumview",
						},
						Resource: &monitoredrespb.MonitoredResource{
							Type:   "global",
							Labels: map[string]string{"project_id": "proj-id"},
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

func newTestCumViewData(v *stats.View, start, end time.Time) *stats.ViewData {
	count1 := stats.CountData(10)
	count2 := stats.CountData(16)
	return &stats.ViewData{
		View: v,
		Rows: []*stats.Row{
			{Data: &count1},
			{Data: &count2},
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
