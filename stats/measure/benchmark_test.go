package measure_test

import (
	"context"
	"log"
	"testing"

	"go.opencensus.io/stats/measure"
	_ "go.opencensus.io/stats/view"
)

var m = makeMeasure()

func BenchmarkRecord(b *testing.B) {
	var ctx = context.Background()
	for i := 0; i < b.N; i++ {
		measure.Record(ctx, m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1))
	}
}

func makeMeasure() *measure.Int64 {
	m, err := measure.NewInt64("m", "test measure", "")
	if err != nil {
		log.Fatal(err)
	}
	return m
}
