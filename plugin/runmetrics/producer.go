package runmetrics

import (
	"errors"
	"go.opencensus.io/metric"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricproducer"
	"runtime"
)

type (
	// Producer produces runtime metrics.
	//
	// Create a new Producer with NewProducer().
	// A Producer should then be registered with the global manager:
	// metricproducer.GlobalManager().AddProducer()
	Producer struct {
		options RunMetricOptions
		reg     *metric.Registry

		memStats *memStats
		cpuStats *cpuStats
	}

	// RunMetricOptions allows to configure Producer.
	RunMetricOptions struct {
		EnableCPU    bool   // EnableCPU whether CPU metrics shall be recorded
		EnableMemory bool   // EnableMemory whether memory metrics shall be recorded
		Prefix       string // Prefix is a custom prefix for metric names
	}

	memStats struct {
		memStats runtime.MemStats

		memAlloc   *metric.Int64GaugeEntry
		memTotal   *metric.Int64GaugeEntry
		memSys     *metric.Int64GaugeEntry
		memLookups *metric.Int64GaugeEntry
		memMalloc  *metric.Int64GaugeEntry
		memFrees   *metric.Int64GaugeEntry

		heapAlloc    *metric.Int64GaugeEntry
		heapSys      *metric.Int64GaugeEntry
		heapIdle     *metric.Int64GaugeEntry
		heapInuse    *metric.Int64GaugeEntry
		heapObjects  *metric.Int64GaugeEntry
		heapReleased *metric.Int64GaugeEntry

		stackInuse       *metric.Int64GaugeEntry
		stackSys         *metric.Int64GaugeEntry
		stackMSpanInuse  *metric.Int64GaugeEntry
		stackMSpanSys    *metric.Int64GaugeEntry
		stackMCacheInuse *metric.Int64GaugeEntry
		stackMCacheSys   *metric.Int64GaugeEntry
	}

	cpuStats struct {
		numGoroutines *metric.Int64GaugeEntry
		numCgoCalls   *metric.Int64GaugeEntry
	}
)

var _ metricproducer.Producer = (*Producer)(nil)

// NewProducer creates a new runtime metrics producer.
//
// Supply RunMetricOptions to configure the behavior of the Producer.
// An error might be returned, if creating metrics gauges fails.
func NewProducer(options RunMetricOptions) (*Producer, error) {
	collector := &Producer{options: options, reg: metric.NewRegistry()}
	var err error

	if options.EnableMemory {
		collector.memStats, err = newMemStats(collector)
		if err != nil {
			return nil, err
		}
	}

	if options.EnableCPU {
		collector.cpuStats, err = newCPUStats(collector)
		if err != nil {
			return nil, err
		}
	}

	return collector, err
}

// Read reads the current runtime metrics.
func (c *Producer) Read() []*metricdata.Metric {
	if c.memStats != nil {
		c.memStats.read()
	}

	if c.cpuStats != nil {
		c.cpuStats.read()
	}

	return c.reg.Read()
}

