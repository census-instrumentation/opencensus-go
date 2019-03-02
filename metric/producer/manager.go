// Copyright 2019, OpenCensus Authors
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

package producer

import (
	"sync"
)

type manager struct {
	mu        sync.RWMutex
	producers map[Producer]struct{}
}

var prodMgr *manager
var once sync.Once

func getManager() *manager {
	once.Do(func() {
		prodMgr = &manager{}
		prodMgr.producers = make(map[Producer]struct{})
	})
	return prodMgr
}

// Add adds the producer to the manager if it is not already present.
// The manager maintains the list of active producers. It provides
// this list to a reader to read metrics from each producer and then export.
func Add(producer Producer) {
	if producer == nil {
		return
	}
	pm := getManager()
	pm.add(producer)
}

// Delete deletes the producer from the manager if it is present.
func Delete(producer Producer) {
	if producer == nil {
		return
	}
	pm := getManager()
	pm.delete(producer)
}

// GetAll returns a slice of all producer currently registered with
// the manager. For each call it generates a new slice. The slice
// should not be cached as registration may change at any time. It is
// typically called periodically by exporter to read metrics from
// the producers.
func GetAll() []Producer {
	pm := getManager()
	return pm.getAll()
}

func (pm *manager) getAll() []Producer {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	producers := make([]Producer, len(pm.producers))
	i := 0
	for producer := range pm.producers {
		producers[i] = producer
		i++
	}
	return producers
}

func (pm *manager) add(producer Producer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.producers[producer] = struct{}{}
}

func (pm *manager) delete(producer Producer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.producers, producer)
}
