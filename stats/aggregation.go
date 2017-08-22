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

import "github.com/google/working-instrumentation-go/tags"

// Aggregation is the generic interface for all aggregtion types.
type Aggregation interface {
	isAggregation() bool
	clearRows()
	collectedRows(keys []tags.Key) []*Row
}

// AggregationInt64 is the generic interface for all aggregtion  of type int64.
type AggregationInt64 interface {
	Aggregation
	addSample(signature string, i int64)
}

// AggregationCountInt64 is the struct representing the count aggregation.
type AggregationCountInt64 struct {
	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its AggregateValueInt64.
	signatures map[string]*AggregateCount
}

func (a *AggregationCountInt64) isAggregation() bool { return true }

func (a *AggregationCountInt64) clearRows() {
	a.signatures = make(map[string]*AggregateCount)
}

func (a *AggregationCountInt64) addSample(s string, i int64) {
	v, ok := a.signatures[s]
	if !ok {
		v := newAggregateCount()
		a.signatures[s] = v
	}
	v.addSample()
}

func (a *AggregationCountInt64) collectedRows(keys []tags.Key) []*Row {
	var rows []*Row

	for sig, v := range a.signatures {
		ts := tags.ToTagSet(sig, keys)
		rows = append(rows, &Row{
			ts,
			v,
		})
	}
	return rows
}

// AggregationDistributionInt64 is the struct representing the distribution
// aggregation of type int64.
type AggregationDistributionInt64 struct {
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

	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its AggregateValueInt64.
	signatures map[string]*AggregateDistribution
}

func (a *AggregationDistributionInt64) isAggregation() bool { return true }

func (a *AggregationDistributionInt64) clearRows() {
	a.signatures = make(map[string]*AggregateDistribution)
}

func (a *AggregationDistributionInt64) addSample(s string, i int64) {
	v, ok := a.signatures[s]
	if !ok {
		v := newAggregateDistribution(a.Bounds)
		a.signatures[s] = v
	}
	v.addSampleInt64(i, a.Bounds)
}

func (a *AggregationDistributionInt64) collectedRows(keys []tags.Key) []*Row {
	var rows []*Row

	for sig, v := range a.signatures {
		ts := tags.ToTagSet(sig, keys)
		rows = append(rows, &Row{
			ts,
			v,
		})
	}
	return rows
}

// AggregationFloat64 is the generic interface for all aggregtion  of type
// float64.
type AggregationFloat64 interface {
	Aggregation
	addSample(signature string, f float64)
}

// AggregationDistributionFloat64 is the struct representing the distribution
// aggregation of type float64.
type AggregationDistributionFloat64 struct {
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

	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its AggregateValueFloat64.
	signatures map[string]*AggregateDistribution
}

func (a *AggregationDistributionFloat64) isAggregation() bool { return true }

func (a *AggregationDistributionFloat64) clearRows() {
	a.signatures = make(map[string]*AggregateDistribution)
}

func (a *AggregationDistributionFloat64) addSample(s string, f float64) {
	v, ok := a.signatures[s]
	if !ok {
		v := newAggregateDistribution(a.Bounds)
		a.signatures[s] = v
	}
	v.addSampleFloat64(f, a.Bounds)
}

func (a *AggregationDistributionFloat64) collectedRows(keys []tags.Key) []*Row {
	var rows []*Row

	for sig, v := range a.signatures {
		ts := tags.ToTagSet(sig, keys)
		rows = append(rows, &Row{
			ts,
			v,
		})
	}
	return rows
}
