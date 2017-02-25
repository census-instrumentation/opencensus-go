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

package stats

import (
	"bytes"
	"fmt"
	"math"
	"time"
)

// DistributionStats records a distribution of float64 sample values.
// It is the result of a DistributionAgg aggregation.
type DistributionStats struct {
	Count               int64
	Min, Mean, Max, Sum float64
	// CountPerBucket is the set of occurrences count per bucket. The
	// buckets bounds are the same as the ones setup in
	// AggregationDesc.
	CountPerBucket []int64
}

func (ds *DistributionStats) stringWithIndent(tabs string) string {
	if ds == nil {
		return "nil"
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%T {\n", ds)
	fmt.Fprintf(&buf, "%v  Count: %v,\n", tabs, ds.Count)
	fmt.Fprintf(&buf, "%v  Min: %v,\n", tabs, ds.Min)
	fmt.Fprintf(&buf, "%v  Mean: %v,\n", tabs, ds.Mean)
	fmt.Fprintf(&buf, "%v  Max: %v,\n", tabs, ds.Max)
	fmt.Fprintf(&buf, "%v  Sum: %v,\n", tabs, ds.Sum)
	fmt.Fprintf(&buf, "%v  CountPerBucket: %v,\n", tabs, ds.CountPerBucket)
	fmt.Fprintf(&buf, "%v}", tabs)
	return buf.String()
}

func (ds *DistributionStats) String() string {
	return ds.stringWithIndent("")
}

// newDistributionAggregator creates a distributionAggregator. For a single
// DistributionAggregationDescriptor it is expected to be called multiple
// times. Once for each unique set of tags.
func newDistributionAggregator(bounds []float64) *distributionAggregator {
	return &distributionAggregator{
		bounds: bounds,
		ds: &DistributionStats{
			Min:            math.MaxFloat64,
			Max:            -math.MaxFloat64,
			CountPerBucket: make([]int64, len(bounds)+1),
		},
	}
}

type distributionAggregator struct {
	bounds []float64
	ds     *DistributionStats
}

func (da *distributionAggregator) addSample(m Measurement, _ time.Time) {
	v := m.float64()
	if v < da.ds.Min {
		da.ds.Min = v
	}
	if v > da.ds.Max {
		da.ds.Max = v
	}
	da.ds.Sum += v
	da.ds.Count++

	if len(da.bounds) == 0 {
		da.ds.CountPerBucket[0]++
		return
	}

	for i, b := range da.bounds {
		if v < b {
			da.ds.CountPerBucket[i]++
			return
		}
	}
	da.ds.CountPerBucket[len(da.bounds)]++
}

func (da *distributionAggregator) retrieveCollected() *DistributionStats {
	if da.ds.Count != 0 {
		da.ds.Mean = da.ds.Sum / float64(da.ds.Count)
	}
	return da.ds
}
