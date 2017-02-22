package main

import (
	"context"
	"fmt"

	"github.com/google/instrumentation-go/stats/views"
	"github.com/google/instrumentation-go/tagging"
)

func main() {
	mu := &views.MeasurementUnit{
		Power10: 6,
		Numerators: []views.BasicUnit{
			views.BytesUnit,
		},
	}

	// Create measure description of type float
	x := views.NewMeasureDescFloat64("DiskRead", "Read MBs", mu)

	// Creates few measurements
	m10 := x.CreateMeasurement(10)
	m100 := x.CreateMeasurement(100)
	m1 := x.CreateMeasurement(1)

	mgr := tagging.DefaultKeyManager()
	k1 := mgr.CreateKeyInt64("key1")
	k2 := mgr.CreateKeyInt64("key2")
	k3 := mgr.CreateKeyString("key3")

	mut1 := k1.CreateMutation(10, tagging.BehaviorAdd)
	mut2 := k2.CreateMutation(20, tagging.BehaviorAdd)
	mut3 := k3.CreateMutation("value3", tagging.BehaviorAdd)

	ts := make(tagging.Tags)
	ts.ApplyMutations(mut1, mut2, mut3)

	views.RecordMeasurements(context.Background(), m10, m100, m1)
	y := x.Meta()
	fmt.Printf("Hello, world:\n%v\n%v\n", m1, y)
}
