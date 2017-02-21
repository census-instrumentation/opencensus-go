package views

import (
	"bytes"
	"fmt"
	"time"
)

// GaugeStringStats records a gauge of string sample values.
type GaugeStringStats struct {
	Value     int64
	TimeStamp time.Time
}

func (gs *GaugeStringStats) String() string {
	if gs == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  GaugeStringStats{\n")
	fmt.Fprintf(&buf, "    Value: %v,\n", gs.Value)
	fmt.Fprintf(&buf, "    TimeStamp: %v,\n", gs.TimeStamp)
	buf.WriteString("  }")
	return buf.String()
}
