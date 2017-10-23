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

// Aggregation represents a data aggregation method. There are several
// aggregation methods made available by the package such as
// CountAggregation and DistributionAggregation.
type Aggregation interface {
	isAggregation() bool
	aggregationValueConstructor() func() AggregationValue
}

// CountAggregation indicates that data collected and aggregated
// with this method will be turned into a count value.
// For example, total number of accepted requests can be
// aggregated by using CountAggregation.
type CountAggregation struct{}

func (a CountAggregation) isAggregation() bool { return true }

func (a CountAggregation) aggregationValueConstructor() func() AggregationValue {
	return func() AggregationValue { return newCountAggregationValue(0) }
}

// DistributionAggregation indicates that the desired aggregation is
// a histogram distribution.
// An distribution aggregation may contain a histogram of the values in the
// population. The bucket boundaries for that histogram are described
// by DistributionAggregation slice. This defines length+1 buckets.
//
// If length >= 2 then the boundaries for bucket index i are:
//
//     [-infinity, bounds[i]) for i = 0
//     [bounds[i-1], bounds[i]) for 0 < i < length
//     [bounds[i-1], +infinity) for i = length
//
// If length is 0 then there is no histogram associated with the
// distribution. There will be a single bucket with boundaries
// (-infinity, +infinity).
//
// If length is 1 then there is no finite buckets, and that single
// element is the common boundary of the overflow and underflow buckets.
type DistributionAggregation []float64

func (a DistributionAggregation) isAggregation() bool { return true }

func (a DistributionAggregation) aggregationValueConstructor() func() AggregationValue {
	return func() AggregationValue { return newDistributionAggregationValue([]float64(a)) }
}
