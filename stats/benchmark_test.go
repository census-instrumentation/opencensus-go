package stats

import (
	"testing"
	"log"
	"context"
)


func BenchmarkRecord(b *testing.B) {
	restart()
	var m = makeMeasure()
	var ctx = context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Record(ctx, m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1))
	}
}

func makeMeasure() *MeasureInt64 {
	m, err := NewMeasureInt64("m", "test measure", "")
	if err != nil {
		log.Fatal(err)
	}
	return m
}
