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

// GaugeBoolStats records a gauge of bool sample values.
type GaugeBoolStats struct {
	Value     bool
	TimeStamp time.Time
}

func (gs *GaugeBoolStats) stringWithIndent(tabs string) string {
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

func (gs *GaugeBoolStats) String() string {
	return gs.stringWithIndent("")
}

// newGaugeAggregatorBool creates a gaugeAggregator of type bool. For a single
// GaugeAggregationDescriptor it is expected to be called multiple times. Once
// for each unique set of tags.
func newGaugeAggregatorBool() *gaugeAggregatorBool {
	return &gaugeAggregatorBool{
		gs: &GaugeBoolStats{},
	}
}

type gaugeAggregatorBool struct {
	gs *GaugeBoolStats
}

func (ga *gaugeAggregatorBool) addSample(m Measurement, now time.Time) {
	ga.gs.Value = m.(*measurementBool).v
	ga.gs.TimeStamp = now
}

func (ga *gaugeAggregatorBool) retrieveCollected() *GaugeBoolStats {
	return ga.gs
}
