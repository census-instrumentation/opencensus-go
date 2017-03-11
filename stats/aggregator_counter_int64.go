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

// CounterInt64Stats records a counter of int64 sample values.
type CounterInt64Stats struct {
	Value     int64
	TimeStamp time.Time
}

func (cs *CounterInt64Stats) stringWithIndent(tabs string) string {
	if cs == nil {
		return "nil"
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", cs)
	fmt.Fprintf(&buf, "%v  Value: %v,\n", tabs, cs.Value)
	fmt.Fprintf(&buf, "%v  TimeStamp: %v,\n", tabs, cs.TimeStamp)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (cs *CounterInt64Stats) String() string {
	return cs.stringWithIndent("")
}

// newCounterAggregatorInt64 creates a counterAggregator of type int64. For a
// single CounterAggregationDescriptor it is expected to be called multiple
// times. Once for each unique set of tags.
func newCounterAggregatorInt64() *counterAggregatorInt64 {
	return &counterAggregatorInt64{
		cs: &CounterInt64Stats{},
	}
}

type counterAggregatorInt64 struct {
	cs *CounterInt64Stats
}

func (ca *counterAggregatorInt64) addSample(m Measurement, now time.Time) {
	ca.cs.Value += m.(*measurementInt64).v
	ca.cs.TimeStamp = now
}

func (ca *counterAggregatorInt64) retrieveCollected() *CounterInt64Stats {
	return ca.cs
}
