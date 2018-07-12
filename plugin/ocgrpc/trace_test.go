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

package ocgrpc_test

import (
	"io"
	"testing"
	"time"

	"encoding/hex"
	"encoding/json"

	"go.opencensus.io/internal/testpb"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"golang.org/x/net/context"
)

type testExporter struct {
	ch chan *trace.SpanData
}

func (t *testExporter) ExportSpan(s *trace.SpanData) {
	go func() { t.ch <- s }()
}

func TestStreaming(t *testing.T) {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	te := testExporter{make(chan *trace.SpanData)}
	trace.RegisterExporter(&te)
	defer trace.UnregisterExporter(&te)

	client, cleanup := testpb.NewTestClient(t, &ocgrpc.ClientHandler{}, &ocgrpc.ServerHandler{})

	stream, err := client.Multiple(context.Background())
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	err = stream.Send(&testpb.FooRequest{})
	if err != nil {
		t.Fatalf("Couldn't send streaming request: %v", err)
	}
	stream.CloseSend()

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Errorf("stream.Recv() = %v; want no errors", err)
		}
	}

	cleanup()

	s1 := <-te.ch
	s2 := <-te.ch

	checkSpanData(t, s1, s2, "testpb.Foo.Multiple", true)

	select {
	case <-te.ch:
		t.Fatal("received extra exported spans")
	case <-time.After(time.Second / 10):
	}
}

func TestStreamingFail(t *testing.T) {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	te := testExporter{make(chan *trace.SpanData)}
	trace.RegisterExporter(&te)
	defer trace.UnregisterExporter(&te)

	client, cleanup := testpb.NewTestClient(t, &ocgrpc.ClientHandler{}, &ocgrpc.ServerHandler{})

	stream, err := client.Multiple(context.Background())
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	err = stream.Send(&testpb.FooRequest{Fail: true})
	if err != nil {
		t.Fatalf("Couldn't send streaming request: %v", err)
	}
	stream.CloseSend()

	for {
		_, err := stream.Recv()
		if err == nil || err == io.EOF {
			t.Errorf("stream.Recv() = %v; want errors", err)
		} else {
			break
		}
	}

	s1 := <-te.ch
	s2 := <-te.ch

	checkSpanData(t, s1, s2, "testpb.Foo.Multiple", false)
	cleanup()

	select {
	case <-te.ch:
		t.Fatal("received extra exported spans")
	case <-time.After(time.Second / 10):
	}
}

func TestSingle(t *testing.T) {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	te := testExporter{make(chan *trace.SpanData)}
	trace.RegisterExporter(&te)
	defer trace.UnregisterExporter(&te)

	client, cleanup := testpb.NewTestClient(t, &ocgrpc.ClientHandler{}, &ocgrpc.ServerHandler{})

	_, err := client.Single(context.Background(), &testpb.FooRequest{})
	if err != nil {
		t.Fatalf("Couldn't send request: %v", err)
	}

	s1 := <-te.ch
	s2 := <-te.ch

	checkSpanData(t, s1, s2, "testpb.Foo.Single", true)
	cleanup()

	select {
	case <-te.ch:
		t.Fatal("received extra exported spans")
	case <-time.After(time.Second / 10):
	}
}

func TestServerSpanDuration(t *testing.T) {
	client, cleanup := testpb.NewTestClient(t, &ocgrpc.ClientHandler{}, &ocgrpc.ServerHandler{})
	defer cleanup()

	te := testExporter{make(chan *trace.SpanData, 100)}
	trace.RegisterExporter(&te)
	defer trace.UnregisterExporter(&te)

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	ctx := context.Background()
	const sleep = 100 * time.Millisecond
	client.Single(ctx, &testpb.FooRequest{SleepNanos: int64(sleep)})

loop:
	for {
		select {
		case span := <-te.ch:
			if span.SpanKind != trace.SpanKindServer {
				continue loop
			}
			if got, want := span.EndTime.Sub(span.StartTime), sleep; got < want {
				t.Errorf("span duration = %dns; want at least %dns", got, want)
			}
			break loop
		default:
			t.Fatal("no more spans")
		}
	}
}

