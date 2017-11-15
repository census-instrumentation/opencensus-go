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

package trace

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

var (
	tid = TraceID{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 4, 8, 16, 32, 64, 128}
	sid = SpanID{1, 2, 4, 8, 16, 32, 64, 128}
)

func init() {
	// no random sampling, but sample children of sampled spans.
	SetDefaultSampler(ProbabilitySampler(0))
}

func TestStrings(t *testing.T) {
	if got, want := tid.String(), "01020304050607080102040810204080"; got != want {
		t.Errorf("TraceID.String: got %q want %q", got, want)
	}
	if got, want := sid.String(), "0102040810204080"; got != want {
		t.Errorf("SpanID.String: got %q want %q", got, want)
	}
}

func TestFromContext(t *testing.T) {
	want := &Span{}
	ctx := WithSpan(context.Background(), want)
	got := FromContext(ctx)
	if got != want {
		t.Errorf("got Span pointer %p want %p", got, want)
	}
}

type foo int

func (f foo) String() string {
	return "foo"
}

// checkChild tests that c has fields set appropriately, given that it is a child span of p.
func checkChild(p SpanContext, c *Span, expectRecordingEvents bool) error {
	if c == nil {
		return fmt.Errorf("got nil child span, want non-nil")
	}
	if got, want := c.spanContext.TraceID, p.TraceID; got != want {
		return fmt.Errorf("got child trace ID %s, want %s", got, want)
	}
	if childID, parentID := c.spanContext.SpanID, p.SpanID; childID == parentID {
		return fmt.Errorf("got child span ID %s, parent span ID %s; want unequal IDs", childID, parentID)
	}
	if got, want := c.spanContext.TraceOptions, p.TraceOptions; got != want {
		return fmt.Errorf("got child trace options %d, want %d", got, want)
	}
	if expectRecordingEvents {
		if c.data == nil {
			return fmt.Errorf("child span is not recording events")
		}
		if got, want := c.data.ParentSpanID, p.SpanID; got != want {
			return fmt.Errorf("got ParentSpanID %s, want %s", got, want)
		}
	} else {
		if c.data != nil {
			return fmt.Errorf("child span is recording events")
		}
	}
	return nil
}

func TestStartSpan(t *testing.T) {
	ctx := StartSpan(context.Background(), "StartSpan")
	if FromContext(ctx).data != nil {
		t.Error("StartSpan: new span is recording events")
	}
	sc, ok := SpanContextFromContext(ctx)
	if !ok {
		t.Error("StartSpan: no span in context")
	}
	if sc.SpanID == (SpanID{}) {
		t.Error("StartSpan: got zero SpanID, want nonzero")
	}
	if sc.TraceOptions != 0 {
		t.Errorf("StartSpan: got TraceOptions %x, want 0", sc.TraceOptions)
	}
}

