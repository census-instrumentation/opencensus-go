// Copyright 2017, OpenCensus Authors
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
//

package tag

import (
	"sync"
)

var km = newKeysManager()

type keysManager struct {
	sync.Mutex
	keys      map[string]Key
	nextKeyID uint16
}

func newKeysManager() *keysManager {
	return &keysManager{
		keys: make(map[string]Key),
	}
}

// newStringKey creates or retrieves a key of type keyString with name/ID
// set to the input argument name. Returns an error if a key with the same name
// exists and is of a different type.
func (km *keysManager) newStringKey(name string) (Key, error) {
	km.Lock()
	defer km.Unlock()

	k, ok := km.keys[name]
	if ok {
		return k, nil
	}

	ks := Key{
		name: name,
		id:   km.nextKeyID,
	}
	km.nextKeyID++
	km.keys[name] = ks
	return ks, nil
}

func (km *keysManager) count() int {
	km.Lock()
	defer km.Unlock()
	return len(km.keys)
}

func (km *keysManager) clear() {
	km.Lock()
	defer km.Unlock()
	for k := range km.keys {
		delete(km.keys, k)
	}
}
