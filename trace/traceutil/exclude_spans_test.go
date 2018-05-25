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

package traceutil

import (
	"testing"

	"go.opencensus.io/trace"
)

type testExporter struct {
	exported []*trace.SpanData
	flushed  bool
}

func (e *testExporter) ExportSpan(spanData *trace.SpanData) {
	e.exported = append(e.exported, spanData)
}

func (e *testExporter) Flush() {
	e.flushed = true
}

func TestExcludeSpans_AllExcluded(t *testing.T) {
	te := &testExporter{}
	e := NewExcludeSpansExporter(te, nil)
	e.ExportSpan(&trace.SpanData{})
	if got, want := len(te.exported), 0; got != want {
		t.Fatalf("len(te.exported) = %d; want %d", got, want)
	}
}

func TestExcludeSpans(t *testing.T) {
	te := &testExporter{}
	e := NewExcludeSpansExporter(te, []SpanMatcher{
		{NamePrefix: "/com.example.HelloWorld", SpanKind: trace.SpanKindClient},
	})

	e.ExportSpan(&trace.SpanData{
		Name:     "/com.example.HelloWorld.SayHello",
		SpanKind: trace.SpanKindClient,
	})
	if got, want := len(te.exported), 0; got != want {
		t.Fatalf("len(te.exported) = %d; want %d", got, want)
	}

	e.ExportSpan(&trace.SpanData{
		Name:     "/com.example.NoMatch.SayHello",
		SpanKind: trace.SpanKindClient,
	})
	if got, want := len(te.exported), 1; got != want {
		t.Fatalf("len(te.exported) = %d; want %d", got, want)
	}

	e.ExportSpan(&trace.SpanData{
		Name:     "/com.example.HelloWorld.SayHello",
		SpanKind: trace.SpanKindServer,
	})
	if got, want := len(te.exported), 2; got != want {
		t.Fatalf("len(te.exported) = %d; want %d", got, want)
	}
}
