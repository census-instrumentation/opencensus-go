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

	"go.opencensus.io/stats/viewexporter"
)

// aggregator receives data points and aggregates them in place.
type aggregator interface {
	addSample(v float64)
	exportTo(d *viewexporter.AggregationData)
}

// TODO(ramonza): remove all the aggregator types and replace with just a
// func(*viewexporter.AggregationData) that updates its argument in-place.

type countData int64

func newCountData(v int64) *countData {
	tmp := countData(v)
	return &tmp
}

func (a *countData) addSample(_ float64) {
	*a = *a + 1
}

func (a *countData) exportTo(ad *viewexporter.AggregationData) {
	ad.Count = int64(*a)
}

type sumData float64

func newSumData(v float64) *sumData {
	tmp := sumData(v)
	return &tmp
}

func (a *sumData) addSample(f float64) {
	*a += sumData(f)
}

func (a *sumData) exportTo(ad *viewexporter.AggregationData) {
	ad.Mean = float64(*a)
	ad.Count = 1
}

type distributionData struct {
	Count           int64     // number of data points aggregated
	Min             float64   // minimum value in the distribution
	Max             float64   // max value in the distribution
	Mean            float64   // mean of the distribution
	SumOfSquaredDev float64   // sum of the squared deviation from the mean
	CountPerBucket  []int64   // number of occurrences per bucket
	Bounds          []float64 // histogram distribution of the values
}

func newDistributionData(bounds []float64) *distributionData {
	return &distributionData{
		CountPerBucket: make([]int64, len(bounds)+1),
		Bounds:         bounds,
		Min:            math.MaxFloat64,
		Max:            math.SmallestNonzeroFloat64,
	}
}

func (a *distributionData) addSample(f float64) {
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

func (a *distributionData) incrementBucketCount(f float64) {
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

func (a *distributionData) exportTo(ad *viewexporter.AggregationData) {
	*ad = viewexporter.AggregationData(*a)
	ad.CountPerBucket = make([]int64, len(a.CountPerBucket))
	copy(ad.CountPerBucket, a.CountPerBucket)
}

// LastValueData returns the last value recorded for LastValue aggregation.
type LastValueData struct {
	Value float64
}

func (l *LastValueData) isAggregationData() bool {
	return true
}

func (l *LastValueData) addSample(v float64) {
	l.Value = v
}

func (l *LastValueData) exportTo(ad *viewexporter.AggregationData) {
	ad.Mean = l.Value
}
