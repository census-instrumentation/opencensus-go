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

package prometheus_test

import (
	"log"
	"net/http"

	"go.opencensus.io/exporter"
	"go.opencensus.io/exporter/prometheus"
)

func Example() {
	e, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatal(err)
	}
	exporter.Register(e)

	// Serve the scrap endpoint at localhost:9999.
	http.Handle("/metrics", e)
	log.Fatal(http.ListenAndServe(":9999", nil))
}
