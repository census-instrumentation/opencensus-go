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

package metric_test

import (
	"context"
	"time"

	"go.opencensus.io/metric"
)

func ExamplePushExporter() {
	push := func(context context.Context, metrics []*metric.Metric) error {
		// publish metrics to monitoring backend ...
		return nil
	}
	var pe metric.PushExporter
	pe.Init(metric.DefaultRegistry(), push)
	pe.Timeout = 10 * time.Second
	pe.ReportingPeriod = 5 * time.Second
	go pe.Run()
	time.Sleep(10 * time.Second)
	pe.Stop()
}
