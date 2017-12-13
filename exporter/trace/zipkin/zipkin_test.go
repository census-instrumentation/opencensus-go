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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/openzipkin/zipkin-go/model"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (r roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return r(req)
}

func TestExport(t *testing.T) {
	ch := make(chan []byte)
	now := time.Now()
	client := http.Client{
		Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
			body, _ := ioutil.ReadAll(req.Body)
			ch <- body
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(""))}, nil
		}),
	}
	exporter, err := NewExporter("foo", nil, httpreporter.Client(&client), httpreporter.BatchInterval(time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	exporter.Export(&trace.SpanData{
		SpanContext: trace.SpanContext{
			TraceID:      trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			SpanID:       trace.SpanID{17, 18, 19, 20, 21, 22, 23, 24},
			TraceOptions: 1,
		},
		Name:      "name",
		StartTime: now,
		EndTime:   now.Add(24 * time.Hour),
		Attributes: map[string]interface{}{
			"stringkey": "value",
			"intkey":    int64(42),
			"boolkey":   true,
		},
		MessageEvents: []trace.MessageEvent{
			trace.MessageEvent{
				Time:                 now,
				EventType:            trace.MessageEventTypeSent,
				MessageID:            12,
				UncompressedByteSize: 99,
				CompressedByteSize:   98,
			},
		},
		Status: trace.Status{
			Code:    3,
			Message: "error",
		},
	})
	var data []byte
	select {
	case data = <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("span was not exported")
	}
	// alter some fields that are custom-encoded by model.SpanModel.MarshalJSON so
	// that we can unmarshal them.
	data = regexp.MustCompile(`"timestamp":[0-9]*`).ReplaceAll(data, []byte(`"timestamp":"2006-01-02T15:04:05+07:00"`))
	data = regexp.MustCompile(`"id":"1112131415161718"`).ReplaceAll(data, []byte(`"id":1`))
	var got []model.SpanModel
	json.Unmarshal(data, &got)
	if len(got) != 1 {
		t.Fatalf("Export: got %d spans, want 1", len(got))
	}
	got[0].Timestamp = time.Time{}
	if len(got[0].Annotations) != 1 {
		t.Fatalf("Export: got %d annotations in span, want 1", len(got[0].Annotations))
	}
	got[0].Annotations[0].Timestamp = time.Time{}
	want := []model.SpanModel{
		model.SpanModel{
			SpanContext: model.SpanContext{
				TraceID: model.TraceID{
					High: 0x102030405060708,
					Low:  0x90a0b0c0d0e0f10,
				},
				ID: 0x1,
			},
			Name:      "name",
			Kind:      "CLIENT",
			Timestamp: time.Time{},
			Duration:  86400000000,
			Shared:    false,
			Annotations: []model.Annotation{
				model.Annotation{
					Timestamp: time.Time{},
					Value:     "SENT",
				},
			},
			Tags: map[string]string{
				"stringkey":                 "value",
				"intkey":                    "42",
				"boolkey":                   "true",
				"census.status_code":        "INVALID_ARGUMENT",
				"census.status_description": "error",
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}
