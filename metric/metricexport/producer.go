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
	"sync"
)

// Producer is a source of metrics.
type Producer interface {
	// Read should return the current values of all metrics supported by this
	// metric provider.
	// The returned metrics should be unique for each combination of name and
	// resource.
	Read() []*Metric
}

// Registry maintains a set of metric producers for exporting. Most users will
// rely on the DefaultRegistry.
type Registry struct {
	mu      sync.RWMutex
	sources map[*uintptr]Producer
	ind     uint64
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	m := &Registry{
		sources: make(map[*uintptr]Producer),
		ind:     0,
	}
	return m
}

// Read returns all the metrics from all the metric produces in this registry.
func (m *Registry) ReadAll() []*Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ms := make([]*Metric, 0, len(m.sources))
	for _, s := range m.sources {
		ms = append(ms, s.Read()...)
	}
	return ms
}

// AddProducer adds a producer to this registry.
func (m *Registry) AddProducer(source Producer) (remove func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	tok := new(uintptr)
	m.sources[tok] = source
	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.sources, tok)
	}
}

var defaultReg = NewRegistry()

// DefaultRegistry returns the default, global metric registry for the current
// process.
// Most applications will rely on this registry but libraries should not assume
// the default registry is used.
func DefaultRegistry() *Registry {
	return defaultReg
}
