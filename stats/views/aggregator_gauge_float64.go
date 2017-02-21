package views

import (
	"bytes"
	"fmt"
	"time"
)

// GaugeFloat64Stats records a gauge of float64 sample values.
type GaugeFloat64Stats struct {
	Value     float64
	TimeStamp time.Time
}

func (gs *GaugeFloat64Stats) String() string {
	if gs == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  GaugeFloat64Stats{\n")
	fmt.Fprintf(&buf, "    Value: %v,\n", gs.Value)
	fmt.Fprintf(&buf, "    TimeStamp: %v,\n", gs.TimeStamp)
	buf.WriteString("  }")
	return buf.String()
}
