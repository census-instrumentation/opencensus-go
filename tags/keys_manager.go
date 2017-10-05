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

package tags

import (
	"fmt"
	"sync"
)

const (
	// maxKeyLength is the maximum value able to be encoded in a 2-byte varint.
	maxKeyLength = 1<<14 - 1

	// validKeys are restricted to US-ASCII subset (range 0x20 (' ') to 0x7e ('~')).
	validKeysMin = 0x20
	validKeysMax = 0x7e
)

type keysManager struct {
	*sync.Mutex
	keys      map[string]Key
	nextKeyID uint16
}

func newKeysManager() *keysManager {
	return &keysManager{
		keys:  make(map[string]Key),
		Mutex: &sync.Mutex{},
	}
}

// CreateKeyString creates or retrieves a key of type keyString with name/ID
// set to the input argument name. Returns an error if a key with the same name
// exists and is of a different type.
func (km *keysManager) createKeyString(name string) (*KeyString, error) {
	if !validateKeyName(name) {
		return nil, fmt.Errorf("key name %v is invalid", name)
	}
	km.Lock()
	defer km.Unlock()

	k, ok := km.keys[name]
	if ok {
		ks, ok := k.(*KeyString)
		if !ok {
			return nil, fmt.Errorf("key with name %v cannot be created/retrieved as type *keyString. It was already registered as type %T", name, k)
		}
		return ks, nil
	}

	ks := &KeyString{
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

func validateKeyName(name string) bool {
	if len(name) >= maxKeyLength {
		return false
	}
	for _, c := range name {
		if (c < validKeysMin) || (c > validKeysMax) {
			return false
		}
	}
	return true
}

func init() {
	km := newKeysManager()
	createKeyString = km.createKeyString
}
