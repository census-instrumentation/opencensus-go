package api

import (
	"time"
)

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
