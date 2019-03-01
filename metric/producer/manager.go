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
	producers []Producer
}

var prodMgr *manager
var once sync.Once

func getManager() *manager {
	once.Do(func() {
		prodMgr = &manager{}
	})
	return prodMgr
}

// Add adds the producer to the manager if it is not already present.
func Add(producer Producer) {
	pm := getManager()
	pm.add(producer)
}

// Delete deletes the producer from the manager if it is present.
func Delete(producer Producer) {
	pm := getManager()
	pm.delete(producer)
}

// GetAll returns all producer registered with the manager. It is typically
// used by exporter to read metrics from producers.
func GetAll() []Producer {
	pm := getManager()
	return pm.producers
}

func (pm *manager) add(producer Producer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, prod := range pm.producers {
		if producer == prod {
			return
		}
	}
	pm.producers = append(pm.producers, producer)
}

func (pm *manager) delete(producer Producer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for index, prod := range pm.producers {
		if producer == prod {
			pm.producers = append(pm.producers[:index], pm.producers[index+1:]...)
		}
	}
}
