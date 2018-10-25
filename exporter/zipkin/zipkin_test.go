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

package zipkin

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	openzipkin "github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (r roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

func TestExport(t *testing.T) {
	// Since Zipkin reports in microsecond resolution let's round our Timestamp,
	// so when deserializing Zipkin data in this test we can properly compare.
	now := time.Now().Round(time.Microsecond)
	tests := []struct {
		span *trace.SpanData
		want model.SpanModel
	}{
		{
			span: &trace.SpanData{
				SpanContext: trace.SpanContext{
					TraceID:      trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					SpanID:       trace.SpanID{17, 18, 19, 20, 21, 22, 23, 24},
					TraceOptions: 1,
				},
				Name:      "name",
				SpanKind:  trace.SpanKindClient,
				StartTime: now,
				EndTime:   now.Add(24 * time.Hour),
				Attributes: map[string]interface{}{
					"stringkey": "value",
					"intkey":    int64(42),
					"boolkey1":  true,
					"boolkey2":  false,
				},
				MessageEvents: []trace.MessageEvent{
					{
						Time:                 now,
						EventType:            trace.MessageEventTypeSent,
						MessageID:            12,
						UncompressedByteSize: 99,
						CompressedByteSize:   98,
					},
				},
				Annotations: []trace.Annotation{
					{
						Time:    now,
						Message: "Annotation",
						Attributes: map[string]interface{}{
							"stringkey": "value",
							"intkey":    int64(42),
							"boolkey1":  true,
							"boolkey2":  false,
						},
					},
				},
				Status: trace.Status{
					Code:    3,
					Message: "error",
				},
			},
			want: model.SpanModel{
				SpanContext: model.SpanContext{
					TraceID: model.TraceID{
						High: 0x0102030405060708,
						Low:  0x090a0b0c0d0e0f10,
					},
					ID:      0x1112131415161718,
					Sampled: &sampledTrue,
				},
				Name:      "name",
				Kind:      model.Client,
				Timestamp: now,
				Duration:  24 * time.Hour,
				Shared:    false,
				Annotations: []model.Annotation{
					{
						Timestamp: now,
						Value:     "Annotation",
					},
					{
						Timestamp: now,
						Value:     "SENT",
					},
				},
				Tags: map[string]string{
					"stringkey":                     "value",
					"intkey":                        "42",
					"boolkey1":                      "true",
					"boolkey2":                      "false",
					"error":                         "INVALID_ARGUMENT",
					"opencensus.status_description": "error",
				},
			},
		},
		{
			span: &trace.SpanData{
				SpanContext: trace.SpanContext{
					TraceID:      trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					SpanID:       trace.SpanID{17, 18, 19, 20, 21, 22, 23, 24},
					TraceOptions: 1,
				},
				Name:      "name",
				StartTime: now,
				EndTime:   now.Add(24 * time.Hour),
			},
			want: model.SpanModel{
				SpanContext: model.SpanContext{
					TraceID: model.TraceID{
						High: 0x0102030405060708,
						Low:  0x090a0b0c0d0e0f10,
					},
					ID:      0x1112131415161718,
					Sampled: &sampledTrue,
				},
				Name:      "name",
				Timestamp: now,
				Duration:  24 * time.Hour,
				Shared:    false,
			},
		},
		{
			span: &trace.SpanData{
				SpanContext: trace.SpanContext{
					TraceID:      trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					SpanID:       trace.SpanID{17, 18, 19, 20, 21, 22, 23, 24},
					TraceOptions: 1,
				},
				Name:      "name",
				StartTime: now,
				EndTime:   now.Add(24 * time.Hour),
				Status: trace.Status{
					Code:    0,
					Message: "there is no cause for alarm",
				},
			},
			want: model.SpanModel{
				SpanContext: model.SpanContext{
					TraceID: model.TraceID{
						High: 0x0102030405060708,
						Low:  0x090a0b0c0d0e0f10,
					},
					ID:      0x1112131415161718,
					Sampled: &sampledTrue,
				},
				Name:      "name",
				Timestamp: now,
				Duration:  24 * time.Hour,
				Shared:    false,
				Tags: map[string]string{
					"opencensus.status_description": "there is no cause for alarm",
				},
			},
		},
		{
			span: &trace.SpanData{
				SpanContext: trace.SpanContext{
					TraceID:      trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
					SpanID:       trace.SpanID{17, 18, 19, 20, 21, 22, 23, 24},
					TraceOptions: 1,
				},
				Name:      "name",
				StartTime: now,
				EndTime:   now.Add(24 * time.Hour),
				Status: trace.Status{
					Code: 1234,
				},
			},
			want: model.SpanModel{
				SpanContext: model.SpanContext{
					TraceID: model.TraceID{
						High: 0x0102030405060708,
						Low:  0x090a0b0c0d0e0f10,
					},
					ID:      0x1112131415161718,
					Sampled: &sampledTrue,
				},
				Name:      "name",
				Timestamp: now,
				Duration:  24 * time.Hour,
				Shared:    false,
				Tags: map[string]string{
					"error": "error code 1234",
				},
			},
		},
	}
	for _, tt := range tests {
		got := zipkinSpan(tt.span, nil, nil)
		if len(got.Annotations) != len(tt.want.Annotations) {
			t.Fatalf("zipkinSpan: got %d annotations in span, want %d", len(got.Annotations), len(tt.want.Annotations))
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("zipkinSpan:\n\tgot  %#v\n\twant %#v", got, tt.want)
		}
	}
	for _, tt := range tests {
		ch := make(chan []byte)
		client := http.Client{
			Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
				body, _ := ioutil.ReadAll(req.Body)
				ch <- body
				return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(""))}, nil
			}),
		}
		reporter := httpreporter.NewReporter("foo", httpreporter.Client(&client), httpreporter.BatchInterval(time.Millisecond))
		exporter := NewExporter(reporter, nil)
		exporter.ExportSpan(tt.span)
		var data []byte
		select {
		case data = <-ch:
		case <-time.After(2 * time.Second):
			t.Fatalf("span was not exported")
		}
		var spans []model.SpanModel
		json.Unmarshal(data, &spans)
		if len(spans) != 1 {
			t.Fatalf("Export: got %d spans, want 1", len(spans))
		}
		got := spans[0]
		got.SpanContext.Sampled = &sampledTrue // Sampled is not set when the span is reported.
		if len(got.Annotations) != len(tt.want.Annotations) {
			t.Fatalf("Export: got %d annotations in span, want %d", len(got.Annotations), len(tt.want.Annotations))
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Export:\n\tgot  %#v\n\twant %#v", got, tt.want)
		}
	}
}

