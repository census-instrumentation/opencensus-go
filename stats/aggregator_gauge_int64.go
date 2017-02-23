package stats

import (
	"bytes"
	"fmt"
	"time"
)

// GaugeInt64Stats records a gauge of int64 sample values.
type GaugeInt64Stats struct {
	Value     int64
	TimeStamp time.Time
}

func (gs *GaugeInt64Stats) String() string {
	if gs == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  GaugeInt64Stats{\n")
	fmt.Fprintf(&buf, "    Value: %v,\n", gs.Value)
	fmt.Fprintf(&buf, "    TimeStamp: %v,\n", gs.TimeStamp)
	buf.WriteString("  }")
	return buf.String()
}

// newGaugeAggregatorInt64 creates a gaugeAggregator of int64. For a single
// GaugeAggregationDescriptor it is expected to be called multiple
// times. Once for each unique set of tags.
func newGaugeAggregatorInt64() *gaugeAggregatorInt64 {
	return &gaugeAggregatorInt64{
		gs: &GaugeInt64Stats{},
	}
}

type gaugeAggregatorInt64 struct {
	gs *GaugeInt64Stats
}

func (ga *gaugeAggregatorInt64) addSample(m Measurement, now time.Time) {
	ga.gs.Value = m.int64()
	ga.gs.TimeStamp = now
}

func (ga *gaugeAggregatorInt64) retrieveCollected() *GaugeInt64Stats {
	return ga.gs
}
