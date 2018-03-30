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

package ochttp_test

import (
	"fmt"
	"runtime"
	"testing"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
)

// This code serves to assert some sanity checks about
// the attributes of views such as the Aggregation.

func TestSanityCheckViewAggregations(t *testing.T) {
	mustHaveAggregation(t, ochttp.ClientResponseCountByStatusCode, view.AggTypeDistribution)
	mustHaveAggregation(t, ochttp.ServerResponseCountByStatusCode, view.AggTypeDistribution)
	mustHaveAggregation(t, ochttp.ClientRequestCountByMethod, view.AggTypeCount)
	mustHaveAggregation(t, ochttp.ServerRequestCountByMethod, view.AggTypeCount)
}

func caller2() string {
	pc, _, line, _ := runtime.Caller(2)
	fn := runtime.FuncForPC(pc)
	return fmt.Sprintf("%s::%d", fn.Name(), line)
}

func mustHaveAggregation(t *testing.T, v *view.View, aggType view.AggType) {
	if g, w := v.Aggregation.Type, aggType; g != w {
		t.Errorf("Location %q:\n\tnon-matching aggregation got = %v want = %v", caller2(), g, w)
	}
}
