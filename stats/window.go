// Copyright 2017, OpenCensus Authors
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

import "time"

// Window represents a time interval or samples count over
// which the aggregation occurs.
type Window interface {
	isWindow()
	newAggregator(now time.Time, aggregationValueConstructor func() AggregationData) aggregator
}

// CumulativeWindow indicates that the aggregation occurs over the lifetime of
// the view.
type CumulativeWindow struct{}

func (w CumulativeWindow) isWindow() {}

func (w CumulativeWindow) newAggregator(now time.Time, aggregationValueConstructor func() AggregationData) aggregator {
	return newAggregatorCumulative(now, aggregationValueConstructor)
}

// SlidingTimeWindow indicates that the aggregation occurs over a sliding
// window of time: last n seconds, minutes, hours.
type SlidingTimeWindow struct {
	Duration  time.Duration
	Intervals int
}

func (w SlidingTimeWindow) isWindow() {}

func (w SlidingTimeWindow) newAggregator(now time.Time, aggregationValueConstructor func() AggregationData) aggregator {
	return newAggregatorSlidingTime(now, w.Duration, w.Intervals, aggregationValueConstructor)
}
