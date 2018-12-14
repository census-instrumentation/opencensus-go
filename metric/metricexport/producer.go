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

package metricexport

import (
	"go.opencensus.io/metric/metricdata"
	"sync"
	"sync/atomic"
)

// Producer is a source of metrics.
type Producer interface {
	// Read should return the current values of all metrics supported by this
	// metric provider.
	// The returned metrics should be unique for each combination of name and
	// resource.
	Read() []*metricdata.Metric
}

// Registry maintains a set of metric producers for exporting. Most users will
// rely on the DefaultRegistry.
type Registry struct {
	mu    sync.RWMutex
	state atomic.Value
}

type registryState struct {
	producers map[Producer]struct{}
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	m := &Registry{}
	m.state.Store(&registryState{
		producers: make(map[Producer]struct{}),
	})
	return m
}

// Read returns all the metrics from all the metric produces in this registry.
func (m *Registry) ReadAll() []*metricdata.Metric {
	s := m.state.Load().(*registryState)
	ms := make([]*metricdata.Metric, 0, len(s.producers))
	for p := range s.producers {
		ms = append(ms, p.Read()...)
	}
	return ms
}

// AddProducer adds a producer to this registry.
func (m *Registry) AddProducer(p Producer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	newState := &registryState{
		make(map[Producer]struct{}),
	}
	state := m.state.Load().(*registryState)
	for producer := range state.producers {
		newState.producers[producer] = struct{}{}
	}
	newState.producers[p] = struct{}{}
	m.state.Store(newState)
}

// RemoveProducer removes the given producer from this registry.
func (m *Registry) RemoveProducer(p Producer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	newState := &registryState{
		make(map[Producer]struct{}),
	}
	state := m.state.Load().(*registryState)
	for producer := range state.producers {
		newState.producers[producer] = struct{}{}
	}
	delete(newState.producers, p)
	m.state.Store(newState)
}

var defaultReg = NewRegistry()

// DefaultRegistry returns the default, global metric registry for the current
// process.
// Most applications will rely on this registry but libraries should not assume
// the default registry is used.
func DefaultRegistry() *Registry {
	return defaultReg
}
