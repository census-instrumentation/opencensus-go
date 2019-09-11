package runmetrics_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricexport"
	"go.opencensus.io/metric/metricproducer"
	"go.opencensus.io/plugin/runmetrics"
	"testing"
)

type testExporter struct {
	data []*metricdata.Metric
}

func (t *testExporter) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	t.data = append(t.data, data...)
	return nil
}

func TestEnable(t *testing.T) {
	tests := []struct {
		name                string
		options             runmetrics.RunMetricOptions
		wantMetricNames     [][]string
		dontWantMetricNames [][]string
	}{
		{
			"no stats",
			runmetrics.RunMetricOptions{
				EnableCPU:    false,
				EnableMemory: false,
			},
			[][]string{},
			[][]string{},
		},
		{
			"cpu and memory stats",
			runmetrics.RunMetricOptions{
				EnableCPU:    true,
				EnableMemory: true,
			},
			[][]string{
				{"mem_alloc", "mem_total", "mem_sys", "mem_lookups", "mem_malloc", "mem_frees"},
				{"heap_alloc", "heap_sys", "heap_idle", "heap_inuse", "heap_objects", "heap_release"},
				{"stack_inuse", "stack_sys", "stack_mspan_inuse", "stack_mspan_sys", "stack_mcache_inuse", "stack_mspan_sys"},
				{"cpu_goroutines", "cpu_cgo_calls"},
			},
			[][]string{},
		},
		{
			"only cpu stats",
			runmetrics.RunMetricOptions{
				EnableCPU:    true,
				EnableMemory: false,
			},
			[][]string{
				{"cpu_goroutines", "cpu_cgo_calls"},
			},
			[][]string{
				{"mem_alloc", "mem_total", "mem_sys", "mem_lookups", "mem_malloc", "mem_frees"},
				{"heap_alloc", "heap_sys", "heap_idle", "heap_inuse", "heap_objects", "heap_release"},
				{"stack_inuse", "stack_sys", "stack_mspan_inuse", "stack_mspan_sys", "stack_mcache_inuse", "stack_mspan_sys"},
			},
		},
		{
			"only memory stats",
			runmetrics.RunMetricOptions{
				EnableCPU:    false,
				EnableMemory: true,
			},
			[][]string{
				{"mem_alloc", "mem_total", "mem_sys", "mem_lookups", "mem_malloc", "mem_frees"},
				{"heap_alloc", "heap_sys", "heap_idle", "heap_inuse", "heap_objects", "heap_release"},
				{"stack_inuse", "stack_sys", "stack_mspan_inuse", "stack_mspan_sys", "stack_mcache_inuse", "stack_mspan_sys"},
			},
			[][]string{
				{"cpu_goroutines", "cpu_cgo_calls"},
			},
		},
		{
			"cpu and memory stats with custom prefix",
			runmetrics.RunMetricOptions{
				EnableCPU:    true,
				EnableMemory: true,
				Prefix:       "test_",
			},
			[][]string{
				{"test_mem_alloc", "test_mem_total", "test_mem_sys", "test_mem_lookups", "test_mem_malloc", "test_mem_frees"},
				{"test_heap_alloc", "test_heap_sys", "test_heap_idle", "test_heap_inuse", "test_heap_objects", "test_heap_release"},
				{"test_stack_inuse", "test_stack_sys", "test_stack_mspan_inuse", "test_stack_mspan_sys", "test_stack_mcache_inuse", "test_stack_mspan_sys"},
				{"test_cpu_goroutines", "test_cpu_cgo_calls"},
			},
			[][]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			err := runmetrics.Enable(test.options)

			if err != nil {
				t.Errorf("want: nil, got: %v", err)
			}

			defer runmetrics.Disable()

			exporter := &testExporter{}
			reader := metricexport.NewReader()
			reader.ReadAndExport(exporter)

			for _, want := range test.wantMetricNames {
				assertNames(t, true, exporter, want)
			}

			for _, dontWant := range test.dontWantMetricNames {
				assertNames(t, false, exporter, dontWant)
			}
		})
	}
}

func assertNames(t *testing.T, wantIncluded bool, exporter *testExporter, expectedNames []string) {
	t.Helper()

	metricNames := make([]string, 0)
	for _, v := range exporter.data {
		metricNames = append(metricNames, v.Descriptor.Name)
	}

	for _, want := range expectedNames {
		if wantIncluded {
			assert.Contains(t, metricNames, want)
		} else {
			assert.NotContains(t, metricNames, want)
		}
	}
}

func TestEnable_RegistersWithGlobalManager(t *testing.T) {
	err := runmetrics.Enable(runmetrics.RunMetricOptions{})
	if err != nil {
		t.Errorf("want: nil, got: %v", err)
	}

	registeredCount := len(metricproducer.GlobalManager().GetAll())
	assert.Equal(t, 1, registeredCount, "expected a producer to be registered")
}

func TestEnable_RegistersNoDuplicates(t *testing.T) {
	err := runmetrics.Enable(runmetrics.RunMetricOptions{})
	if err != nil {
		t.Errorf("want: nil, got: %v", err)
	}

	err = runmetrics.Enable(runmetrics.RunMetricOptions{})
	if err != nil {
		t.Errorf("want: nil, got: %v", err)
	}

	producerCount := len(metricproducer.GlobalManager().GetAll())
	assert.Equal(t, 1, producerCount, "expected one registered producer")
}

func TestDisable(t *testing.T) {
	err := runmetrics.Enable(runmetrics.RunMetricOptions{})
	if err != nil {
		t.Errorf("want: nil, got: %v", err)
	}

	runmetrics.Disable()

	producerCount := len(metricproducer.GlobalManager().GetAll())
	assert.Equal(t, 0, producerCount, "expected one registered producer")
}
