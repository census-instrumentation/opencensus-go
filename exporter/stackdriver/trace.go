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
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.opencensus.io/internal"
	"go.opencensus.io/trace/propagation"

	tracingclient "cloud.google.com/go/trace/apiv2"
	"go.opencensus.io/trace"
	"google.golang.org/api/option"
	"google.golang.org/api/support/bundler"
	tracepb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v2"
)

const (
	httpHeader        = `X-Cloud-Trace-Context`
	httpHeaderMaxSize = 200
)

// traceExporter is an implementation of trace.Exporter that uploads spans to
// Stackdriver.
//
// traceExporter also implements trace/propagation.HTTPFormat and can
// propagate Stackdriver Traces over HTTP requests.
type traceExporter struct {
	projectID string
	bundler   *bundler.Bundler
	// uploadFn defaults to uploadToStackdriver; it can be replaced for tests.
	uploadFn func(spans []*trace.SpanData)
	overflowLogger
	client *tracingclient.Client
}

var _ trace.Exporter = (*traceExporter)(nil)
var _ propagation.HTTPFormat = (*traceExporter)(nil)

func newTraceExporter(o Options) (*traceExporter, error) {
	co := []option.ClientOption{
		option.WithUserAgent(internal.UserAgent),
		// NB: NewClient below adds WithEndpoint, WithScopes options.
	}
	co = append(co, o.ClientOptions...)
	client, err := tracingclient.NewClient(context.Background(), co...)
	if err != nil {
		return nil, fmt.Errorf("stackdriver: couldn't initialize trace client: %v", err)
	}
	return newTraceExporterWithClient(o, client), nil
}

func newTraceExporterWithClient(o Options, c *tracingclient.Client) *traceExporter {
	e := &traceExporter{
		projectID: o.ProjectID,
		client:    c,
	}
	bundler := bundler.NewBundler((*trace.SpanData)(nil), func(bundle interface{}) {
		e.uploadFn(bundle.([]*trace.SpanData))
	})
	if o.BundleDelayThreshold > 0 {
		bundler.DelayThreshold = o.BundleDelayThreshold
	} else {
		bundler.DelayThreshold = 2 * time.Second
	}
	if o.BundleCountThreshold > 0 {
		bundler.BundleCountThreshold = o.BundleCountThreshold
	} else {
		bundler.BundleCountThreshold = 50
	}
	// The measured "bytes" are not really bytes, see exportReceiver.
	bundler.BundleByteThreshold = bundler.BundleCountThreshold * 200
	bundler.BundleByteLimit = bundler.BundleCountThreshold * 1000
	bundler.BufferedByteLimit = bundler.BundleCountThreshold * 2000

	e.bundler = bundler
	e.uploadFn = e.uploadToStackdriver
	return e
}

// ExportSpan exports a SpanData to Stackdriver Trace.
func (e *traceExporter) ExportSpan(s *trace.SpanData) {
	// n is a length heuristic.
	n := 1
	n += len(s.Attributes)
	n += len(s.Annotations)
	n += len(s.MessageEvents)
	n += len(s.StackTrace)
	err := e.bundler.Add(s, n)
	switch err {
	case nil:
		return
	case bundler.ErrOversizedItem:
		go e.uploadFn([]*trace.SpanData{s})
	case bundler.ErrOverflow:
		e.overflowLogger.log()
	default:
		log.Println("OpenCensus Stackdriver exporter: failed to upload span:", err)
	}
}

// Flush waits for exported trace spans to be uploaded.
//
// This is useful if your program is ending and you do not want to lose recent
// spans.
func (e *traceExporter) Flush() {
	e.bundler.Flush()
}

// uploadToStackdriver uploads a set of spans to Stackdriver.
func (e *traceExporter) uploadToStackdriver(spans []*trace.SpanData) {
	req := tracepb.BatchWriteSpansRequest{
		Name:  "projects/" + e.projectID,
		Spans: make([]*tracepb.Span, 0, len(spans)),
	}
	for _, span := range spans {
		req.Spans = append(req.Spans, protoFromSpanData(span, e.projectID))
	}
	err := e.client.BatchWriteSpans(context.Background(), &req)
	if err != nil {
		log.Printf("OpenCensus Stackdriver exporter: failed to upload %d spans: %v", len(spans), err)
	}
}

// FromRequest extracts a Stakdriver Trace span context from incoming requests.
func (e *traceExporter) FromRequest(req *http.Request) (sc trace.SpanContext, ok bool) {
	h := req.Header.Get(httpHeader)
	// See https://cloud.google.com/trace/docs/faq for the header format.
	// Return if the header is empty or missing, or if the header is unreasonably
	// large, to avoid making unnecessary copies of a large string.
	if h == "" || len(h) > httpHeaderMaxSize {
		return trace.SpanContext{}, false
	}

	// Parse the trace id field.
	slash := strings.Index(h, `/`)
	if slash == -1 {
		return trace.SpanContext{}, false
	}
	tid, h := h[:slash], h[slash+1:]

	buf, err := hex.DecodeString(tid)
	if err != nil {
		return trace.SpanContext{}, false
	}
	copy(sc.TraceID[:], buf)

	// Parse the span id field.
	spanstr := h
	semicolon := strings.Index(h, `;`)
	if semicolon != -1 {
		spanstr, h = h[:semicolon], h[semicolon+1:]
	}
	sid, err := strconv.ParseUint(spanstr, 10, 64)
	if err != nil {
		return trace.SpanContext{}, false
	}

	buf = make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, sid)
	copy(sc.SpanID[:], buf)

	// Parse the options field, options field is optional.
	if !strings.HasPrefix(h, "o=") {
		return sc, true
	}
	o, err := strconv.ParseUint(h[2:], 10, 64)
	if err != nil {
		return trace.SpanContext{}, false
	}
	sc.TraceOptions = trace.TraceOptions(o)
	return sc, true
}

// ToRequest modifies the given request to include a Stackdriver Trace header.
func (e *traceExporter) ToRequest(sc trace.SpanContext, req *http.Request) {
	sid, _ := binary.Uvarint(sc.SpanID[:])
	header := fmt.Sprintf("%s/%d;o=%d", hex.EncodeToString(sc.TraceID[:]), sid, int64(sc.TraceOptions))
	req.Header.Set(httpHeader, header)
}

// overflowLogger ensures that at most one overflow error log message is
// written every 5 seconds.
type overflowLogger struct {
	mu    sync.Mutex
	pause bool
	accum int
}

func (o *overflowLogger) delay() {
	o.pause = true
	time.AfterFunc(5*time.Second, func() {
		o.mu.Lock()
		defer o.mu.Unlock()
		switch {
		case o.accum == 0:
			o.pause = false
		case o.accum == 1:
			log.Println("OpenCensus Stackdriver exporter: failed to upload span: buffer full")
			o.accum = 0
			o.delay()
		default:
			log.Printf("OpenCensus Stackdriver exporter: failed to upload %d spans: buffer full", o.accum)
			o.accum = 0
			o.delay()
		}
	})
}

func (o *overflowLogger) log() {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.pause {
		log.Println("OpenCensus Stackdriver exporter: failed to upload span: buffer full")
		o.delay()
	} else {
		o.accum++
	}
}