// Ensure that we can pass in a remote endpoint but also that it is
// transmitted to its origina. Issue #959
func TestRemoteEndpointOptionAndTransmission(t *testing.T) {
	type lockableBuffer struct {
		sync.Mutex
		*bytes.Buffer
	}

	buf := &lockableBuffer{Mutex: sync.Mutex{}, Buffer: new(bytes.Buffer)}

	cst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		blob, _ := ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
		buf.Lock()
		buf.Write(blob)
		buf.Unlock()
	}))
	defer cst.Close()

	reporter := httpreporter.NewReporter(cst.URL, httpreporter.BatchInterval(10*time.Millisecond))
	localEndpoint, _ := openzipkin.NewEndpoint("app", "10.0.0.17")
	remoteEndpoint, _ := openzipkin.NewEndpoint("memcached", "10.0.0.42")
	exp := NewExporter(reporter, localEndpoint, WithRemoteEndpoint(remoteEndpoint))
	exp.ExportSpan(&trace.SpanData{
		Name: "Test",
	})

	// Wait for the upload
	<-time.After(300 * time.Millisecond)

	want := `[{
            "traceId":"0000000000000000",
            "id":"0000000000000000",
            "name":"Test",
            "localEndpoint":{
                "serviceName":"app",
                "ipv4":"10.0.0.17"
            },
            "remoteEndpoint":{
                "serviceName":"memcached","ipv4":"10.0.0.42"
            }
        }]`

	buf.Lock()
	got := buf.String()
	buf.Unlock()

	// Since the reported JSON could contain spaces and other indentation,
	// strip spaces out but also the fields could be mangled so we'll instead
	// just use an anagram equivalence to ensure all the output is present
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "")
	wj := replacer.Replace(want)
	gj := replacer.Replace(got)
	if !anagrams(gj, wj) {
		t.Errorf("Mismatched JSON content\nGot:\n\t%s\nWant:\n\t%s", gj, wj)
	}
}

func anagrams(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	if s1 == "" && s2 == "" {
		return true
	}
	m1 := make(map[byte]int)
	for i := range s1 {
		m1[s1[i]] += 1
		m1[s2[i]] -= 1
	}

	// Finally check that all the values are at 0
	// that is, all the letters in s1 were matched in s2
	for _, count := range m1 {
		if count != 0 {
			return false
		}
	}
	return true
}
