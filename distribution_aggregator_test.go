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
	"math"
	"reflect"
	"testing"
	"time"
)

func TestDistributionAggregator(t *testing.T) {
	tds := []struct {
		bounds, values                      []float64
		wantCount                           int64
		wantMin, wantMax, wantMean, wantSum float64
		wantCountPerBucket                  []int64
	}{
		{
			[]float64{},
			[]float64{},
			0,
			math.MaxFloat64,
			-math.MaxFloat64,
			0,
			0,
			[]int64{0},
		},
		{
			[]float64{},
			[]float64{10},
			1,
			10,
			10,
			10,
			10,
			[]int64{1},
		},
		{
			[]float64{},
			[]float64{10, 20},
			2,
			10,
			20,
			15,
			30,
			[]int64{2},
		},
		{
			[]float64{},
			[]float64{10, 20, 30},
			3,
			10,
			30,
			20,
			60,
			[]int64{3},
		},
		{
			[]float64{15},
			[]float64{},
			0,
			math.MaxFloat64,
			-math.MaxFloat64,
			0,
			0,
			[]int64{0, 0},
		},
		{
			[]float64{15},
			[]float64{10},
			1,
			10,
			10,
			10,
			10,
			[]int64{1, 0},
		},
		{
			[]float64{15},
			[]float64{10, 20},
			2,
			10,
			20,
			15,
			30,
			[]int64{1, 1},
		},
		{
			[]float64{15},
			[]float64{10, 20, 30},
			3,
			10,
			30,
			20,
			60,
			[]int64{1, 2},
		},
		{
			[]float64{15, 25},
			[]float64{},
			0,
			math.MaxFloat64,
			-math.MaxFloat64,
			0,
			0,
			[]int64{0, 0, 0},
		},
		{
			[]float64{15, 25},
			[]float64{10},
			1,
			10,
			10,
			10,
			10,
			[]int64{1, 0, 0},
		},
		{
			[]float64{15, 25},
			[]float64{10, 20},
			2,
			10,
			20,
			15,
			30,
			[]int64{1, 1, 0},
		},
		{
			[]float64{15, 25},
			[]float64{10, 20, 30},
			3,
			10,
			30,
			20,
			60,
			[]int64{1, 1, 1},
		},
	}

	for _, td := range tds {
		da := newDistributionAggregator(td.bounds)
		for _, v := range td.values {
			da.addSample(v, time.Time{})
		}

		ds := da.retrieveCollected()

		if got, want := ds.Count, td.wantCount; got != want {
			t.Errorf("got count %v (test case: %v) , want %v", got, td, want)
		}

		if got, want := ds.Min, td.wantMin; got != want {
			t.Errorf("got min %v (test case: %v) , want %v", got, td, want)
		}

		if got, want := ds.Max, td.wantMax; got != want {
			t.Errorf("got max %v (test case: %v) , want %v", got, td, want)
		}

		if got, want := ds.Mean, td.wantMean; got != want {
			t.Errorf("got mean %v (test case: %v) , want %v", got, td, want)
		}

		if got, want := ds.Sum, td.wantSum; got != want {
			t.Errorf("got sum %v (test case: %v) , want %v", got, td, want)
		}

		if got, want := ds.CountPerBucket, td.wantCountPerBucket; !reflect.DeepEqual(got, want) {
			t.Errorf("got CountPerBucket %v (test case: %v) , want %v", got, td, want)
		}
	}
}
