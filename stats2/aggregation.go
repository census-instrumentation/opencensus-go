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

package stats2

import "github.com/google/working-instrumentation-go/tags"

type Rows struct {
	Tags           []tags.Tags
	AggregateValue AggregateValue
}

type AggregateValue interface {
	isAggregate() bool
}

type Aggregation interface {
	isAggregation() bool
}

// aggDescContinuousFloat64 holds the parameters describing a histogram
// distribution.
type aggDescContinuousFloat64 struct {
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

// aggDescGaugeFloat64 describes a gauge distribution.
type aggDescGaugeFloat64 struct {
}

// aggDescContinuousInt64 holds the parameters describing a histogram
// distribution.
type aggDescContinuousInt64 struct {
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
	Bounds []int64
}

// aggDescGaugeInt64 describes a gauge distribution.
type aggDescGaugeInt64 struct {
}

// aggDescGaugeBool describes a gauge distribution.
type aggDescGaugeBool struct {
}

// aggDescGaugeString describes a gauge distribution.
type aggDescGaugeString struct {
}

type AggValueContinuousStatsFloat64 struct {
	Count         int64
	Min, Max, Sum float64
	// The sum of squared deviations from the mean of the values in the
	// population. For values x_i this is:
	//
	//     Sum[i=1..n]((x_i - mean)^2)
	//
	// Knuth, "The Art of Computer Programming", Vol. 2, page 323, 3rd edition
	// describes Welford's method for accumulating this sum in one pass.
	SumOfSquaredDeviation float64
	// CountPerBucket is the set of occurrences count per bucket. The
	// buckets bounds are the same as the ones setup in
	// Aggregation.
	CountPerBucket []int64
	tags           []tags.Tag
}

type AggValueGaugeStatsFloat64 struct {
	Value float64
	tags  []tags.Tag
}
