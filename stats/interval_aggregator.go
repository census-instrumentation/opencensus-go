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

import (
	"bytes"
	"fmt"
	"time"
)

type intervalsAggregator struct {
	buffers []*timeSeriesBuffer
}

// newIntervalsAggregator creates an intervalsAggregator. For a single
// IntervalAggregationDescriptor it is expected to be called multiple times.
// Once for each unique set of tags.
func newIntervalsAggregator(now time.Time, intervals []time.Duration, subIntervalsCount int) *intervalsAggregator {
	var buffers []*timeSeriesBuffer
	for _, d := range intervals {
		buffers = append(buffers, newTimeSeriesBuffer(now, d, subIntervalsCount))
	}
	return &intervalsAggregator{
		buffers: buffers,
	}
}

func (ia *intervalsAggregator) addSample(v float64, now time.Time) {
	for _, b := range ia.buffers {
		b.moveToCurrentEntry(now)
		e := b.entries[b.idx]
		e.count++
		e.sum += v
	}
}

func (ia *intervalsAggregator) retrieveCollected(now time.Time) []*IntervalStats {
	var ret []*IntervalStats
	for _, b := range ia.buffers {
		is := b.retrieveCollected(now)
		ret = append(ret, is)
	}
	return ret
}

type timeSeriesBuffer struct {
	// keptDuration is the full duration that needs to be kept in memory in
	// order to retrieve the intervalStats whenever it is requested. Its size
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

func newTimeSeriesBuffer(now time.Time, d time.Duration, subIntervalsCount int) *timeSeriesBuffer {
	subDuration := d / time.Duration(subIntervalsCount)
	start := now.Add(-subDuration * time.Duration(subIntervalsCount))
	var entries []*timeSerieEntry
	// Keeps track of subIntervalsCount+1 entries in order to approximate the
	// collected stats without storing every instance with its timestamp.
	for i := 0; i <= subIntervalsCount; i++ {
		entries = append(entries, &timeSerieEntry{
			endTime: start.Add(subDuration),
		})
		start = start.Add(subDuration)
	}

	ts := &timeSeriesBuffer{
		keptDuration:    subDuration * time.Duration(len(entries)),
		desiredDuration: subDuration * time.Duration(len(entries)-1),
		subDuration:     subDuration,
		entries:         entries,
		idx:             subIntervalsCount,
	}

	return ts
}

func (ts *timeSeriesBuffer) moveToCurrentEntry(now time.Time) {
	e := ts.entries[ts.idx]
	for {
		if e.endTime.After(now) {
			break
		}
		ts.idx = (ts.idx + 1) % len(ts.entries)
		e = ts.entries[ts.idx]
		e.endTime = e.endTime.Add(ts.keptDuration)
		e.count = 0
		e.sum = 0
	}
}

func (ts *timeSeriesBuffer) retrieveCollected(now time.Time) *IntervalStats {
	ts.moveToCurrentEntry(now)

	e := ts.entries[ts.idx]
	remaining := float64(e.endTime.Sub(now)) / float64(ts.subDuration)
	oldestIdx := (ts.idx + 1) % len(ts.entries)

	e = ts.entries[oldestIdx]
	ret := &IntervalStats{
		Duration: ts.desiredDuration,
		Count:    e.count * remaining,
		Sum:      e.sum * remaining,
	}
	for j := 1; j < len(ts.entries); j++ {
		oldestIdx = (oldestIdx + 1) % len(ts.entries)
		e = ts.entries[oldestIdx]
		ret.Count += e.count
		ret.Sum += e.sum
	}
	return ret
}

func (ts *timeSeriesBuffer) String() string {
	if ts == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("{")
	fmt.Fprintf(&buf, "idx: %v,\n", ts.idx)
	buf.WriteString("entries: {\n")
	for _, e := range ts.entries {
		buf.WriteString("  {\n")
		fmt.Fprintf(&buf, "    sum: %v,\n", e.sum)
		fmt.Fprintf(&buf, "    count: %v,\n", e.count)
		fmt.Fprintf(&buf, "    endTime: %v,\n", e.endTime)
		buf.WriteString("  },\n")
	}
	buf.WriteString("},\n")
	buf.WriteString("}")
	return buf.String()
}

type timeSerieEntry struct {
	endTime    time.Time
	count, sum float64
}
