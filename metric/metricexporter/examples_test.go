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

package metricexporter_test

import (
	"context"
	"go.opencensus.io/metric/metricexporter"
	"time"

	"go.opencensus.io/metric"
)

func ExamplePush() {
	push := func(context context.Context, metrics []*metric.Metric) error {
		// publish metrics to monitoring backend ...
		return nil
	}
	// Usually, Push will be embedded in your own custom push exporter and your
	// exporter's constructor function should call Init.
	var pe metricexporter.Push
	pe.Init(push)
	pe.Timeout = 10 * time.Second
	pe.ReportingPeriod = 5 * time.Second
	go pe.Run()
	time.Sleep(10 * time.Second)
	pe.Stop()
}
