package view

import "testing"

// Exporting functionality is tested in `worker_test.go`. These are a trivial
// set of tests to make sure that these package-level exports will correctly
// initialize defaultWorker if necessary.
func TestRegisterExporter(t *testing.T) {
	stopAndClearDefaultWorker()

	e := &countExporter{}
	RegisterExporter(e)

	if _, ok := defaultWorker.exporters[e]; !ok {
		t.Errorf("exporter doesn't appear to be registered with the default worker")
	}
}

func TestUnregisterExporter(t *testing.T) {
	stopAndClearDefaultWorker()

	e := &countExporter{}
	UnregisterExporter(e)
}
