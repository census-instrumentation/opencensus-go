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

// Package stats defines the stats collection API and its native Go
// implementation.
package stats

import "time"

// aggregatorSlidingTime indicates that the aggregation occurs over a sliding
// window of time: i.e. last n seconds, minutes, hours...
type aggregatorSlidingTime struct {
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

// newAggregatorSlidingTime creates an aggregatorSlidingTime.
func newAggregatorSlidingTime(now time.Time, d time.Duration, subIntervalsCount int, newAggregationValue func() AggregationValue) *aggregatorSlidingTime {
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

	return &aggregatorSlidingTime{
		keptDuration:    subDuration * time.Duration(len(entries)),
		desiredDuration: subDuration * time.Duration(len(entries)-1), // this is equal to d
		subDuration:     subDuration,
		entries:         entries,
		idx:             subIntervalsCount,
	}
}

func (a *aggregatorSlidingTime) isAggregator() bool {
	return true
}

func (a *aggregatorSlidingTime) addSample(v interface{}, now time.Time) {
	a.moveToCurrentEntry(now)
	e := a.entries[a.idx]
	e.av.addSample(v)
}

func (a *aggregatorSlidingTime) retrieveCollected(now time.Time) AggregationValue {
	a.moveToCurrentEntry(now)

	e := a.entries[a.idx]
	remaining := float64(e.endTime.Sub(now)) / float64(a.subDuration)
	oldestIdx := (a.idx + 1) % len(a.entries)

	e = a.entries[oldestIdx]
	ret := e.av.multiplyByFraction(remaining)

	for j := 1; j < len(a.entries); j++ {
		oldestIdx = (oldestIdx + 1) % len(a.entries)
		e = a.entries[oldestIdx]
		ret.addToIt(e.av)
	}
	return ret
}

func (a *aggregatorSlidingTime) moveToCurrentEntry(now time.Time) {
	e := a.entries[a.idx]
	for {
		if e.endTime.After(now) {
			break
		}
		a.idx = (a.idx + 1) % len(a.entries)
		e = a.entries[a.idx]
		e.endTime = e.endTime.Add(a.keptDuration)
		e.av.clear()
	}
}

type timeSerieEntry struct {
	endTime time.Time
	av      AggregationValue
}
