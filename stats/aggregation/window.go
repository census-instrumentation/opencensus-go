// Copyright 2018, OpenCensus Authors
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

package aggregation

import (
	"time"
)

// WindowAggregator represents the interface for the aggregators for the various windows.
type WindowAggregator interface {
	AddSample(v interface{}, now time.Time)
	Collect(start time.Time) Data
}

func NewCumulativeAggregator(now time.Time, fn func() Data) WindowAggregator {
	return newAggregatorCumulative(now, fn)
}

func NewIntervalAggregator(duration time.Duration, subintervals int, start time.Time, fn func() Data) WindowAggregator {
	return newAggregatorInterval(start, duration, subintervals, fn)
}

// aggregatorCumulative indicates that the aggregation occurs over all samples
// seen since the view collection started.
type aggregatorCumulative struct {
	data Data
}

// newAggregatorCumulative creates an aggregatorCumulative.
func newAggregatorCumulative(now time.Time, newAggregationValue func() Data) *aggregatorCumulative {
	return &aggregatorCumulative{
		data: newAggregationValue(),
	}
}

func (a *aggregatorCumulative) AddSample(v interface{}, now time.Time) {
	// TODO(jbd): Add sample request to a buffer and process the buffer
	// at every 100 ms.
	a.data.AddSample(v)
}

func (a *aggregatorCumulative) Collect(start time.Time) Data {
	return a.data
}

// aggregatorInterval indicates that the aggregation occurs over a
// window of time.
type aggregatorInterval struct {
	// keptDuration is the full duration that needs to be kept in memory in
	// order to retrieve the aggregated data whenever it is requested. Its size
	// is subDuration*len(entries+1). The actual desiredDuration interval is
	// slightly shorter: subDuration*len(entries). The extra subDuration is
	// needed to compute an approximation of the collected stats over the last
	// desiredDuration without storing every instance with its timestamp.
	keptDuration    time.Duration
	desiredDuration time.Duration
	subDuration     time.Duration
	entries         []*timeSerieEntry
	idx             int
}

// newAggregatorInterval creates an aggregatorSlidingTime.
func newAggregatorInterval(now time.Time, d time.Duration, subIntervalsCount int, newAggregationValue func() Data) *aggregatorInterval {
	subDuration := d / time.Duration(subIntervalsCount)
	start := now.Add(-subDuration * time.Duration(subIntervalsCount))
	var entries []*timeSerieEntry
	// Keeps track of subIntervalsCount+1 entries in order to approximate the
	// collected stats without storing every instance with its timestamp.
	for i := 0; i <= subIntervalsCount; i++ {
		entries = append(entries, &timeSerieEntry{
			endTime: start.Add(subDuration),
			av:      newAggregationValue(),
		})
		start = start.Add(subDuration)
	}

	return &aggregatorInterval{
		keptDuration:    subDuration * time.Duration(len(entries)),
		desiredDuration: subDuration * time.Duration(len(entries)-1), // this is equal to d
		subDuration:     subDuration,
		entries:         entries,
		idx:             subIntervalsCount,
	}
}

func (a *aggregatorInterval) AddSample(v interface{}, now time.Time) {
	// TODO(jbd): Add sample request to a buffer and process the buffer
	// at every 100 ms.
	a.moveToCurrentEntry(now)
	e := a.entries[a.idx]
	e.av.AddSample(v)
}

func (a *aggregatorInterval) Collect(start time.Time) Data {
	a.moveToCurrentEntry(start)

	e := a.entries[a.idx]
	remaining := float64(e.endTime.Sub(start)) / float64(a.subDuration)
	oldestIdx := (a.idx + 1) % len(a.entries)

	e = a.entries[oldestIdx]
	ret := e.av.MultiplyByFraction(remaining)

	for j := 1; j < len(a.entries); j++ {
		oldestIdx = (oldestIdx + 1) % len(a.entries)
		e = a.entries[oldestIdx]
		ret.AddData(e.av)
	}
	return ret
}

func (a *aggregatorInterval) moveToCurrentEntry(now time.Time) {
	e := a.entries[a.idx]
	for {
		if e.endTime.After(now) {
			break
		}
		a.idx = (a.idx + 1) % len(a.entries)
		e = a.entries[a.idx]
		e.endTime = e.endTime.Add(a.keptDuration)
		e.av.Clear()
	}
}

type timeSerieEntry struct {
	endTime time.Time
	av      Data
}