func TestSampling(t *testing.T) {
	for _, test := range []struct {
		remoteParent        bool
		localParent         bool
		parentTraceOptions  TraceOptions
		recordEvents        bool
		sampler             Sampler
		wantRecordingEvents bool
		wantTraceOptions    TraceOptions
	}{
		{true, false, 0, false, nil, false, 0},
		{true, false, 0, true, nil, true, 0},
		{true, false, 1, false, nil, true, 1},
		{true, false, 1, true, nil, true, 1},
		{true, false, 0, false, NeverSample(), false, 0},
		{true, false, 0, true, NeverSample(), true, 0},
		{true, false, 1, false, NeverSample(), false, 0},
		{true, false, 1, true, NeverSample(), true, 0},
		{true, false, 0, false, AlwaysSample(), true, 1},
		{true, false, 0, true, AlwaysSample(), true, 1},
		{true, false, 1, false, AlwaysSample(), true, 1},
		{true, false, 1, true, AlwaysSample(), true, 1},
		{false, true, 0, false, NeverSample(), false, 0},
		{false, true, 0, true, NeverSample(), true, 0},
		{false, true, 1, false, NeverSample(), false, 0},
		{false, true, 1, true, NeverSample(), true, 0},
		{false, true, 0, false, AlwaysSample(), true, 1},
		{false, true, 0, true, AlwaysSample(), true, 1},
		{false, true, 1, false, AlwaysSample(), true, 1},
		{false, true, 1, true, AlwaysSample(), true, 1},
		{false, false, 0, false, nil, false, 0},
		{false, false, 0, true, nil, true, 0},
		{false, false, 0, false, NeverSample(), false, 0},
		{false, false, 0, true, NeverSample(), true, 0},
		{false, false, 0, false, AlwaysSample(), true, 1},
		{false, false, 0, true, AlwaysSample(), true, 1},
	} {
		var ctx context.Context
		if test.remoteParent {
			sc := SpanContext{
				TraceID:      tid,
				SpanID:       sid,
				TraceOptions: test.parentTraceOptions,
			}
			ctx = StartSpanWithRemoteParent(context.Background(), "foo", sc, StartSpanOptions{
				RecordEvents: test.recordEvents,
				Sampler:      test.sampler,
			})
		} else if test.localParent {
			sampler := NeverSample()
			if test.parentTraceOptions == 1 {
				sampler = AlwaysSample()
			}
			ctx2 := StartSpanWithOptions(context.Background(), "foo", StartSpanOptions{Sampler: sampler})
			ctx = StartSpanWithOptions(ctx2, "foo", StartSpanOptions{
				RecordEvents: test.recordEvents,
				Sampler:      test.sampler,
			})
		} else {
			ctx = StartSpanWithOptions(context.Background(), "foo", StartSpanOptions{
				RecordEvents: test.recordEvents,
				Sampler:      test.sampler,
			})
		}
		sc, ok := SpanContextFromContext(ctx)
		if !ok {
			t.Errorf("case %#v: starting new span: no span in context", test)
			continue
		}
		if sc.SpanID == (SpanID{}) {
			t.Errorf("case %#v: starting new span: got zero SpanID, want nonzero", test)
		}
		if recording := FromContext(ctx).data != nil; recording != test.wantRecordingEvents {
			t.Errorf("case %#v: starting new span: recording events is %t, want %t", test, recording, test.wantRecordingEvents)
		}
		if sc.TraceOptions != test.wantTraceOptions {
			t.Errorf("case %#v: starting new span: got TraceOptions %x, want %x", test, sc.TraceOptions, test.wantTraceOptions)
		}
	}

	// Test that for children of local spans, the default sampler has no effect.
	for _, test := range []struct {
		parentTraceOptions  TraceOptions
		recordEvents        bool
		wantRecordingEvents bool
		wantTraceOptions    TraceOptions
	}{
		{0, false, false, 0},
		{0, true, true, 0},
		{1, false, true, 1},
		{1, true, true, 1},
	} {
		for _, defaultSampler := range []Sampler{
			NeverSample(),
			AlwaysSample(),
			ProbabilitySampler(0),
		} {
			SetDefaultSampler(defaultSampler)
			sampler := NeverSample()
			if test.parentTraceOptions == 1 {
				sampler = AlwaysSample()
			}
			ctx2 := StartSpanWithOptions(context.Background(), "foo", StartSpanOptions{Sampler: sampler})
			ctx := StartSpanWithOptions(ctx2, "foo", StartSpanOptions{
				RecordEvents: test.recordEvents,
			})
			sc, ok := SpanContextFromContext(ctx)
			if !ok {
				t.Errorf("case %#v: starting new child of local span: no span in context", test)
				continue
			}
			if sc.SpanID == (SpanID{}) {
				t.Errorf("case %#v: starting new child of local span: got zero SpanID, want nonzero", test)
			}
			if recording := FromContext(ctx).data != nil; recording != test.wantRecordingEvents {
				t.Errorf("case %#v: starting new child of local span: recording events is %t, want %t", test, recording, test.wantRecordingEvents)
			}
			if sc.TraceOptions != test.wantTraceOptions {
				t.Errorf("case %#v: starting new child of local span: got TraceOptions %x, want %x", test, sc.TraceOptions, test.wantTraceOptions)
			}
		}
	}
	SetDefaultSampler(ProbabilitySampler(0)) // reset the default sampler.
}

