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
	Tags           []tags.Tag
	AggregateValue AggregateValue
}

type AggregateValue interface {
	isAggregate() bool
}

type AggregateValueFloat64 interface {
	AggregateValue
	addSampleFloat64(f float64)
}

type AggregateValueInt64 interface {
	AggregateValue
	addSampleInt64(f float64)
}


type AggregateCount int64

func (a *AggregateCount) isAggregate() bool { return true }

func (a *AggregateCount) addSample() {
	*a = *a + 1
}

type AggregateDistribution struct {
	Count               int64
	Min, Mean, Max, Sum float64
	// CountPerBucket is the set of occurrences count per bucket. The
	// buckets bounds are the same as the ones setup in
	// AggregationDesc.
	CountPerBucket []int64
}

func (a *AggregateDistribution) isAggregate() bool { return true }

func (a *AggregateDistribution) addSampleFloat64(f float64, bounds []float64) {
	if f < a.Min {
		a.Min = f
	}
	if f > a.Max {
		a.Max = f
	}
	a.Sum += f
	a.Count++

	if len(bounds) == 0 {
		a.CountPerBucket[0]++
		return
	}

	for i, b := range bounds {
		if f < b {
			a.CountPerBucket[i]++
			return
		}
	}
	a.CountPerBucket[len(bounds)]++
}
