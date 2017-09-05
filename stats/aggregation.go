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

// Package stats defines the stats collection API and its implementation.
package stats

// Aggregation is the generic interface for all aggregtion types.
type Aggregation interface {
	isAggregation() bool
	aggregationValueConstructor() func() AggregationValue
}

// AggregationCount indicates that the desired aggregation is count.
type AggregationCount struct{}

// NewAggregationCount creates a new aggregation of type count.
func NewAggregationCount() *AggregationCount {
	return &AggregationCount{}
}

func (a *AggregationCount) isAggregation() bool { return true }

func (a *AggregationCount) aggregationValueConstructor() func() AggregationValue {
	return func() AggregationValue { return newAggregationCountValue(0) }
}

// AggregationDistribution indicates that the desired aggregation is a histograms
// distribution.
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
	bounds []float64
}

// NewAggregationDistribution creates a new aggregation of type distribution
// a.k.a histogram.
func NewAggregationDistribution(bounds []float64) *AggregationDistribution {
	var copyBounds []float64
	for _, b := range bounds {
		copyBounds = append(copyBounds, b)
	}

	return &AggregationDistribution{
		bounds: copyBounds,
	}
}

func (a *AggregationDistribution) isAggregation() bool { return true }

func (a *AggregationDistribution) aggregationValueConstructor() func() AggregationValue {
	return func() AggregationValue { return newAggregationDistributionValue(a.bounds) }
}