func TestProbabilitySampler(t *testing.T) {
	exported := 0
	for i := 0; i < 1000; i++ {
		ctx := StartSpanWithOptions(context.Background(), "foo", StartSpanOptions{
			Sampler: ProbabilitySampler(0.3),
		})
		if IsSampled(ctx) {
			exported++
		}
	}
	if exported < 200 || exported > 400 {
		t.Errorf("got %f%% exported spans, want approximately 30%%", float64(exported)*0.1)
	}
}

func TestStartSpanWithRemoteParent(t *testing.T) {
	sc := SpanContext{
		TraceID:      tid,
		SpanID:       sid,
		TraceOptions: 0x0,
	}
	ctx := StartSpanWithRemoteParent(context.Background(), "StartSpanWithRemoteParent", sc, StartSpanOptions{})
	if err := checkChild(sc, FromContext(ctx), false); err != nil {
		t.Error(err)
	}

	ctx = StartSpanWithRemoteParent(context.Background(), "StartSpanWithRemoteParent", sc, StartSpanOptions{RecordEvents: true})
	if err := checkChild(sc, FromContext(ctx), true); err != nil {
		t.Error(err)
	}

	sc = SpanContext{
		TraceID:      tid,
		SpanID:       sid,
		TraceOptions: 0x1,
	}
	ctx = StartSpanWithRemoteParent(context.Background(), "StartSpanWithRemoteParent", sc, StartSpanOptions{})
	if err := checkChild(sc, FromContext(ctx), true); err != nil {
		t.Error(err)
	}

	ctx = StartSpanWithRemoteParent(context.Background(), "StartSpanWithRemoteParent", sc, StartSpanOptions{RecordEvents: true})
	if err := checkChild(sc, FromContext(ctx), true); err != nil {
		t.Error(err)
	}

	ctx2 := StartSpan(ctx, "StartSpan")
	parent, _ := SpanContextFromContext(ctx)
	if err := checkChild(parent, FromContext(ctx2), true); err != nil {
		t.Error(err)
	}
}

// startSpan returns a context with a new Span that is recording events and will be exported.
func startSpan() context.Context {
	return StartSpanWithRemoteParent(context.Background(), "span0",
		SpanContext{
			TraceID:      tid,
			SpanID:       sid,
			TraceOptions: 1,
		},
		StartSpanOptions{
			RecordEvents: true,
		})
}

type testExporter struct {
	spans []*SpanData
}

func (t *testExporter) Export(s *SpanData) {
	t.spans = append(t.spans, s)
}

// endSpan ends the Span in the context and returns the exported SpanData.
//
// It also does some tests on the Span, and tests and clears some fields in the SpanData.
func endSpan(ctx context.Context) (*SpanData, error) {
	if !IsRecordingEvents(ctx) {
		return nil, fmt.Errorf("IsRecordingEvents: got false, want true")
	}
	if !IsSampled(ctx) {
		return nil, fmt.Errorf("IsSampled: got false, want true")
	}
	var te testExporter
	RegisterExporter(&te)
	EndSpan(ctx)
	UnregisterExporter(&te)
	if len(te.spans) != 1 {
		return nil, fmt.Errorf("got exported spans %#v, want one span", te.spans)
	}
	got := te.spans[0]
	if got.SpanContext.SpanID == (SpanID{}) {
		return nil, fmt.Errorf("exporting span: expected nonzero SpanID")
	}
	got.SpanContext.SpanID = SpanID{}
	if !checkTime(&got.StartTime) {
		return nil, fmt.Errorf("exporting span: expected nonzero StartTime")
	}
	if !checkTime(&got.EndTime) {
		return nil, fmt.Errorf("exporting span: expected nonzero EndTime")
	}
	return got, nil
}

