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

package trace

import (
	"context"
	"testing"
)

func BenchmarkStartEndSpan_noExporters_neverSample(b *testing.B) {
	ApplyConfig(Config{DefaultSampler: NeverSample()})
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.End()
	}
}

func BenchmarkStartEndSpan_noExporters_alwaysSample(b *testing.B) {
	ApplyConfig(Config{DefaultSampler: AlwaysSample()})
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.End()
	}
}

type anExporter int

func (ae *anExporter) ExportSpan(s *SpanData) {}

var _ Exporter = (*anExporter)(nil)

func BenchmarkStartEndSpan_withExporters_neverSample(b *testing.B) {
	ae := new(anExporter)
	RegisterExporter(ae)
	defer UnregisterExporter(ae)

	ApplyConfig(Config{DefaultSampler: NeverSample()})
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.End()
	}
}

func BenchmarkStartEndSpan_withExporters_alwaysSample(b *testing.B) {
	ae := new(anExporter)
	RegisterExporter(ae)
	defer UnregisterExporter(ae)

	ApplyConfig(Config{DefaultSampler: AlwaysSample()})
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.End()
	}
}

func BenchmarkSpanWithAnnotations_3_noExporters_neverSample(b *testing.B) {
	ctx := context.Background()
	ApplyConfig(Config{DefaultSampler: NeverSample()})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.AddAttributes(
			BoolAttribute("key2", true),
			StringAttribute("key4", "hello"),
			Int64Attribute("key5", 123),
		)
		span.End()
	}
}

func BenchmarkSpanWithAnnotations_3_noExporters_alwaysSample(b *testing.B) {
	ctx := context.Background()
	ApplyConfig(Config{DefaultSampler: AlwaysSample()})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.AddAttributes(
			BoolAttribute("key2", true),
			StringAttribute("key4", "hello"),
			Int64Attribute("key5", 123),
		)
		span.End()
	}
}

func BenchmarkSpanWithAnnotations_3_withExporters_neverSample(b *testing.B) {
	ae := new(anExporter)
	RegisterExporter(ae)
	defer UnregisterExporter(ae)

	ctx := context.Background()
	ApplyConfig(Config{DefaultSampler: NeverSample()})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.AddAttributes(
			BoolAttribute("key2", true),
			StringAttribute("key4", "hello"),
			Int64Attribute("key5", 123),
		)
		span.End()
	}
}

func BenchmarkSpanWithAnnotations_3_withExporters_alwaysSample(b *testing.B) {
	ae := new(anExporter)
	RegisterExporter(ae)
	defer UnregisterExporter(ae)

	ctx := context.Background()
	ApplyConfig(Config{DefaultSampler: AlwaysSample()})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.AddAttributes(
			BoolAttribute("key2", true),
			StringAttribute("key4", "hello"),
			Int64Attribute("key5", 123),
		)
		span.End()
	}
}

func BenchmarkSpanWithAnnotations_6_noExporters_neverSample(b *testing.B) {
	ctx := context.Background()
	ApplyConfig(Config{DefaultSampler: NeverSample()})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.AddAttributes(
			BoolAttribute("key1", false),
			BoolAttribute("key2", true),
			StringAttribute("key3", "hello"),
			StringAttribute("key4", "hello"),
			Int64Attribute("key5", 123),
			Int64Attribute("key6", 456),
		)
		span.End()
	}
}

func BenchmarkSpanWithAnnotations_6_noExporters_alwaysSample(b *testing.B) {
	ctx := context.Background()
	ApplyConfig(Config{DefaultSampler: AlwaysSample()})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.AddAttributes(
			BoolAttribute("key1", false),
			BoolAttribute("key2", true),
			StringAttribute("key3", "hello"),
			StringAttribute("key4", "hello"),
			Int64Attribute("key5", 123),
			Int64Attribute("key6", 456),
		)
		span.End()
	}
}

func BenchmarkSpanWithAnnotations_6_withExporters_neverSample(b *testing.B) {
	ae := new(anExporter)
	RegisterExporter(ae)
	defer UnregisterExporter(ae)

	ctx := context.Background()
	ApplyConfig(Config{DefaultSampler: NeverSample()})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.AddAttributes(
			BoolAttribute("key1", false),
			BoolAttribute("key2", true),
			StringAttribute("key3", "hello"),
			StringAttribute("key4", "hello"),
			Int64Attribute("key5", 123),
			Int64Attribute("key6", 456),
		)
		span.End()
	}
}

func BenchmarkSpanWithAnnotations_6_withExporters_alwaysSample(b *testing.B) {
	ae := new(anExporter)
	RegisterExporter(ae)
	defer UnregisterExporter(ae)

	ctx := context.Background()
	ApplyConfig(Config{DefaultSampler: AlwaysSample()})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, span := StartSpan(ctx, "/foo")
		span.AddAttributes(
			BoolAttribute("key1", false),
			BoolAttribute("key2", true),
			StringAttribute("key3", "hello"),
			StringAttribute("key4", "hello"),
			Int64Attribute("key5", 123),
			Int64Attribute("key6", 456),
		)
		span.End()
	}
}

func BenchmarkSpanID_DotString_noExporters_neverSample(b *testing.B) {
	ApplyConfig(Config{DefaultSampler: NeverSample()})
	s := SpanID{0x0D, 0x0E, 0x0A, 0x0D, 0x0B, 0x0E, 0x0E, 0x0F}
	want := "0d0e0a0d0b0e0e0f"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := s.String(); got != want {
			b.Fatalf("got = %q want = %q", got, want)
		}
	}
}

func BenchmarkSpanID_DotString_withExporters_neverSample(b *testing.B) {
	ae := new(anExporter)
	RegisterExporter(ae)
	defer UnregisterExporter(ae)

	ApplyConfig(Config{DefaultSampler: NeverSample()})
	s := SpanID{0x0D, 0x0E, 0x0A, 0x0D, 0x0B, 0x0E, 0x0E, 0x0F}
	want := "0d0e0a0d0b0e0e0f"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := s.String(); got != want {
			b.Fatalf("got = %q want = %q", got, want)
		}
	}
}

func BenchmarkSpanID_DotString_noExporters_alwaysSample(b *testing.B) {
	ApplyConfig(Config{DefaultSampler: AlwaysSample()})
	s := SpanID{0x0D, 0x0E, 0x0A, 0x0D, 0x0B, 0x0E, 0x0E, 0x0F}
	want := "0d0e0a0d0b0e0e0f"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := s.String(); got != want {
			b.Fatalf("got = %q want = %q", got, want)
		}
	}
}

func BenchmarkSpanID_DotString_withExporters_alwaysSample(b *testing.B) {
	ae := new(anExporter)
	RegisterExporter(ae)
	defer UnregisterExporter(ae)

	ApplyConfig(Config{DefaultSampler: AlwaysSample()})
	s := SpanID{0x0D, 0x0E, 0x0A, 0x0D, 0x0B, 0x0E, 0x0E, 0x0F}
	want := "0d0e0a0d0b0e0e0f"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := s.String(); got != want {
			b.Fatalf("got = %q want = %q", got, want)
		}
	}
}
