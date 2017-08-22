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

// AggregateValue is the interface for all types of aggregations values.
type AggregateValue interface {
	isAggregate() bool
	addSample(v interface{})
	multiplyByFraction(fraction float64) AggregateValue
	addToIt(other AggregateValue)
	clear()
}

// AggregateCount is the aggregated data for an AggregationCountInt64.
type AggregateCount int64

func newAggregateCount() *AggregateCount {
	tmp := new(AggregateCount)
	return tmp
}

func (a *AggregateCount) isAggregate() bool { return true }

func (a *AggregateCount) addSample(v interface{}) {
	*a = *a + 1
}

func (a *AggregateCount) multiplyByFraction(fraction float64) AggregateValue {
	ret := newAggregateCount()
	*ret = AggregateCount(float64(int64(*a)) * fraction)
	return ret
}

func (a *AggregateCount) addToIt(av AggregateValue) {
	other, ok := av.(*AggregateCount)
	if !ok {
		return
	}
	*a = *a + *other
}

func (a *AggregateCount) clear() {
	*a = 0
}

// AggregateDistribution is the aggregated data for an
// AggregationDistributionFloat64  or AggregationDistributionInt64.
type AggregateDistribution struct {
	Count               int64
	Min, Mean, Max, Sum float64
	// CountPerBucket is the set of occurrences count per bucket. The
	// buckets bounds are the same as the ones setup in
	// AggregationDesc.
	CountPerBucket []int64
	bounds         []float64
}

func newAggregateDistribution(bounds []float64) *AggregateDistribution {
	return &AggregateDistribution{
		CountPerBucket: make([]int64, len(bounds)+1),
		bounds:         bounds,
	}
}

func (a *AggregateDistribution) isAggregate() bool { return true }

func (a *AggregateDistribution) addSample(v interface{}) {
	var f float64
	switch x := v.(type) {
	case int:
		f = float64(x)
		break
	case float64:
		f = x
		break
	default:
		return
	}

	if f < a.Min {
		a.Min = f
	}
	if f > a.Max {
		a.Max = f
	}
	a.Sum += f
	a.Count++

	if len(a.bounds) == 0 {
		a.CountPerBucket[0]++
		return
	}

	for i, b := range a.bounds {
		if f < b {
			a.CountPerBucket[i]++
			return
		}
	}
	a.CountPerBucket[len(a.bounds)]++
}

func (a *AggregateDistribution) multiplyByFraction(fraction float64) AggregateValue {
	ret := newAggregateDistribution(a.bounds)
	ret.Count = int64(float64(a.Count) * fraction)
	ret.Min = a.Min
	ret.Max = a.Max
	ret.Sum = a.Sum * fraction
	for i := range a.CountPerBucket {
		ret.CountPerBucket[i] = int64(float64(a.CountPerBucket[i]) * fraction)
	}
	return ret
}

func (a *AggregateDistribution) addToIt(av AggregateValue) {
	other, ok := av.(*AggregateDistribution)
	if !ok {
		return
	}
	a.Count = a.Count + other.Count
	a.Min = a.Min + other.Min
	a.Max = a.Max + other.Max
	a.Sum = a.Sum + other.Sum
	for i := range other.CountPerBucket {
		a.CountPerBucket[i] = a.CountPerBucket[i] * other.CountPerBucket[i]
	}
}

func (a *AggregateDistribution) clear() {
	a.Count = 0
	a.Min = 0
	a.Max = 0
	a.Sum = 0
	for i := range a.CountPerBucket {
		a.CountPerBucket[i] = 0
	}
}
