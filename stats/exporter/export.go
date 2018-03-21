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

package exporter

import "sync"

var (
	exportersMu sync.RWMutex // guards exporters
	exporters   = make(map[Exporter]struct{})
)

// Exporter exports the collected records as view data.
//
// The ExportView method should return quickly; if an
// Exporter takes a significant amount of time to
// process a ViewData, that work should be done on another goroutine.
//
// The ViewData should not be modified.
type Exporter interface {
	ExportView(viewData *ViewData)
}

// Register registers an exporter.
// Collected data will be reported via all the
// registered exporters. Once you no longer
// want data to be exported, invoke Unregister
// with the previously registered exporter.
func Register(e Exporter) {
	exportersMu.Lock()
	defer exportersMu.Unlock()

	exporters[e] = struct{}{}
}

// Unregister unregisters an exporter.
func Unregister(e Exporter) {
	exportersMu.Lock()
	defer exportersMu.Unlock()

	delete(exporters, e)
}

func ExportToAll(viewData *ViewData) {
	exportersMu.Lock()
	for e := range exporters {
		e.ExportView(viewData)
	}
	exportersMu.Unlock()
}
