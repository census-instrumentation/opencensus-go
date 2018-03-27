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
	viewExportersMu sync.RWMutex // guards exporters
	viewExporters   = make(map[ViewExporter]struct{})
)

// ViewExporter exports the collected view data to a monitoring backend.
//
// The ExportView method should return quickly; if an
// Exporter takes a significant amount of time to
// process a ViewData, that work should be done on another goroutine.
//
// The ViewData should not be modified.
type ViewExporter interface {
	ExportView(viewData *ViewData)
}

// Register registers a View exporter.
// Collected data will be reported via all the registered exporters.
// If you no longer want data to be exported, you can invoke Unregister
// with the previously registered exporter.
// Exporters are required to be valid map keys.
func Register(exporter ViewExporter) {
	viewExportersMu.Lock()
	defer viewExportersMu.Unlock()
	if ev, ok := exporter.(ViewExporter); ok {
		viewExporters[ev] = struct{}{}
	}
}

// Unregister unregisters a previously registered exporter.
func Unregister(exporter ViewExporter) {
	viewExportersMu.Lock()
	defer viewExportersMu.Unlock()
	if ev, ok := exporter.(ViewExporter); ok {
		delete(viewExporters, ev)
	}
}

// CallViewExporters calls all registered View exporters with the given ViewData.
func CallViewExporters(viewData *ViewData) {
	viewExportersMu.Lock()
	defer viewExportersMu.Unlock()
	for e := range viewExporters {
		e.ExportView(viewData)
	}
}
