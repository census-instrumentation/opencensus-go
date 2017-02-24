package main

import (
	"context"
	"fmt"

	"github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tagging"
)

func main() {
	// Creates/Retrieves tags keys
	mgr := tagging.DefaultKeyManager()
	k1, err := mgr.CreateKeyInt64("key1")
	if err != nil {
		panic(fmt.Sprintf("Key k1 not created %v", err))
	}
	k2, err := mgr.CreateKeyInt64("key2")
	if err != nil {
		panic(fmt.Sprintf("Key k2 not created %v", err))
	}
	k3, err := mgr.CreateKeyString("key3")
	if err != nil {
		panic(fmt.Sprintf("Key k3 not created %v", err))
	}

	// Set tags values in mutations
	mut1 := k1.CreateMutation(10, tagging.BehaviorAdd)
	mut2 := k2.CreateMutation(20, tagging.BehaviorAdd)
	mut3 := k3.CreateMutation("value3", tagging.BehaviorAdd)

	// Create context
	ctx := tagging.NewContextWithMutations(context.Background(), mut1, mut2, mut3)

	//...
	// DoStuff()
	//...

	// Create measure description of type float
	mu := &stats.MeasurementUnit{
		Power10: 6,
		Numerators: []stats.BasicUnit{
			stats.BytesUnit,
		},
	}
	mDesc1 := stats.NewMeasureDescFloat64("DiskRead", "Read MBs", mu)
	mDesc2 := stats.NewMeasureDescFloat64("DiskWrites", "Write MBs", mu)

	// Creates few measurements
	m1_1 := mDesc1.CreateMeasurement(50)
	m1_2 := mDesc1.CreateMeasurement(100)
	m1_3 := mDesc1.CreateMeasurement(200)

	m2_1 := mDesc2.CreateMeasurement(5)
	m2_2 := mDesc2.CreateMeasurement(10)
	m2_3 := mDesc2.CreateMeasurement(20)

	// Record usage
	stats.RecordMeasurements(ctx, m1_1, m1_2, m1_3, m2_1, m2_2, m2_3)

	// Retrieve Views[]
	views := stats.RetrieveView("some view")

	// Console out
	for _, v := range views {
		fmt.Printf("View: %v", v)
	}

	/*y := mDesc1.Meta()
	fmt.Printf("Hello, world:\n%v\n%v\n", m1_1, y)
	*/
}
