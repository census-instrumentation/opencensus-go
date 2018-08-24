package stats

import (
	"sync/atomic"

	"go.opencensus.io/tag"
)

var (
	registry     = map[Exporter]struct{}{}
	registrySize int32
)

// Data encapsulates the contents of the stats to be exported
type Data struct {
	Tags         *tag.Map
	Measurements []Measurement
}

// Exporter exports the collected records as view data.
//
// The ExportView method should return quickly; if an
// Exporter takes a significant amount of time to
// process a Data, that work should be done on another goroutine.
//
// The Data should not be modified.
type Exporter interface {
	ExportStats(data Data)
}

// RegisterExporter registers an exporter.
// Collected data will be reported via all the
// registered exporters. Once you no longer
// want data to be exported, invoke UnregisterExporter
// with the previously registered exporter.
//
// Binaries can register exporters, libraries shouldn't register exporters.
func RegisterExporter(exporter Exporter) {
	mu.Lock()
	defer mu.Unlock()

	updated := map[Exporter]struct{}{}
	for k, v := range registry {
		updated[k] = v
	}

	updated[exporter] = struct{}{}
	registry = updated
	atomic.StoreInt32(&registrySize, int32(len(updated)))
}

// UnregisterExporter unregisters an exporter.
func UnregisterExporter(exporter Exporter) {
	mu.Lock()
	defer mu.Unlock()

	updated := map[Exporter]struct{}{}
	for k, v := range registry {
		updated[k] = v
	}

	delete(updated, exporter)
	registry = updated
}
