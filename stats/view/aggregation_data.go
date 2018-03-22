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

package view

import (
	"math"
)

// aggregator represents an aggregated value from a collection.
// They are reported on the view data during exporting.
// Mosts users won't directly access aggregration data.
type aggregator interface {
	addSample(v float64)
	writeTo(d *AggregationData)
}

const epsilon = 1e-9

// countData is the aggregated data for the Count aggregation.
// A count aggregation processes data and counts the recordings.
//
// Most users won't directly access count data.
type countData int64

func newCountData(v int64) *countData {
	tmp := countData(v)
	return &tmp
}

func newCountDist(v int64) AggregationData {
	var dd AggregationData
	newCountData(v).writeTo(&dd)
	return dd
}

func (a *countData) addSample(_ float64) {
	*a = *a + 1
}

func (a *countData) writeTo(dd *AggregationData) {
	dd.Count = int64(*a)
}

// sumData is the aggregated data for the Sum aggregation.
// A sum aggregation processes data and sums up the recordings.
//
// Most users won't directly access sum data.
type sumData float64

func newSumData(v float64) *sumData {
	tmp := sumData(v)
	return &tmp
}

func newSumDist(v float64) AggregationData {
	return AggregationData{Count: 1, Mean: v}
}

func (a *sumData) addSample(f float64) {
	*a += sumData(f)
}

func (a *sumData) writeTo(dd *AggregationData) {
	dd.Mean = float64(*a)
	dd.Count = 1
}

// meanData is the aggregated data for the Mean aggregation.
// A mean aggregation processes data and maintains the mean value.
//
// Most users won't directly access mean data.
type meanData struct {
	Count int64   // number of data points aggregated
	Mean  float64 // mean of all data points
}

func newMeanData(mean float64, count int64) *meanData {
	return &meanData{
		Mean:  mean,
		Count: count,
	}
}

func newMeanDist(mean float64, count int64) AggregationData {
	return AggregationData{
		Mean:  mean,
		Count: count,
	}
}

// Sum returns the sum of all samples collected.
func (a *meanData) Sum() float64 { return a.Mean * float64(a.Count) }

func (a *meanData) addSample(f float64) {
	a.Count++
	if a.Count == 1 {
		a.Mean = f
		return
	}
	a.Mean = a.Mean + (f-a.Mean)/float64(a.Count)
}

func (a *meanData) writeTo(dd *AggregationData) {
	dd.Count = a.Count
	dd.Mean = a.Mean
}

// AggregationData is the aggregated data for exporters. Depending on the view
// AggType, not all fields may be set.
//
// Most users won't directly access AggregationData, only exporters.
type AggregationData struct {
	Count           int64     // number of data points aggregated
	Min             float64   // minimum value in the distribution
	Max             float64   // max value in the distribution
	Mean            float64   // mean of the distribution
	SumOfSquaredDev float64   // sum of the squared deviation from the mean
	CountPerBucket  []int64   // number of occurrences per bucket
	bounds          []float64 // histogram distribution of the values
}

func (a AggregationData) Equal(a2 AggregationData) bool {
	if len(a.CountPerBucket) != len(a2.CountPerBucket) {
		return false
	}
	for i := range a.CountPerBucket {
		if a.CountPerBucket[i] != a2.CountPerBucket[i] {
			return false
		}
	}
	return a.Count == a2.Count && a.Min == a2.Min && a.Max == a2.Max && math.Pow(a.Mean-a2.Mean, 2) < epsilon && math.Pow(a.variance()-a2.variance(), 2) < epsilon
}

func newDistributionData(bounds []float64) *AggregationData {
	return &AggregationData{
		CountPerBucket: make([]int64, len(bounds)+1),
		bounds:         bounds,
		Min:            math.MaxFloat64,
		Max:            math.SmallestNonzeroFloat64,
	}
}

// Sum returns the sum of all samples collected.
func (a *AggregationData) Sum() float64 { return a.Mean * float64(a.Count) }

func (a *AggregationData) variance() float64 {
	if a.Count <= 1 {
		return 0
	}
	return a.SumOfSquaredDev / float64(a.Count-1)
}

func (a *AggregationData) addSample(f float64) {
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

func (a *AggregationData) incrementBucketCount(f float64) {
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

func (a *AggregationData) writeTo(dd *AggregationData) {
	*dd = *a
	dd.CountPerBucket = make([]int64, len(a.CountPerBucket))
	copy(dd.CountPerBucket, a.CountPerBucket)
}