func newMemStats(producer *Producer) (*memStats, error) {
	var err error
	memStats := &memStats{}

	// General
	memStats.memAlloc, err = producer.createInt64GaugeEntry("mem_alloc", "Bytes of allocated heap objects", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.memTotal, err = producer.createInt64GaugeEntry("mem_total", "Cumulative bytes allocated for heap objects", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.memSys, err = producer.createInt64GaugeEntry("mem_sys", "Total bytes of memory obtained from the OS", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.memLookups, err = producer.createInt64GaugeEntry("mem_lookups", "Number of pointer lookups performed by the runtime", metricdata.UnitDimensionless)
	if err != nil {
		return nil, err
	}

	memStats.memMalloc, err = producer.createInt64GaugeEntry("mem_malloc", "Cumulative count of heap objects allocated", metricdata.UnitDimensionless)
	if err != nil {
		return nil, err
	}

	memStats.memFrees, err = producer.createInt64GaugeEntry("mem_frees", "Cumulative count of heap objects freed", metricdata.UnitDimensionless)
	if err != nil {
		return nil, err
	}

	// Heap
	memStats.heapAlloc, err = producer.createInt64GaugeEntry("heap_alloc", "Process heap allocation", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.heapSys, err = producer.createInt64GaugeEntry("heap_sys", "todo", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.heapIdle, err = producer.createInt64GaugeEntry("heap_idle", "todo", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.heapInuse, err = producer.createInt64GaugeEntry("heap_inuse", "todo", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.heapObjects, err = producer.createInt64GaugeEntry("heap_objects", "The number of objects allocated on the heap", metricdata.UnitDimensionless)
	if err != nil {
		return nil, err
	}

	memStats.heapReleased, err = producer.createInt64GaugeEntry("heap_release", "The number of objects released from the heap", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	// Stack
	memStats.stackInuse, err = producer.createInt64GaugeEntry("stack_inuse", "Bytes in stack spans", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.stackSys, err = producer.createInt64GaugeEntry("stack_sys", "The memory used by stack spans and OS thread stacks", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.stackMSpanInuse, err = producer.createInt64GaugeEntry("stack_mspan_inuse", "Bytes of allocated mspan structures", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.stackMSpanSys, err = producer.createInt64GaugeEntry("stack_mspan_sys", "Bytes of memory obtained from the OS for mspan structures", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.stackMCacheInuse, err = producer.createInt64GaugeEntry("stack_mcache_inuse", "Bytes of allocated mcache structures", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	memStats.stackMCacheSys, err = producer.createInt64GaugeEntry("stack_mcache_sys", "Bytes of memory obtained from the OS for mcache structures", metricdata.UnitBytes)
	if err != nil {
		return nil, err
	}

	return memStats, nil
}

func (m *memStats) read() {
	runtime.ReadMemStats(&m.memStats)

	m.memAlloc.Set(int64(m.memStats.Alloc))
	m.memTotal.Set(int64(m.memStats.TotalAlloc))
	m.memSys.Set(int64(m.memStats.Sys))
	m.memLookups.Set(int64(m.memStats.Lookups))
	m.memMalloc.Set(int64(m.memStats.Mallocs))
	m.memFrees.Set(int64(m.memStats.Frees))

	m.heapAlloc.Set(int64(m.memStats.HeapAlloc))
	m.heapSys.Set(int64(m.memStats.HeapSys))
	m.heapIdle.Set(int64(m.memStats.HeapIdle))
	m.heapInuse.Set(int64(m.memStats.HeapInuse))
	m.heapReleased.Set(int64(m.memStats.HeapReleased))
	m.heapObjects.Set(int64(m.memStats.HeapObjects))

	m.stackInuse.Set(int64(m.memStats.StackInuse))
	m.stackSys.Set(int64(m.memStats.StackSys))
	m.stackMSpanInuse.Set(int64(m.memStats.MSpanInuse))
	m.stackMSpanSys.Set(int64(m.memStats.MSpanSys))
	m.stackMCacheInuse.Set(int64(m.memStats.MCacheInuse))
	m.stackMCacheSys.Set(int64(m.memStats.MCacheSys))
}

func newCPUStats(collector *Producer) (*cpuStats, error) {
	cpuStats := &cpuStats{}
	var err error

	cpuStats.numGoroutines, err = collector.createInt64GaugeEntry("cpu_goroutines", "Number of goroutines that currently exist", metricdata.UnitDimensionless)
	if err != nil {
		return nil, err
	}

	cpuStats.numCgoCalls, err = collector.createInt64GaugeEntry("cpu_cgo_calls", "Number of cgo calls made by the current process", metricdata.UnitDimensionless)
	if err != nil {
		return nil, err
	}

	return cpuStats, nil
}

func (c *cpuStats) read() {
	c.numGoroutines.Set(int64(runtime.NumGoroutine()))
	c.numCgoCalls.Set(runtime.NumCgoCall())
}

func (c *Producer) createInt64GaugeEntry(name string, description string, unit metricdata.Unit) (*metric.Int64GaugeEntry, error) {
	if len(c.options.Prefix) > 0 {
		name = c.options.Prefix + name
	}

	gauge, err := c.reg.AddInt64Gauge(
		name,
		metric.WithDescription(description),
		metric.WithUnit(unit))
	if err != nil {
		return nil, errors.New("error creating gauge for " + name + ": " + err.Error())
	}

	entry, err := gauge.GetEntry()
	if err != nil {
		return nil, errors.New("error getting gauge entry for " + name + ": " + err.Error())
	}

	return entry, nil
}
