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

	"go.opencensus.io/log"
)

func TestGlobalApplyConfig(t *testing.T) {
	ctx := context.Background()
	e := &Capturer{}
	global := log.String("global", "field")

	log.RegisterExporter(e)
	log.ApplyConfig(log.Config{
		Fields: []log.Field{global},
	})

	log.Info(ctx, "message")

	if got := len(e.Records); got != 1 {
		t.Fatalf("got %v, want 1", got)
	}

	data := e.Records[0]
	if want := []log.Field{global}; !reflect.DeepEqual(data.Fields, want) {
		t.Errorf("got %v, want %v", data.Fields, want)
	}
}

func TestGlobalInfo(t *testing.T) {
	ctx := context.Background()
	e := &Capturer{}
	log.RegisterExporter(e)

	log.Info(ctx, "message")

	if got := len(e.Records); got != 1 {
		t.Fatalf("got %v, want 1", got)
	}
}

func TestGlobalDebug(t *testing.T) {
	ctx := context.Background()
	e := &Capturer{}
	log.ApplyConfig(log.Config{
		LogLevel: log.DebugLevel,
	})
	log.RegisterExporter(e)

	log.Debug(ctx, "message")

	if got := len(e.Records); got != 1 {
		t.Fatalf("got %v, want 1", got)
	}
}

func TestGlobalUnregister(t *testing.T) {
	ctx := context.Background()
	e := &Capturer{}
	log.RegisterExporter(e)
	log.UnregisterExporter(e)
	log.Info(ctx, "ignored")

	if got := len(e.Records); got != 0 {
		t.Fatalf("got %v, want 0", got)
	}
}
