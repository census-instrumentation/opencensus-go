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

// Window represents the interval/samples count over which the aggregation
// occurs.
type Window interface {
	isWindow() bool
	newAggregator(now time.Time, aggregationValueConstructor func() AggregationValue) aggregator
}

// WindowCumulative indicates that the aggregation occurs over the lifetime of
// the view.
type WindowCumulative struct{}

// NewWindowCumulative creates a new aggregation window of type cumulative.
func NewWindowCumulative() *WindowCumulative {
	return &WindowCumulative{}
}

func (w *WindowCumulative) isWindow() bool { return true }

func (w *WindowCumulative) newAggregator(now time.Time, aggregationValueConstructor func() AggregationValue) aggregator {
	return newAggregatorCumulative(now, aggregationValueConstructor)
}

// WindowSlidingTime indicates that the aggregation occurs over a sliding
// window of time: i.e. last n seconds, minutes, hours...
type WindowSlidingTime struct {
	duration     time.Duration
	subIntervals int
}

// NewWindowSlidingTime creates a new aggregation window of type sliding time
// a.k.a time interval.
func NewWindowSlidingTime(duration time.Duration, subIntervals int) *WindowSlidingTime {
	return &WindowSlidingTime{
		duration:     duration,
		subIntervals: subIntervals,
	}
}

func (w *WindowSlidingTime) isWindow() bool { return true }

func (w *WindowSlidingTime) newAggregator(now time.Time, aggregationValueConstructor func() AggregationValue) aggregator {
	return newAggregatorSlidingTime(now, w.duration, w.subIntervals, aggregationValueConstructor)
}

// WindowSlidingCount indicates that the aggregation occurs over a sliding
// number of samples.
type WindowSlidingCount struct {
	n       uint64
	subSets int
}

// NewWindowSlidingCount creates a new aggregation window of type sliding count
// a.k.a last n samples.
func NewWindowSlidingCount(count uint64, subSets int) *WindowSlidingCount {
	return &WindowSlidingCount{
		n:       count,
		subSets: subSets,
	}
}

func (w *WindowSlidingCount) isWindow() bool { return true }

func (w *WindowSlidingCount) newAggregator(now time.Time, aggregationValueConstructor func() AggregationValue) aggregator {
	return newAggregatorSlidingCount(now, w.n, w.subSets, aggregationValueConstructor)
}
