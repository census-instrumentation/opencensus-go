package runmetrics_test

import (
	"context"
	"fmt"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricexport"
	"go.opencensus.io/metric/metricproducer"
	"go.opencensus.io/plugin/runmetrics"
	"log"
	"sort"
)

type printExporter struct {
}

func (l *printExporter) ExportMetrics(ctx context.Context, data []*metricdata.Metric) error {
	mapData := make(map[string]metricdata.Metric, 0)

	for _, v := range data {
		mapData[v.Descriptor.Name] = *v
	}

	mapKeys := make([]string, 0, len(mapData))
	for key := range mapData {
		mapKeys = append(mapKeys, key)
	}
	sort.Strings(mapKeys)

	// for the sake of a simple example, we cannot use the real value here
	simpleVal := func(v interface{}) int { return 42 }

	for _, k := range mapKeys {
		v := mapData[k]
		fmt.Printf("%s %d\n", k, simpleVal(v.TimeSeries[0].Points[0].Value))
	}

	return nil
}

func ExampleProducer() {

	// Create a new runmetrics.Producer and supply options
	runtimeMetrics, err := runmetrics.NewProducer(runmetrics.RunMetricOptions{
		EnableCPU:    true,
		EnableMemory: true,
		Prefix:       "mayapp_",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Register the producer
	metricproducer.GlobalManager().AddProducer(runtimeMetrics)

	// Use your reader/exporter to extract values
	// This part is not specific to runtime metrics and only here to make it a complete example.
	metricexport.NewReader().ReadAndExport(&printExporter{})

	// output:
	// mayapp_cpu_cgo_calls 42
	// mayapp_cpu_goroutines 42
	// mayapp_heap_alloc 42
	// mayapp_heap_idle 42
	// mayapp_heap_inuse 42
	// mayapp_heap_objects 42
	// mayapp_heap_release 42
	// mayapp_heap_sys 42
	// mayapp_mem_alloc 42
	// mayapp_mem_frees 42
	// mayapp_mem_lookups 42
	// mayapp_mem_malloc 42
	// mayapp_mem_sys 42
	// mayapp_mem_total 42
	// mayapp_stack_inuse 42
	// mayapp_stack_mcache_inuse 42
	// mayapp_stack_mcache_sys 42
	// mayapp_stack_mspan_inuse 42
	// mayapp_stack_mspan_sys 42
	// mayapp_stack_sys 42
}
