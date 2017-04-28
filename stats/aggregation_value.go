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
	"fmt"
	"math"
)

// AggregationValue is the interface for all types of aggregations values.
type AggregationValue interface {
	isAggregate() bool
	addSample(v interface{})
	multiplyByFraction(fraction float64) AggregationValue
	addToIt(other AggregationValue)
	clear()
}

func aggregationValueAreEqual(av1, av2 AggregationValue) bool {
	switch v1 := av1.(type) {
	case *AggregationCountValue:
		switch v2 := av2.(type) {
		case *AggregationCountValue:
			return int64(*v1) == int64(*v2)
		default:
			return false
		}
	case *AggregationDistributionValue:
		switch v2 := av2.(type) {
		case *AggregationDistributionValue:
			return aggregationDistributionValueAreEqual(v1, v2)
		default:
			return false
		}
	default:
		return false
	}
}

// AggregationCountValue is the aggregated data for an AggregationCountInt64.
type AggregationCountValue int64

func newAggregationCountValue(v int64) *AggregationCountValue {
	tmp := AggregationCountValue(v)
	return &tmp
}

func (a *AggregationCountValue) isAggregate() bool { return true }

func (a *AggregationCountValue) addSample(v interface{}) {
	*a = *a + 1
}

func (a *AggregationCountValue) multiplyByFraction(fraction float64) AggregationValue {
	return newAggregationCountValue(int64(float64(int64(*a))*fraction + 0.5)) // adding 0.5 because go runtime will take floor instead of rounding

}

func (a *AggregationCountValue) addToIt(av AggregationValue) {
	other, ok := av.(*AggregationCountValue)
	if !ok {
		return
	}
	*a = *a + *other
}

func (a *AggregationCountValue) clear() {
	*a = 0
}

func (a *AggregationCountValue) String() string {
	return fmt.Sprintf("{%v}", *a)
}

// AggregationDistributionValue is the aggregated data for an
// AggregationDistributionFloat64  or AggregationDistributionInt64.
type AggregationDistributionValue struct {
	count         int64
	min, max, sum float64
	// countPerBucket is the set of occurrences count per bucket. The buckets
	// bounds are the same as the ones setup in AggregationDistribution.
	countPerBucket []int64
	bounds         []float64
}

func newAggregationDistributionValue(bounds []float64) *AggregationDistributionValue {
	return &AggregationDistributionValue{
		countPerBucket: make([]int64, len(bounds)+1),
		bounds:         bounds,
		min:            math.MaxFloat64,
		max:            math.SmallestNonzeroFloat64,
	}
}

// Count returns the count of all samples collected.
func (a *AggregationDistributionValue) Count() int64 { return a.count }

// Min returns the min of all samples collected.
func (a *AggregationDistributionValue) Min() float64 { return a.min }

// Mean returns the mean of all samples collected.
func (a *AggregationDistributionValue) Mean() float64 { return a.sum / float64(a.count) }

// Max returns the max of all samples collected.
func (a *AggregationDistributionValue) Max() float64 { return a.max }

// Sum returns the sum of all samples collected.
func (a *AggregationDistributionValue) Sum() float64 { return a.sum }

func (a *AggregationDistributionValue) String() string {
	return fmt.Sprintf("{%v %v %v %v %v %v %v}", a.Count(), a.Min(), a.Max(), a.Mean(), a.Sum(), a.countPerBucket, a.bounds)
}

// CountPerBucket returns count per bucket. The buckets bounds are the same as
// the ones setup in AggregationDistribution.
func (a *AggregationDistributionValue) CountPerBucket() []int64 {
	var ret []int64
	for _, c := range a.countPerBucket {
		ret = append(ret, c)
	}
	return ret
}

func (a *AggregationDistributionValue) isAggregate() bool { return true }

func (a *AggregationDistributionValue) addSample(v interface{}) {
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

	if f < a.min {
		a.min = f
	}
	if f > a.max {
		a.max = f
	}
	a.sum += f
	a.count++

	a.incrementBucketCount(f)
}

