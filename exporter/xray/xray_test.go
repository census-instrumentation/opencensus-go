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

package xray

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/aws/aws-sdk-go/service/xray/xrayiface"
	"go.opencensus.io/trace"
)

func makeTraceUrl(region, traceID string) string {
	return fmt.Sprintf("https://%v.console.aws.amazon.com/xray/home?region=us-west-2#/traces/%v\n", region, traceID)
}

func TestLiveExporter(t *testing.T) {
	if key, secret := os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"); key == "" || secret == "" {
		t.SkipNow()
	}

	var published = make(chan struct{})
	var onExport = func(in OnExport) {
		select {
		case <-published:
		default:
			close(published)
		}

		region := os.Getenv("AWS_DEFAULT_REGION")
		if region == "" {
			region = os.Getenv("AWS_REGION")
		}
		fmt.Println(makeTraceUrl(region, in.TraceID))
	}

	exporter, err := NewExporter(WithOnExport(onExport), WithOrigin(OriginECS))
	if err != nil {
		t.Errorf("expected nil; got %v", err)
	}

	trace.RegisterExporter(exporter)
	trace.SetDefaultSampler(trace.AlwaysSample())

	attributes := []trace.Attribute{
		trace.StringAttribute{Key: "key", Value: "value"},
	}

	ctx, parent := trace.StartSpan(context.Background(), "parent")
	parent.Annotate(attributes, "did the thing")
	parent.SetAttributes(trace.StringAttribute{Key: "hello", Value: "world"})

	time.Sleep(75 * time.Millisecond)
	_, child := trace.StartSpan(ctx, "child")
	child.SetStatus(trace.Status{
		Code:    500,
		Message: "boom!",
	})
	time.Sleep(100 * time.Millisecond)
	child.End()
	time.Sleep(150 * time.Millisecond)

	parent.End()

	<-published // don't close until the message has been sent
}

type mockSegments struct {
	xrayiface.XRayAPI
	ch chan segment
}

func (m *mockSegments) PutTraceSegments(in *xray.PutTraceSegmentsInput) (*xray.PutTraceSegmentsOutput, error) {
	for _, doc := range in.TraceSegmentDocuments {
		var s segment
		if err := json.Unmarshal([]byte(*doc), &s); err != nil {
			return nil, err
		}
		m.ch <- s
	}
	return nil, nil
}

type spec struct {
	Name     string
	Status   trace.Status
	Children []spec
}

func walk(ctx context.Context, input spec) {
	ctx, span := trace.StartSpan(ctx, input.Name)
	defer span.End()

	if input.Status.Code != 0 {
		span.SetStatus(input.Status)
	}

	for _, child := range input.Children {
		walk(ctx, child)
	}
}

func assertSegmentsEqual(t *testing.T, expected, actual segment) {
	if actual.ID == "" {
		t.Errorf("expected id to be set")
	}
	if expected.Name != actual.Name {
		t.Errorf("want name, %v; got %v", expected.Name, actual.Name)
	}
	if expected.Type != actual.Type {
		t.Errorf("want type, %v; got %v", expected.Type, actual.Type)
	}
	if expected.Error != actual.Error {
		t.Errorf("want fault, %v; got %v", expected.Error, actual.Error)
	}
	if expected.Fault != actual.Fault {
		t.Errorf("want fault, %v; got %v", expected.Fault, actual.Fault)
	}
	if e, a := expected.Cause, actual.Cause; e == nil && a != nil || e != nil && a == nil {
		t.Errorf("want cause, %#v; got %#v", expected.Cause, actual.Cause)
	} else if e != nil && a != nil {
		if len(e.Exceptions) != len(a.Exceptions) {
			t.Errorf("want exceptions, %#v; got %#v", e.Exceptions, a.Exceptions)
		} else {
			for index := range e.Exceptions {
				if e.Exceptions[index].Message != a.Exceptions[index].Message {
					t.Errorf("want message, %#v; got %#v", e.Exceptions[index].Message, a.Exceptions[index].Message)
				}
			}
		}
	}
}

func TestExporter(t *testing.T) {
	testCases := map[string]string{
		"simple span":   "testdata/simple.json",
		"parent child":  "testdata/parent-child.json",
		"deeply nested": "testdata/deeply-nested.json",
		"error":         "testdata/error.json",
		"fault":         "testdata/fault.json",
	}

	for label, filename := range testCases {
		t.Run(label, func(t *testing.T) {
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Fatalf("unable to open file, %v", filename)
			}

			// Given
			var (
				api     = &mockSegments{ch: make(chan segment, 16)}
				content struct {
					Input    spec
					Expected []segment
				}
			)

			if err := json.Unmarshal(data, &content); err != nil {
				t.Fatalf("unable to parse json file, %v", filename)
			}

			exporter, err := NewExporter(WithAPI(api), WithInterval(100*time.Millisecond))
			if err != nil {
				t.Fatalf("expected to create exporter; got %v", err)
			}
			trace.RegisterExporter(exporter)
			trace.SetDefaultSampler(trace.AlwaysSample())

			// When - we create a span structure
			walk(context.Background(), content.Input)

			for _, expected := range content.Expected {
				// Then
				select {
				case segment := <-api.ch:
					assertSegmentsEqual(t, expected, segment)

				case <-time.After(time.Second):
					t.Fatalf("timeout waiting for span to be processed")
				}
			}
		})
	}
}

