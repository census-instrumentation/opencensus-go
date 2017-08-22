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

import (
	"time"

	"github.com/google/working-instrumentation-go/tags"
)

// Aggregation is the generic interface for all aggregtion types.
type Aggregation interface {
	isAggregation() bool
	clearRows()
	collectedRows(keys []tags.Key) []*Row
}

type AggregationCount struct{}

func (a AggregationCount) isAggregation() bool { return true }

type AggregationDistribution struct {
	// An aggregation distribution may contain a histogram of the values in the
	// population. The bucket boundaries for that histogram are described
	// by Bounds. This defines len(Bounds)+1 buckets.
	//
	// if len(Bounds) >= 2 then the boundaries for bucket index i are:
	// [-infinity, bounds[i]) for i = 0
	// [bounds[i-1], bounds[i]) for 0 < i < len(Bounds)
	// [bounds[i-1], +infinity) for i = len(Bounds)
	//
	// if len(Bounds) == 0 then there is no histogram associated with the
	// distribution. There will be a single bucket with boundaries
	// (-infinity, +infinity).
	//
	// if len(Bounds) == 1 then there is no finite buckets, and that single
	// element is the common boundary of the overflow and underflow buckets.
	Bounds []float64
}

func (a AggregationDistribution) isAggregation() bool { return true }

//--------------------------------------------------------------------
//--------------------------------------------------------------------

type AggregationBase struct {
	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its Window.
	signatures map[string]aggregator
	w          Window
	a          Aggregation
}

func (ab *AggregationBase) addSample(s string, v interface{}, now time.Time) {
	aggregator, ok := ab.signatures[s]
	if !ok {
		// TODO: use a type switch statement to define newAggregateValue
		newAggregateValue := func() AggregateValue { return newAggregateDistribution([]float64{}) }
		//v := newAggregateDistribution(a.Bounds)

		switch w := ab.w.(type) {
		case WindowCumulative:
			aggregator = newAggregatorCumulative(now, newAggregateValue)
		case WindowSlidingTime:
			aggregator = newAggregatorSlidingTime(now, w.duration, w.subIntervals, newAggregateValue)
		default:
			// TODO: panic here. This should never be reached. If it is, then it is a bug.
		}
		ab.signatures[s] = aggregator
	}
	aggregator.addSample(v, now)
}

func (ab *AggregationBase) collectedRows(keys []tags.Key, now time.Time) []*Row {
	var rows []*Row

	for sig, aggregator := range ab.signatures {
		ts := tags.ToTagSet(sig, keys)
		rows = append(rows, &Row{
			ts,
			aggregator.retrieveCollected(now),
		})
	}
	return rows
}

func (ab *AggregationBase) clearRows() {
	ab.signatures = make(map[string]aggregator)
}
