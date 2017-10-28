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
	"math"
)

// AggregationValue represents an aggregated value from a collection.
// They are reported on the view data during exporting or force
// collection. Mosts users won't directly access aggregration values.
type AggregationValue interface {
	equal(other AggregationValue) bool
	isAggregate() bool
	addSample(v interface{})
	multiplyByFraction(fraction float64) AggregationValue
	addToIt(other AggregationValue)
	clear()
}

// CountAggregationValue is the aggregated data for a count aggregation.
// A count aggregation processes data and counts the recordings.
// CountAggregationValue is the value for CountAggregation.
type CountAggregationValue int64

func newCountAggregationValue(v int64) *CountAggregationValue {
	tmp := CountAggregationValue(v)
	return &tmp
}

func (a *CountAggregationValue) isAggregate() bool { return true }

func (a *CountAggregationValue) addSample(v interface{}) {
	*a = *a + 1
}

func (a *CountAggregationValue) multiplyByFraction(fraction float64) AggregationValue {
	return newCountAggregationValue(int64(float64(int64(*a))*fraction + 0.5)) // adding 0.5 because go runtime will take floor instead of rounding

}

func (a *CountAggregationValue) addToIt(av AggregationValue) {
	other, ok := av.(*CountAggregationValue)
	if !ok {
		return
	}
	*a = *a + *other
}

func (a *CountAggregationValue) clear() {
	*a = 0
}

func (a *CountAggregationValue) equal(other AggregationValue) bool {
	a2, ok := other.(*CountAggregationValue)
	if !ok {
		return false
	}

	return int64(*a) == int64(*a2)
}

// DistributionAggregationValue is the aggregated data for an
// DistributionAggregation.
type DistributionAggregationValue struct {
	Count           int64     // number of data points aggregated
	Min             float64   // minimum value in the distribution
	Max             float64   // max value in the distribution
	Mean            float64   // mean of the distribution
	SumOfSquaredDev float64   // sum of the squared deviation from the mean
	CountPerBucket  []int64   // number of occurrences per bucket
	Bounds          []float64 // histogram distribution of the values
}

func newDistributionAggregationValue(bounds []float64) *DistributionAggregationValue {
	return &DistributionAggregationValue{
		CountPerBucket: make([]int64, len(bounds)+1),
		Bounds:         bounds,
		Min:            math.MaxFloat64,
		Max:            math.SmallestNonzeroFloat64,
	}
}

// Sum returns the sum of all samples collected.
func (a *DistributionAggregationValue) Sum() float64 { return a.Mean * float64(a.Count) }

func (a *DistributionAggregationValue) variance() float64 {
	if a.Count <= 1 {
		return 0
	}
	return a.SumOfSquaredDev / float64(a.Count-1)
}

func (a *DistributionAggregationValue) isAggregate() bool { return true }

func (a *DistributionAggregationValue) addSample(v interface{}) {
	var f float64
	switch x := v.(type) {
	case int64:
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
	a.Count++
	a.incrementBucketCount(f)

	if a.Count == 1 {
		a.Mean = f
		return
	}

	oldMean := a.Mean
	a.Mean = a.Mean + (f-a.Mean)/float64(a.Count)
	a.SumOfSquaredDev = a.SumOfSquaredDev + (f-oldMean)*(f-a.Mean)
}

func (a *DistributionAggregationValue) incrementBucketCount(f float64) {
	if len(a.Bounds) == 0 {
		a.CountPerBucket[0]++
		return
	}

	for i, b := range a.Bounds {
		if f < b {
			a.CountPerBucket[i]++
			return
		}
	}
	a.CountPerBucket[len(a.Bounds)]++
}

// DistributionAggregationValue will not multiply by the fraction for this type
// of aggregation. The 'fraction' argument is there just to satisfy the
// interface 'AggregationValue'. For simplicity, we include the oldest partial
// bucket in its entirety when the aggregation is a distribution. We do not try
//  to multiply it by the fraction as it would make the calculation too complex
// and will create inconsistencies between sumOfSquaredDev, min, max and the
// various buckets of the histogram.
func (a *DistributionAggregationValue) multiplyByFraction(fraction float64) AggregationValue {
	ret := newDistributionAggregationValue(a.Bounds)
	for i, c := range a.CountPerBucket {
		ret.CountPerBucket[i] = c
	}
	ret.Count = a.Count
	ret.Min = a.Min
	ret.Max = a.Max
	ret.Mean = a.Mean
	ret.SumOfSquaredDev = a.SumOfSquaredDev
	return ret
}

func (a *DistributionAggregationValue) addToIt(av AggregationValue) {
	other, ok := av.(*DistributionAggregationValue)
	if !ok {
		return
	}
	if other.Count == 0 {
		return
	}
	if other.Min < a.Min {
		a.Min = other.Min
	}
	if other.Max > a.Max {
		a.Max = other.Max
	}
	delta := other.Mean - a.Mean
	a.SumOfSquaredDev = a.SumOfSquaredDev + other.SumOfSquaredDev + math.Pow(delta, 2)*float64(a.Count*other.Count)/(float64(a.Count+other.Count))

	a.Mean = (a.Sum() + other.Sum()) / float64(a.Count+other.Count)
	a.Count = a.Count + other.Count
	for i := range other.CountPerBucket {
		a.CountPerBucket[i] = a.CountPerBucket[i] + other.CountPerBucket[i]
	}
}

func (a *DistributionAggregationValue) clear() {
	a.Count = 0
	a.Min = math.MaxFloat64
	a.Max = math.SmallestNonzeroFloat64
	a.Mean = 0
	a.SumOfSquaredDev = 0
	for i := range a.CountPerBucket {
		a.CountPerBucket[i] = 0
	}
}

func (a *DistributionAggregationValue) equal(other AggregationValue) bool {
	a2, ok := other.(*DistributionAggregationValue)
	if !ok {
		return false
	}
	if a2 == nil {
		return false
	}
	if len(a.CountPerBucket) != len(a2.CountPerBucket) {
		return false
	}
	for i := range a.CountPerBucket {
		if a.CountPerBucket[i] != a2.CountPerBucket[i] {
			return false
		}
	}
	epsilon := math.Pow10(-9)
	return a.Count == a2.Count && a.Min == a2.Min && a.Max == a2.Max && math.Pow(a.Mean-a2.Mean, 2) < epsilon && math.Pow(a.variance()-a2.variance(), 2) < epsilon
}
