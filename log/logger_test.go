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

package log_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"go.opencensus.io/log"
	"go.opencensus.io/tag"
)

type Capturer struct {
	Records []log.Data
}

func (c *Capturer) ExportLog(d log.Data) {
	c.Records = append(c.Records, d)
}

func TestLoggerInfo(t *testing.T) {
	exporter := &Capturer{}
	logger := &log.Logger{}
	logger.RegisterExporter(exporter)

	// When
	message := "hello world"
	logger.Info(context.Background(), message)

	// Then
	if got := len(exporter.Records); got != 1 {
		t.Fatalf("got %v; want 1 record", got)
	}

	data := exporter.Records[0]
	if data.Message != message {
		t.Errorf("got %v; want %v", message, data.Message)
	}
	if data.Timestamp.IsZero() {
		t.Error("expected Timestamp to be set")
	}
	if want := log.InfoLevel; data.LogLevel != want {
		t.Errorf("got %v; want %v", data.LogLevel, want)
	}
}

func TestLoggerDebug(t *testing.T) {
	exporter := &Capturer{}
	logger := &log.Logger{}
	logger.RegisterExporter(exporter)

	// When
	message := "hello world"
	logger.Debug(context.Background(), message)

	// Then
	if got := len(exporter.Records); got != 1 {
		t.Fatalf("got %v; want 1 record", got)
	}

	data := exporter.Records[0]
	if want := log.DebugLevel; data.LogLevel != want {
		t.Errorf("got %v; want %v", data.LogLevel, want)
	}
}

func TestLoggerHandlesMultipleExporters(t *testing.T) {
	e1 := &Capturer{}
	e2 := &Capturer{}
	logger := &log.Logger{}
	logger.RegisterExporter(e1)
	logger.RegisterExporter(e2)

	// When
	logger.Info(context.Background(), "hello world")

	// Then
	if got := len(e1.Records); got != 1 {
		t.Fatalf("got %v; want 1 record", got)
	}
	if got := len(e2.Records); got != 1 {
		t.Fatalf("got %v; want 1 record", got)
	}
}

func TestLoggerCustomTimeFunc(t *testing.T) {
	now := time.Date(2018, time.July, 1, 2, 3, 4, 0, time.UTC)
	e := &Capturer{}
	logger := &log.Logger{}
	logger.RegisterExporter(e)
	logger.ApplyConfig(log.Config{
		TimeFunc: func() time.Time { return now },
	})

	logger.Info(context.Background(), "hello world")

	// Then
	if got := len(e.Records); got != 1 {
		t.Fatalf("got %v; want 1 record", got)
	}

	if data := e.Records[0]; data.Timestamp != now {
		t.Errorf("got %v; want %v", data.Timestamp.Format(time.RFC3339), now.Format(time.RFC3339))
	}
}

func TestLoggerGlobalFields(t *testing.T) {
	e := &Capturer{}
	logger := &log.Logger{}
	logger.RegisterExporter(e)
	global := log.String("global", "field")
	logger.ApplyConfig(log.Config{
		Fields: []log.Field{
			global,
		},
	})

	logger.Info(context.Background(), "hello world")

	// Then
	if got := len(e.Records); got != 1 {
		t.Fatalf("got %v; want 1 record", got)
	}

	data := e.Records[0]
	if want := []log.Field{global}; !reflect.DeepEqual(data.Fields, want) {
		t.Errorf("got %v; want %v", data.Fields, want)
	}
}

func TestLoggerTags(t *testing.T) {
	key, _ := tag.NewKey("uid")

	e := &Capturer{}
	logger := &log.Logger{}
	logger.RegisterExporter(e)
	logger.ApplyConfig(log.Config{
		Tags: []tag.Key{
			key,
		},
	})

	ctx, _ := tag.New(context.Background(), tag.Insert(key, "abc"))
	logger.Info(ctx, "hello world")

	// Then
	if got := len(e.Records); got != 1 {
		t.Fatalf("got %v; want 1 record", got)
	}

	data := e.Records[0]
	if want := map[string]string{"uid": "abc"}; !reflect.DeepEqual(data.Tags, want) {
		t.Errorf("got %v; want %v", data.Tags, want)
	}
}

func TestLoggerUnregisterExporter(t *testing.T) {
	exporter := &Capturer{}
	logger := &log.Logger{}
	logger.RegisterExporter(exporter)
	logger.UnregisterExporter(exporter)

	// When
	logger.Info(context.Background(), "hello world")

	// Then
	if len(exporter.Records) != 0 {
		t.Fatalf("expected 0 records to be captured when no exporter registered")
	}
}
