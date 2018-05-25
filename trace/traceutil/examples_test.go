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

package traceutil_test

import (
	"fmt"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/traceutil"
	"golang.org/x/net/context"
)

type testExporter struct {
	buf string
}

func (e *testExporter) ExportSpan(spanData *trace.SpanData) {
	e.buf += spanData.Name + "\n"
}

func (e *testExporter) Flush() {
	fmt.Print(e.buf)
}

func ExampleNewExcludeSpansExporter() {
	ctx := context.Background()
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	filtered := traceutil.NewExcludeSpansExporter(
		new(testExporter),
		[]traceutil.SpanMatcher{
			{NamePrefix: "io.opencensus.Excluded/", SpanKind: trace.SpanKindClient},
		})
	trace.RegisterExporter(filtered)

	// Sampled and exported span:
	_, span := trace.StartSpan(ctx, "io.opencensus.Included/SomeMethod", trace.WithSpanKind(trace.SpanKindClient))
	span.End()

	// Sampled, but not exported span:
	_, span2 := trace.StartSpan(ctx, "io.opencensus.Excluded/SomeMethod", trace.WithSpanKind(trace.SpanKindClient))
	span2.End()

	filtered.Flush()
	// Output: io.opencensus.Included/SomeMethod
}