func TestOptions(t *testing.T) {
	t.Run("SetOutput", func(t *testing.T) {
		var output = os.Stderr
		config, err := buildConfig(WithOutput(output))
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}
		if config.output != output {
			t.Fatalf("want %v; got %v", output, config.interval)
		}
	})

	t.Run("SetInterval", func(t *testing.T) {
		const interval = time.Minute
		config, err := buildConfig(WithInterval(interval))
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}
		if config.interval != interval {
			t.Fatalf("want %v; got %v", interval, config.interval)
		}
	})

	t.Run("SetBufferSize", func(t *testing.T) {
		const bufferSize = 15
		config, err := buildConfig(WithBufferSize(bufferSize))
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}
		if config.bufferSize != bufferSize {
			t.Fatalf("want %v; got %v", bufferSize, config.interval)
		}
	})

	t.Run("SetVersion", func(t *testing.T) {
		const version = "latest"
		config, err := buildConfig(WithVersion(version))
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		want := &service{
			Version: version,
		}
		if !reflect.DeepEqual(want, config.service) {
			t.Fatalf("want %v; got %v", want, config.service)
		}
	})

	t.Run("SetOrigin", func(t *testing.T) {
		const origin = OriginEB
		config, err := buildConfig(WithOrigin(origin))
		if err != nil {
			t.Fatalf("want nil; got %v", err)
		}

		if config.origin != origin {
			t.Fatalf("want %v; got %v", origin, config.origin)
		}
	})

	t.Run("end to end", func(t *testing.T) {
		var (
			version  = "blah"
			origin   = OriginEB
			exported = make(chan struct{})
			api      = &mockSegments{ch: make(chan segment, 1)}
			onExport = func(export OnExport) {
				select {
				case <-exported:
				default:
					close(exported)
				}
			}
			exporter, _ = NewExporter(
				WithAPI(api),
				WithOrigin(origin),
				WithVersion(version),
				WithOnExport(onExport),
				WithInterval(100*time.Millisecond),
			)
		)

		buildConfig()

		trace.RegisterExporter(exporter)
		trace.SetDefaultSampler(trace.AlwaysSample())

		// When
		_, span := trace.StartSpan(context.Background(), "span")
		span.End()

		// Then
		select {
		case segment := <-api.ch:
			if segment.Service == nil || segment.Service.Version != version {
				t.Errorf("expected %v; got %#v", version, segment.Service)
			}
			if string(origin) != segment.Origin {
				t.Errorf("expected %v; got %v", origin, segment.Origin)
			}

			select {
			case <-exported:
				//ok
			case <-time.After(time.Second):
				t.Errorf("timeout waiting for onExport to be called")
			}

		case <-time.After(time.Second):
			t.Errorf("timeout waiting for span to be processed")
		}
	})
}

func TestSetBufferSizeTrigger(t *testing.T) {
	var (
		api         = &mockSegments{ch: make(chan segment, 1)}
		exporter, _ = NewExporter(WithAPI(api), WithBufferSize(1))
	)

	trace.RegisterExporter(exporter)
	trace.SetDefaultSampler(trace.AlwaysSample())

	// When
	_, span := trace.StartSpan(context.Background(), "span")
	span.End()

	// Then
	select {
	case <-api.ch:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected 1 segment to have been flushed")
	}
}

func TestFlush(t *testing.T) {
	var (
		api         = &mockSegments{ch: make(chan segment, 1)}
		exporter, _ = NewExporter(WithAPI(api))
	)

	trace.RegisterExporter(exporter)
	trace.SetDefaultSampler(trace.AlwaysSample())

	_, span := trace.StartSpan(context.Background(), "span")
	span.End()

	// When
	exporter.Flush()

	// Then
	select {
	case <-api.ch:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("expected 1 segment to have been flushed")
	}
}

func TestClose(t *testing.T) {
	t.Run("flushes buffer", func(t *testing.T) {
		var (
			api         = &mockSegments{ch: make(chan segment, 1)}
			exporter, _ = NewExporter(WithAPI(api))
		)

		trace.RegisterExporter(exporter)
		trace.SetDefaultSampler(trace.AlwaysSample())

		_, span := trace.StartSpan(context.Background(), "span")
		span.End()

		// When
		exporter.Close()

		// Then
		select {
		case <-api.ch:
		case <-time.After(50 * time.Millisecond):
			t.Fatal("expected 1 segment to have been flushed")
		}
	})

	t.Run("additional messages dropped after exporter is Closed", func(t *testing.T) {
		var (
			api         = &mockSegments{ch: make(chan segment, 1)}
			exporter, _ = NewExporter(WithAPI(api))
		)

		trace.RegisterExporter(exporter)
		trace.SetDefaultSampler(trace.AlwaysSample())

		// When
		exporter.Close()

		// Then
		_, span := trace.StartSpan(context.Background(), "span")
		span.End()

		// Then
		select {
		case <-api.ch:
			t.Fatal("closed exporters should not publish spans")
		case <-time.After(50 * time.Millisecond):
		}

	})
}