// checkTime checks that a nonzero time was set in x, then clears it.
func checkTime(x *time.Time) bool {
	if x.IsZero() {
		return false
	}
	*x = time.Time{}
	return true
}

func TestStackTrace(t *testing.T) {
	ctx := startSpan()
	SetStackTrace(ctx)
	got, err := endSpan(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(got.StackTrace) == 0 || got.StackTrace[0] == 0 {
		t.Error("exporting span: expected stack trace")
	}
	got.StackTrace = nil

	want := &SpanData{
		SpanContext: SpanContext{
			TraceID:      tid,
			SpanID:       SpanID{},
			TraceOptions: 0x1,
		},
		ParentSpanID:    sid,
		Name:            "span0",
		HasRemoteParent: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("exporting span: got %#v want %#v", got, want)
	}
}

func TestSetSpanAttributes(t *testing.T) {
	ctx := startSpan()
	SetSpanAttributes(ctx, StringAttribute{"key1", "value1"})
	got, err := endSpan(ctx)
	if err != nil {
		t.Fatal(err)
	}

	want := &SpanData{
		SpanContext: SpanContext{
			TraceID:      tid,
			SpanID:       SpanID{},
			TraceOptions: 0x1,
		},
		ParentSpanID:    sid,
		Name:            "span0",
		Attributes:      map[string]interface{}{"key1": "value1"},
		HasRemoteParent: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("exporting span: got %#v want %#v", got, want)
	}
}

func TestAnnotations(t *testing.T) {
	ctx := startSpan()
	LazyPrint(ctx, foo(1))
	LazyPrintWithAttributes(ctx, []Attribute{StringAttribute{"key2", "value2"}}, foo(2))
	LazyPrintf(ctx, "%f", -1.5)
	LazyPrintfWithAttributes(ctx, []Attribute{StringAttribute{"key3", "value3"}}, "%f", 1.5)
	Print(ctx, "Print")
	PrintWithAttributes(ctx, []Attribute{StringAttribute{"key4", "value4"}}, "PrintWithAttributes")
	got, err := endSpan(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for i := range got.Annotations {
		if !checkTime(&got.Annotations[i].Time) {
			t.Error("exporting span: expected nonzero Annotation Time")
		}
	}

	want := &SpanData{
		SpanContext: SpanContext{
			TraceID:      tid,
			SpanID:       SpanID{},
			TraceOptions: 0x1,
		},
		ParentSpanID: sid,
		Name:         "span0",
		Annotations: []Annotation{
			Annotation{Message: "foo", Attributes: nil},
			Annotation{Message: "foo", Attributes: map[string]interface{}{"key2": "value2"}},
			Annotation{Message: "-1.500000", Attributes: nil},
			Annotation{Message: "1.500000", Attributes: map[string]interface{}{"key3": "value3"}},
			Annotation{Message: "Print", Attributes: nil},
			Annotation{Message: "PrintWithAttributes", Attributes: map[string]interface{}{"key4": "value4"}},
		},
		HasRemoteParent: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("exporting span: got %#v want %#v", got, want)
	}
}

func TestMessageEvents(t *testing.T) {
	ctx := startSpan()
	AddMessageReceiveEvent(ctx, 3, 400, 300)
	AddMessageSendEvent(ctx, 1, 200, 100)
	got, err := endSpan(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for i := range got.MessageEvents {
		if !checkTime(&got.MessageEvents[i].Time) {
			t.Error("exporting span: expected nonzero MessageEvent Time")
		}
	}

	want := &SpanData{
		SpanContext: SpanContext{
			TraceID:      tid,
			SpanID:       SpanID{},
			TraceOptions: 0x1,
		},
		ParentSpanID: sid,
		Name:         "span0",
		MessageEvents: []MessageEvent{
			MessageEvent{EventType: 2, MessageID: 0x3, UncompressedByteSize: 0x190, CompressedByteSize: 0x12c},
			MessageEvent{EventType: 1, MessageID: 0x1, UncompressedByteSize: 0xc8, CompressedByteSize: 0x64},
		},
		HasRemoteParent: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("exporting span: got %#v want %#v", got, want)
	}
}

func TestSetSpanStatus(t *testing.T) {
	ctx := startSpan()
	SetSpanStatus(ctx, Status{Code: int32(1), Message: "request failed"})
	got, err := endSpan(ctx)
	if err != nil {
		t.Fatal(err)
	}

	want := &SpanData{
		SpanContext: SpanContext{
			TraceID:      tid,
			SpanID:       SpanID{},
			TraceOptions: 0x1,
		},
		ParentSpanID:    sid,
		Name:            "span0",
		Status:          Status{Code: 1, Message: "request failed"},
		HasRemoteParent: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("exporting span: got %#v want %#v", got, want)
	}
}

func TestAddLink(t *testing.T) {
	ctx := startSpan()
	AddLink(ctx, Link{
		TraceID:    tid,
		SpanID:     sid,
		Type:       LinkTypeParent,
		Attributes: map[string]interface{}{"key5": "value5"},
	})
	got, err := endSpan(ctx)
	if err != nil {
		t.Fatal(err)
	}

	want := &SpanData{
		SpanContext: SpanContext{
			TraceID:      tid,
			SpanID:       SpanID{},
			TraceOptions: 0x1,
		},
		ParentSpanID: sid,
		Name:         "span0",
		Links: []Link{Link{
			TraceID:    tid,
			SpanID:     sid,
			Type:       2,
			Attributes: map[string]interface{}{"key5": "value5"},
		}},
		HasRemoteParent: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("exporting span: got %#v want %#v", got, want)
	}
}

func TestUnregisterExporter(t *testing.T) {
	var te testExporter
	RegisterExporter(&te)
	UnregisterExporter(&te)

	ctx := startSpan()
	endSpan(ctx)
	if len(te.spans) != 0 {
		t.Error("unregistered Exporter was called")
	}
}

func TestBucket(t *testing.T) {
	// make a bucket of size 5 and add 10 spans
	b := makeBucket(5)
	for i := 1; i <= 10; i++ {
		b.nextTime = time.Time{} // reset the time so that the next span is accepted.
		// add a span, with i stored in the TraceID so we can test for it later.
		b.add(&SpanData{SpanContext: SpanContext{TraceID: TraceID{byte(i)}}, EndTime: time.Now()})
		if i <= 5 {
			if b.size() != i {
				t.Fatalf("got bucket size %d, want %d %#v\n", b.size(), i, b)
			}
			for j := 0; j < i; j++ {
				if b.span(j).TraceID[0] != byte(j+1) {
					t.Errorf("got span index %d, want %d\n", b.span(j).TraceID[0], j+1)
				}
			}
		} else {
			if b.size() != 5 {
				t.Fatalf("got bucket size %d, want 5\n", b.size())
			}
			for j := 0; j < 5; j++ {
				want := i - 4 + j
				if b.span(j).TraceID[0] != byte(want) {
					t.Errorf("got span index %d, want %d\n", b.span(j).TraceID[0], want)
				}
			}
		}
	}
	// expand the bucket
	b.resize(20)
	if b.size() != 5 {
		t.Fatalf("after resizing upwards: got bucket size %d, want 5\n", b.size())
	}
	for i := 0; i < 5; i++ {
		want := 6 + i
		if b.span(i).TraceID[0] != byte(want) {
			t.Errorf("after resizing upwards: got span index %d, want %d\n", b.span(i).TraceID[0], want)
		}
	}
	// shrink the bucket
	b.resize(3)
	if b.size() != 3 {
		t.Fatalf("after resizing downwards: got bucket size %d, want 3\n", b.size())
	}
	for i := 0; i < 3; i++ {
		want := 8 + i
		if b.span(i).TraceID[0] != byte(want) {
			t.Errorf("after resizing downwards: got span index %d, want %d\n", b.span(i).TraceID[0], want)
		}
	}
}
