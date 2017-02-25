// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package stats

import (
	"bytes"
	"fmt"
	"time"
)

// GaugeStringStats records a gauge of string sample values.
type GaugeStringStats struct {
	Value     string
	TimeStamp time.Time
}

func (gs *GaugeStringStats) stringWithIndent(tabs string) string {
	if gs == nil {
		return "nil"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", gs)
	fmt.Fprintf(&buf, "%v  Value: %v,\n", tabs, gs.Value)
	fmt.Fprintf(&buf, "%v  TimeStamp: %v,\n", tabs, gs.TimeStamp)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (gs *GaugeStringStats) String() string {
	return gs.stringWithIndent("")
}

// newGaugeAggregatorString creates a gaugeAggregator of type string. For a
// single GaugeAggregationDescriptor it is expected to be called multiple
// times. Once for each unique set of tags.
func newGaugeAggregatorString() *gaugeAggregatorString {
	return &gaugeAggregatorString{
		gs: &GaugeStringStats{},
	}
}

type gaugeAggregatorString struct {
	gs *GaugeStringStats
}

func (ga *gaugeAggregatorString) addSample(m Measurement, now time.Time) {
	ga.gs.Value = m.string()
	ga.gs.TimeStamp = now
}

func (ga *gaugeAggregatorString) retrieveCollected() *GaugeStringStats {
	return ga.gs
}
