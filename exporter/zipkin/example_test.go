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

package zipkin_test

import (
	"log"

	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/trace"
)

func Example() {
	exporter, err := zipkin.NewExporterWithOptions(&zipkin.Options{
		EndpointHostPort: "192.168.1.5:5454",
		ReporterURI:      "localhost:9411/api/v2/spans",
	})
	if err != nil {
		log.Fatalf("zipkin.NewExporter: %v", err)
	}
	trace.RegisterExporter(exporter)
}
