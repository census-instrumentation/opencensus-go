package main

import (
	"fmt"

	"github.com/google/instrumentation-go/stats/views"
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

	
	y := x.Meta()
	fmt.Printf("Hello, world:\n%v\n%v\n", m, y)
}
