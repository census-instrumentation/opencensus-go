// Copyright 2018, OpenCensus Authors
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

package viewexporter

import (
	"math"
)

const epsilon = 1e-9

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
	Bounds          []float64 // histogram distribution of the values
}

// Equal compares two data points and returns true if they are mostly the same,
// except for small floating point differences. Useful for testing.
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

// Sum returns the sum of all samples collected.
func (a *AggregationData) Sum() float64 { return a.Mean * float64(a.Count) }

func (a *AggregationData) variance() float64 {
	if a.Count <= 1 {
		return 0
	}
	return a.SumOfSquaredDev / float64(a.Count-1)
}