func TestSingleFail(t *testing.T) {
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	te := testExporter{make(chan *trace.SpanData)}
	trace.RegisterExporter(&te)
	defer trace.UnregisterExporter(&te)

	client, cleanup := testpb.NewTestClient(t, &ocgrpc.ClientHandler{}, &ocgrpc.ServerHandler{})

	_, err := client.Single(context.Background(), &testpb.FooRequest{Fail: true})
	if err == nil {
		t.Fatalf("Got nil error from request, want non-nil")
	}

	s1 := <-te.ch
	s2 := <-te.ch

	checkSpanData(t, s1, s2, "testpb.Foo.Single", false)
	cleanup()

	select {
	case <-te.ch:
		t.Fatal("received extra exported spans")
	case <-time.After(time.Second / 10):
	}
}

func checkSpanData(t *testing.T, s1, s2 *trace.SpanData, methodName string, success bool) {
	t.Helper()

	if s1.SpanKind == trace.SpanKindServer {
		s1, s2 = s2, s1
	}

	if got, want := s1.Name, methodName; got != want {
		t.Errorf("Got name %q want %q", got, want)
	}
	if got, want := s2.Name, methodName; got != want {
		t.Errorf("Got name %q want %q", got, want)
	}
	if got, want := s2.SpanContext.TraceID, s1.SpanContext.TraceID; got != want {
		t.Errorf("Got trace IDs %s and %s, want them equal", got, want)
	}
	if got, want := s2.ParentSpanID, s1.SpanContext.SpanID; got != want {
		t.Errorf("Got ParentSpanID %s, want %s", got, want)
	}
	if got := (s1.Status.Code == 0); got != success {
		t.Errorf("Got success=%t want %t", got, success)
	}
	if got := (s2.Status.Code == 0); got != success {
		t.Errorf("Got success=%t want %t", got, success)
	}
	if s1.HasRemoteParent {
		t.Errorf("Got HasRemoteParent=%t, want false", s1.HasRemoteParent)
	}
	if !s2.HasRemoteParent {
		t.Errorf("Got HasRemoteParent=%t, want true", s2.HasRemoteParent)
	}
}

type testProp struct {
	t *testing.T
}

func (p *testProp) InjectSpanContext(sc trace.SpanContext, appendHeader func(string, string)) {
	traceHdr, err := json.Marshal(sc)
	if err != nil {
		p.t.Fatal(err)
	}
	appendHeader("x-json-trace", hex.EncodeToString(traceHdr))
}

func (p *testProp) ExtractSpanContext(readHeader func(string) []string) (trace.SpanContext, bool) {
	hdr := readHeader("x-json-trace")
	if len(hdr) == 0 {
		return trace.SpanContext{}, false
	}
	b, err := hex.DecodeString(hdr[0])
	if err != nil {
		p.t.Fatal(err)
	}
	var sc trace.SpanContext
	err = json.Unmarshal(b, &sc)
	if err != nil {
		p.t.Fatal(err)
	}
	return sc, true
}

func TestCustomPropagation(t *testing.T) {
	te := &testExporter{ch: make(chan *trace.SpanData, 1000)}
	trace.RegisterExporter(te)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	p := &testProp{t: t}
	clientHandler := &ocgrpc.ClientHandler{
		Propagation: p,
	}
	serverHandler := &ocgrpc.ServerHandler{
		Propagation: p,
	}
	client, done := testpb.NewTestClient(t, clientHandler, serverHandler)
	defer done()
	ctx := context.Background()
	_, err := client.Single(ctx, &testpb.FooRequest{})
	if err != nil {
		t.Fatal(err)
	}
	spans := 0
	var serverSpan, clientSpan *trace.SpanData
	for sd := range te.ch {
		spans++
		if sd.SpanKind == trace.SpanKindServer {
			serverSpan = sd
		} else if sd.SpanKind == trace.SpanKindClient {
			clientSpan = sd
		}
		if spans == 2 {
			break
		}
	}
	if clientSpan.TraceID != serverSpan.TraceID {
		t.Errorf("client and server spans don't have the same TraceID")
	}
	if got, want := serverSpan.ParentSpanID, clientSpan.SpanID; got != want {
		t.Errorf("serverSpan.ParentSpanID = %x, want %x", got, want)
	}
}
