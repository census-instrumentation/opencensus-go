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

func TestResolveOptions(t *testing.T) {
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
	view.SetReportingPeriod(100 * time.Millisecond)
	if err := view.Register(v...); err != nil {
		t.Fatalf("Failed to register view: %v", err)
	}
	defer view.Unregister(v...)

	attachments := map[string]interface{}{metricdata.AttachmentKeySpanContext: spanCtx}
	ctx, err := tag.New(context.Background(), tag.Insert(k1, "foo"), tag.Insert(k2, "foo"))
	if err != nil {
		t.Fatalf("Failed to set context: %v", err)
	}
	ro, err := stats.ResolveOptions(ctx,
		stats.WithTags(tag.Upsert(k1, "bar"), tag.Insert(k2, "bar")),
		stats.WithAttachments(attachments),
		stats.WithMeasurements(m1.M(12), m2.M(5)))
	if err != nil {
		t.Fatalf("Failed to resolve data point: %v", err)
	}

	s, ok := ro.Attachments[metricdata.AttachmentKeySpanContext]
	if !ok || s != spanCtx {
		t.Errorf("Unexpected SpanContext: want %v, got %v", spanCtx, s)
	}
	if len(ro.Attachments) != 1 {
		t.Errorf("Expected only one attachment (SpanContext), got %v", ro.Attachments)
	}

	if len(ro.Measures) != 2 {
		t.Errorf("Expected two measurements, got %v", ro.Measures)
	}
	mWant := []stats.Measurement{m1.M(12), m2.M(5)}
	if ro.Measures[0] != mWant[0] || ro.Measures[1] != mWant[1] {
		t.Errorf("Unexpected measurements: want %v, got %v", mWant, ro.Measures)
	}

	// k2 was Insert() ed, and shouldn't update the value that was in the supplied context.
	tCtx, err := tag.New(context.Background(), tag.Insert(k1, "bar"), tag.Insert(k2, "foo"))
	if err != nil {
		t.Fatalf("Failed to construct tWant: %v", err)
	}
	tWant := tag.FromContext(tCtx)
	if ro.Tags.String() != tWant.String() {
		t.Errorf("Unexpected tags: want %v, got %v", tWant, ro.Tags)
	}
}
