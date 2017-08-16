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

package stats2

//-----------------------------------------------------------------------------
// Part of measurements.go
//-----------------------------------------------------------------------------

/* TODO(acetechnologist): add support for other types: AggContinuousStatsInt64,
// AggGaugeStatsInt64, AggGaugeStatsBool, AggGaugeStatsString
type AggContinuousStatsInt64 struct {
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

type AggGaugeStatsInt64 struct {
	Value float64
	tags  []tags.Tag
}

type AggGaugeStatsBool struct {
	Value bool
	tags  []tags.Tag
}

type AggGaugeStatsString struct {
	Value bool
	tags  []tags.Tag
}
*/
