package main

import (
	"fmt"

	"github.com/google/instrumentation-go/stats/api"
)

func main() {
	x := api.NewMeasureDescFloat64("test", "description", nil)
	m := x.CreateMeasurement(24)
	y := x.Meta()
	fmt.Printf("Hello, world.\n")
}
