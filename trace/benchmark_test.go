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

func doWork(ctx context.Context) context.Context {
	vx, _ := ctx.Value("introduction").(string)
	if len(vx) > 100 {
		vx = vx[:15]
		ctx = context.WithValue(ctx, "introduction", vx)
	}
	return ctx
}

func BenchmarkWithSpan_NewContextInLoop_doWork(b *testing.B) {
	traceBenchmark(b, func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			_, span := StartSpan(ctx, "/foo")
			ctx = doWork(ctx)
			span.End()
		}
	})
}

func BenchmarkWithSpan_ContextReuse_doWork(b *testing.B) {
	traceBenchmark(b, func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, span := StartSpan(ctx, "/foo")
			ctx = doWork(ctx)
			span.End()
		}
	})
}

var sink context.Context

func BenchmarkNoTracing_ContextReuse_doWork(b *testing.B) {
	traceBenchmark(b, func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			sink = doWork(ctx)
		}
		if sink == nil {
			b.Fatalf("Benchmark didn't run!")
		}
	})
}

func BenchmarkNoTracing_NewContextInLoop_doWork(b *testing.B) {
	traceBenchmark(b, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			sink = doWork(ctx)
		}
		if sink == nil {
			b.Fatalf("Benchmark didn't run!")
		}
	})
}

func BenchmarkStartEndSpan(b *testing.B) {
	traceBenchmark(b, func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, span := StartSpan(ctx, "/foo")
			span.End()
		}
	})
}

func BenchmarkSpanWithAnnotations_3(b *testing.B) {
	traceBenchmark(b, func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := StartSpan(ctx, "/foo")
			span.SetAttributes(
				BoolAttribute("key1", false),
				StringAttribute("key2", "hello"),
				Int64Attribute("key3", 123),
			)
			span.End()
		}
	})
}

func BenchmarkSpanWithAnnotations_6(b *testing.B) {
	traceBenchmark(b, func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, span := StartSpan(ctx, "/foo")
			span.SetAttributes(
				BoolAttribute("key1", false),
				BoolAttribute("key2", true),
				StringAttribute("key3", "hello"),
				StringAttribute("key4", "hello"),
				Int64Attribute("key5", 123),
				Int64Attribute("key6", 456),
			)
			span.End()
		}
	})
}

func BenchmarkTraceID_DotString(b *testing.B) {
	traceBenchmark(b, func(b *testing.B) {
		t := TraceID{0x0D, 0x0E, 0x0A, 0x0D, 0x0B, 0x0E, 0x0E, 0x0F, 0x0F, 0x0E, 0x0E, 0x0B, 0x0D, 0x0A, 0x0E, 0x0D}
		want := "0d0e0a0d0b0e0e0f0f0e0e0b0d0a0e0d"
		for i := 0; i < b.N; i++ {
			if got := t.String(); got != want {
				b.Fatalf("got = %q want = %q", got, want)
			}
		}
	})
}

func BenchmarkSpanID_DotString(b *testing.B) {
	traceBenchmark(b, func(b *testing.B) {
		s := SpanID{0x0D, 0x0E, 0x0A, 0x0D, 0x0B, 0x0E, 0x0E, 0x0F}
		want := "0d0e0a0d0b0e0e0f"
		for i := 0; i < b.N; i++ {
			if got := s.String(); got != want {
				b.Fatalf("got = %q want = %q", got, want)
			}
		}
	})
}

func traceBenchmark(b *testing.B, fn func(*testing.B)) {
	b.Run("AlwaysSample", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		SetDefaultSampler(AlwaysSample())
		fn(b)
	})
	b.Run("NeverSample", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		SetDefaultSampler(NeverSample())
		fn(b)
	})
}
