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
	exportersMu   sync.RWMutex // guards exporters
	viewExporters = make(map[View]struct{})
)

// View exports the collected view data.
//
// The ExportView method should return quickly; if an
// Exporter takes a significant amount of time to
// process a ViewData, that work should be done on another goroutine.
//
// The ViewData should not be modified.
type View interface {
	ExportView(viewData *ViewData)
}

// Register registers an exporter. The exporter should implement one of the
// export interfaces: exporter.View.
// Collected data will be reported via all the registered exporters.
// If you no longer want data to be exported, invoke Unregister with the
// previously registered exporter.
// Exporters are required to be valid map keys.
func Register(exporter interface{}) {
	exportersMu.Lock()
	defer exportersMu.Unlock()
	if ev, ok := exporter.(View); ok {
		viewExporters[ev] = struct{}{}
	}
}

// Unregister unregisters a previously registered exporter.
func Unregister(exporter interface{}) {
	exportersMu.Lock()
	defer exportersMu.Unlock()
	if ev, ok := exporter.(View); ok {
		delete(viewExporters, ev)
	}
}

// ExportViewData calls all registered View exporters with the given ViewData.
func ExportViewData(viewData *ViewData) {
	exportersMu.Lock()
	defer exportersMu.Unlock()
	for e := range viewExporters {
		e.ExportView(viewData)
	}
}
