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

package gauge_test

import (
	"net/http"

	"go.opencensus.io/metric/metricexport"

	"go.opencensus.io/metric"
	"go.opencensus.io/metric/gauge"
)

func ExampleInt64() {
	g := gauge.NewInt64("active_request", "Number of active requests, per method.", metric.UnitDimensionless, "method")
	metricexport.DefaultRegistry().AddProducer(g)

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		g.Add(1, metric.NewLabelValue(request.Method))
		defer g.Add(-1, metric.NewLabelValue(request.Method))
		// process request ...
	})
}
