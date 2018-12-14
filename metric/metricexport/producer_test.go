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

package metricexport

import (
	"go.opencensus.io/metric/metricdata"
	"testing"
)

func TestRegistry_AddProducer(t *testing.T) {
	r := NewRegistry()
	m1 := &metricdata.Metric{
		Descriptor: metricdata.Descriptor{
			Name: "test",
			Unit: metricdata.UnitDimensionless,
		},
	}
	p := &constProducer{m1}
	r.AddProducer(p)
	if got, want := len(r.ReadAll()), 1; got != want {
		t.Fatal("Expected to read a single metric")
	}
	r.RemoveProducer(p)
	if got, want := len(r.ReadAll()), 0; got != want {
		t.Fatal("Expected to read no metrics")
	}
}

type constProducer []*metricdata.Metric

func (cp constProducer) Read() []*metricdata.Metric {
	return cp
}
