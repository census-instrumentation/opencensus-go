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
	"testing"
	"time"
)

func TestIntervalAggregator(t *testing.T) {
	type record struct {
		t time.Time
		v float64
	}

	now := time.Now()
	tds := []struct {
		intervalDuration   time.Duration
		subIntervalsCount  int
		records            []record
		retrieveTime       time.Time
		wantCount, wantSum float64
	}{
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
			},
			now.Add(5 * time.Second),
			1,
			2,
		},
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
			},
			now.Add(6 * time.Second),
			0,
			0,
		},
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
			},
			now.Add(5500 * time.Millisecond),
			0.5,
			1,
		},
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
			},
			now.Add(5900 * time.Millisecond),
			0.1,
			0.2,
		},
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
				{now.Add(time.Second), 2},
				{now.Add(2 * time.Second), 2},
				{now.Add(3 * time.Second), 2},
				{now.Add(4 * time.Second), 2},
				{now.Add(5 * time.Second), 2},
			},
			now.Add(5 * time.Second),
			6,
			12,
		},
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
				{now.Add(time.Second), 2},
				{now.Add(2 * time.Second), 2},
				{now.Add(3 * time.Second), 2},
				{now.Add(4 * time.Second), 2},
				{now.Add(5 * time.Second), 2},
			},
			now.Add(6 * time.Second),
			5,
			10,
		},
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
				{now.Add(5 * time.Second), 2},
			},
			now.Add(6 * time.Second),
			1,
			2,
		},
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
				{now.Add(5 * time.Second), 2},
				{now.Add(6 * time.Second), 2},
			},
			now.Add(8 * time.Second),
			2,
			4,
		},
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
				{now.Add(5 * time.Second), 2},
				{now.Add(6 * time.Second), 2},
			},
			now.Add(11 * time.Second),
			1,
			2,
		},
		{
			5 * time.Second,
			5,
			[]record{
				{now, 2},
				{now.Add(5 * time.Second), 2},
				{now.Add(6 * time.Second), 2},
			},
			now.Add(10500 * time.Millisecond),
			1.5,
			3,
		},
	}

	for _, td := range tds {
		ia := newIntervalsAggregator(now, []time.Duration{td.intervalDuration}, td.subIntervalsCount)
		for _, r := range td.records {
			ia.addSample(r.v, r.t)
		}

		is := ia.retrieveCollected(td.retrieveTime)[0]

		if got, want := is.Duration, td.intervalDuration; got != want {
			t.Errorf("got duration %v (test case: %v) , want %v", got, td, want)
		}

		if got, want := is.Count, td.wantCount; got != want {
			t.Errorf("got count %v (test case: %v) , want %v", got, td, want)
		}

		if got, want := is.Sum, td.wantSum; got != want {
			t.Errorf("got sum %v (test case: %v) , want %v", got, td, want)
		}
	}
}
