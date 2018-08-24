package stats

import (
	"context"
	"testing"
)

type NoOp struct {
	count int64
}

func (_ NoOp) ExportStats(data Data) {
}

func BenchmarkRecordWithExporter(b *testing.B) {
	e := NoOp{}
	RegisterExporter(e)
	defer UnregisterExporter(e)

	measure := Int64("name", "description", "kg")
	m := measure.M(123)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		Record(ctx, m)
	}
}
