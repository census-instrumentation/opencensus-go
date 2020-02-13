// Copyright 2019, OpenCensus Authors
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

package stats_test

import (
	"context"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

var (
	tid     = trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 4, 8, 16, 32, 64, 128}
	sid     = trace.SpanID{1, 2, 4, 8, 16, 32, 64, 128}
	spanCtx = trace.SpanContext{
		TraceID:      tid,
		SpanID:       sid,
		TraceOptions: 1,
	}
)

func TestRecordWithAttachments(t *testing.T) {
	k1 := tag.MustNewKey("k1")
	k2 := tag.MustNewKey("k2")
	distribution := view.Distribution(5, 10)
	m := stats.Int64("TestRecordWithAttachments/m1", "", stats.UnitDimensionless)
	v := &view.View{
		Name:        "test_view",
		TagKeys:     []tag.Key{k1, k2},
		Measure:     m,
		Aggregation: distribution,
	}
	view.SetReportingPeriod(100 * time.Millisecond)
	if err := view.Register(v); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}
	defer view.Unregister(v)

	attachments := map[string]interface{}{metricdata.AttachmentKeySpanContext: spanCtx}
	stats.RecordWithOptions(context.Background(), stats.WithAttachments(attachments), stats.WithMeasurements(m.M(12)))
	rows, err := view.RetrieveData("test_view")
	if err != nil {
		t.Errorf("Failed to retrieve data %v", err)
	}
	if len(rows) == 0 {
		t.Errorf("No data was recorded.")
	}
	data := rows[0].Data
	dis, ok := data.(*view.DistributionData)
	if !ok {
		t.Errorf("want DistributionData, got %+v", data)
	}
	wantBuckets := []int64{0, 0, 1}
	if !reflect.DeepEqual(dis.CountPerBucket, wantBuckets) {
		t.Errorf("want buckets %v, got %v", wantBuckets, dis.CountPerBucket)
	}
	for i, e := range dis.ExemplarsPerBucket {
		// Exemplar slice should be [nil, nil, exemplar]
		if i != 2 && e != nil {
			t.Errorf("want nil exemplar, got %v", e)
		}
		if i == 2 {
			wantExemplar := &metricdata.Exemplar{Value: 12, Attachments: attachments}
			if diff := cmpExemplar(e, wantExemplar); diff != "" {
				t.Fatalf("Unexpected Exemplar -got +want: %s", diff)
			}
		}
	}
}

// Compare exemplars while ignoring exemplar timestamp, since timestamp is non-deterministic.
func cmpExemplar(got, want *metricdata.Exemplar) string {
	return cmp.Diff(got, want, cmpopts.IgnoreFields(metricdata.Exemplar{}, "Timestamp"), cmpopts.IgnoreUnexported(metricdata.Exemplar{}))
}

func TestRecordWithMeter(t *testing.T) {
	meter := view.NewMeter()
	meter.Start()
	defer meter.Stop()
	k1 := tag.MustNewKey("k1")
	k2 := tag.MustNewKey("k2")
	m1 := stats.Int64("TestResolveOptions/m1", "", stats.UnitDimensionless)
	m2 := stats.Int64("TestResolveOptions/m2", "", stats.UnitDimensionless)
	v := []*view.View{{
		Name:        "test_view",
		TagKeys:     []tag.Key{k1, k2},
		Measure:     m1,
		Aggregation: view.Distribution(5, 10),
	}, {
		Name:        "second_view",
		TagKeys:     []tag.Key{k1},
		Measure:     m2,
		Aggregation: view.Count(),
	}}
	meter.SetReportingPeriod(100 * time.Millisecond)
	if err := meter.Register(v...); err != nil {
		t.Fatalf("Failed to register view: %v", err)
	}
	defer meter.Unregister(v...)

	attachments := map[string]interface{}{metricdata.AttachmentKeySpanContext: spanCtx}
	ctx, err := tag.New(context.Background(), tag.Insert(k1, "foo"), tag.Insert(k2, "foo"))
	if err != nil {
		t.Fatalf("Failed to set context: %v", err)
	}
	err = stats.RecordWithOptions(ctx,
		stats.WithTags(tag.Upsert(k1, "bar"), tag.Insert(k2, "bar")),
		stats.WithAttachments(attachments),
		stats.WithMeasurements(m1.M(12), m1.M(6), m2.M(5)),
		stats.WithRecorder(meter))
	if err != nil {
		t.Fatalf("Failed to resolve data point: %v", err)
	}

	rows, err := meter.RetrieveData("test_view")
	if err != nil {
		t.Fatalf("Unable to retrieve data for test_view: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("Expected one row, got %d rows: %+v", len(rows), rows)
	}
	if len(rows[0].Tags) != 2 {
		t.Errorf("Wrong number of tags %d: %v", len(rows[0].Tags), rows[0].Tags)
	}
	// k2 was Insert() ed, and shouldn't update the value that was in the supplied context.
	wantTags := []tag.Tag{{Key: k1, Value: "bar"}, {Key: k2, Value: "foo"}}
	for i, tag := range rows[0].Tags {
		if tag.Key != wantTags[i].Key {
			t.Errorf("Incorrect tag %d, want: %q, got: %q", i, wantTags[i].Key, tag.Key)
		}
		if tag.Value != wantTags[i].Value {
			t.Errorf("Incorrect tag for %s, want: %q, got: %v", tag.Key, wantTags[i].Value, tag.Value)
		}

	}
	wantBuckets := []int64{0, 1, 1}
	gotBuckets := rows[0].Data.(*view.DistributionData)
	if !reflect.DeepEqual(gotBuckets.CountPerBucket, wantBuckets) {
		t.Fatalf("want buckets %v, got %v", wantBuckets, gotBuckets)
	}
	for i, e := range gotBuckets.ExemplarsPerBucket {
		if gotBuckets.CountPerBucket[i] == 0 {
			if e != nil {
				t.Errorf("Unexpected exemplar for bucket")
			}
			continue
		}
		// values from the metrics above
		exemplarValues := []float64{0, 6, 12}
		wantExemplar := &metricdata.Exemplar{Value: exemplarValues[i], Attachments: attachments}
		if diff := cmpExemplar(e, wantExemplar); diff != "" {
			t.Errorf("Bad exemplar for %d: %+v", i, diff)
		}
	}

	rows2, err := meter.RetrieveData("second_view")
	if err != nil {
		t.Fatalf("Failed to read second_view: %v", err)
	}
	if len(rows2) != 1 {
		t.Fatalf("Expected one row, got %d rows: %v", len(rows2), rows2)
	}
	if len(rows2[0].Tags) != 1 {
		t.Errorf("Expected one tag, got %d tags: %v", len(rows2[0].Tags), rows2[0].Tags)
	}
	wantTags = []tag.Tag{{Key: k1, Value: "bar"}}
	for i, tag := range rows2[0].Tags {
		if wantTags[i].Key != tag.Key {
			t.Errorf("Wrong key for %d, want %q, got %q", i, wantTags[i].Key, tag.Key)
		}
		if wantTags[i].Value != tag.Value {
			t.Errorf("Wrong value for tag %s, want %q got %q", tag.Key, wantTags[i].Value, tag.Value)
		}
	}
	gotCount := rows2[0].Data.(*view.CountData)
	if gotCount.Value != 1 {
		t.Errorf("Wrong count for second_view, want %d, got %d", 1, gotCount.Value)
	}
}