func (a *AggregationDistributionValue) incrementBucketCount(f float64) {
	if len(a.bounds) == 0 {
		a.countPerBucket[0]++
		return
	}

	for i, b := range a.bounds {
		if f < b {
			a.countPerBucket[i]++
			return
		}
	}
	a.countPerBucket[len(a.bounds)]++
}

func (a *AggregationDistributionValue) multiplyByFraction(fraction float64) AggregationValue {
	ret := newAggregationDistributionValue(a.bounds)
	ret.count = int64(float64(a.count)*fraction + 0.5) // adding 0.5 because go runtime will take floor instead of rounding
	if ret.count == 0 {
		return ret
	}

	if ret.count == 1 {
		ret.min = (a.min + a.max) / 2
		ret.max = ret.min
		ret.sum = ret.min
		ret.incrementBucketCount(ret.min)
		return ret
	}

	ret.min = a.min
	ret.max = a.max
	ret.sum = ret.min + ret.max
	ret.incrementBucketCount(ret.min)
	ret.incrementBucketCount(ret.max)

	// decrementing the bucket with the lowest values to account for min
	// already added to bucket.
	for i := range a.countPerBucket {
		if a.countPerBucket[i] > 0 {
			a.countPerBucket[i] = a.countPerBucket[i] - 1
			break
		}
	}

	// decrementing the bucket with the largest values to account for max
	// already added to bucket.
	for i := len(a.countPerBucket) - 1; i >= 0; i-- {
		if a.countPerBucket[i] > 0 {
			a.countPerBucket[i] = a.countPerBucket[i] - 1
			break
		}
	}

	if len(a.bounds) == 0 {
		n := int64(float64(a.countPerBucket[0])*fraction + 0.5) // adding 0.5 because go runtime will take floor instead of rounding
		ret.countPerBucket[0] += n
		ret.sum += float64(n) * (ret.min + ret.max) / 2
		return ret
	}

	for i := range a.countPerBucket {
		n := int64(float64(a.countPerBucket[i])*fraction + 0.5) // adding 0.5 because go runtime will take floor instead of rounding
		ret.countPerBucket[i] += n

		if i == 0 {
			ret.sum += float64(n) * (ret.min + math.Min(ret.bounds[i], ret.max)) / 2
			continue
		}

		if i == len(a.countPerBucket) {
			ret.sum += float64(n) * (ret.bounds[i-1] + ret.bounds[i]) / 2
			continue
		}

		ret.sum += float64(n) * (math.Max(ret.bounds[i-1], ret.min) + ret.max) / 2
	}

	return ret
}

func (a *AggregationDistributionValue) addToIt(av AggregationValue) {
	other, ok := av.(*AggregationDistributionValue)
	if !ok {
		return
	}

	if other.min < a.min {
		a.min = other.min
	}
	if other.max > a.max {
		a.max = other.max
	}

	a.sum = a.sum + other.sum
	a.count = a.count + other.count
	for i := range other.countPerBucket {
		a.countPerBucket[i] = a.countPerBucket[i] + other.countPerBucket[i]
	}
}

func (a *AggregationDistributionValue) clear() {
	a.count = 0
	a.min = math.MaxFloat64
	a.max = math.SmallestNonzeroFloat64
	a.sum = 0
	for i := range a.countPerBucket {
		a.countPerBucket[i] = 0
	}
}

func aggregationDistributionValueAreEqual(v1, v2 *AggregationDistributionValue) bool {
	for i := range v1.countPerBucket {
		if v1.countPerBucket[i] != v2.countPerBucket[i] {
			return false
		}
	}
	return v1.Count() == v2.Count() && v1.Min() == v2.Min() && v1.Mean() == v2.Mean() && v1.Max() == v2.Max() && v1.Sum() == v2.Sum()
}
