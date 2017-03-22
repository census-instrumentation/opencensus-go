// Copyright 2017 Google Inc.
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

package tagging

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

// KeysManager is the interface that a keys manager implementation needs to
// satisfy. The keys manager is invoked to create/retrieve a key given its
// name/ID. It ensures that keys have unique names/IDs.
type KeysManager interface {
	CreateKeyStringUTF8(name string) (KeyStringUTF8, error)
	CreateKeyInt64(name string) (KeyInt64, error)
	CreateKeyBool(name string) (KeyBool, error)
	CreateKeyBytes(name string) (KeyBytes, error)
	Count() int
	Clear()
}

type keysManager struct {
	*sync.Mutex
	keys      map[string]Key
	nextKeyID int32
}

var defaultKeysManager = &keysManager{
	keys:  make(map[string]Key),
	Mutex: &sync.Mutex{},
}

// DefaultKeyManager returns the singleton defaultKeysManager. Because it is a
// singleton, the defaultKeysManager can easily ensure the keys have unique
// names/IDs.
func DefaultKeyManager() KeysManager {
	return defaultKeysManager
}

// CreateKeyString creates or retrieves a key of type keyString with name/ID
// set to the input argument name. Returns an error if a key with the same name
// exists and is of a different type.
func (km *keysManager) CreateKeyStringUTF8(name string) (KeyStringUTF8, error) {
	if !validateKeyName(name) {
		return nil, fmt.Errorf("key name %v is invalid", name)
	}
	km.Lock()
	defer km.Unlock()

	k, ok := km.keys[name]
	if ok {
		ks, ok := k.(*keyStringUTF8)
		if !ok {
			return nil, fmt.Errorf("key with name %v cannot be created/retrieved as type *keyString. It was already registered as type %T", name, k)
		}
		return ks, nil
	}

	ks := &keyStringUTF8{
		name: name,
		id:   km.nextKeyID,
	}
	km.nextKeyID++
	km.keys[name] = ks
	return ks, nil
}

// CreateKeyBytes creates or retrieves a key of type keyBytes with name/ID set
// to the input argument name. Returns an error if a key with the same name
// exists and is of a different type.
func (km *keysManager) CreateKeyBytes(name string) (KeyBytes, error) {
	if !validateKeyName(name) {
		return nil, fmt.Errorf("key name %v is invalid", name)
	}
	km.Lock()
	defer km.Unlock()

	k, ok := km.keys[name]
	if ok {
		ks, ok := k.(*keyBytes)
		if !ok {
			return nil, fmt.Errorf("key with name %v cannot be created/retrieved as type *keyBytes. It was already registered as type %T", name, k)
		}
		return ks, nil
	}

	ks := &keyBytes{
		name: name,
		id:   km.nextKeyID,
	}
	km.nextKeyID++
	km.keys[name] = ks
	return ks, nil
}

// CreateKeyBool creates or retrieves a key of type keyBool with name/ID set to
// the input argument name. Returns an error if a key with the same name exists
// and is of a different type.
func (km *keysManager) CreateKeyBool(name string) (KeyBool, error) {
	if !validateKeyName(name) {
		return nil, fmt.Errorf("key name %v is invalid", name)
	}
	km.Lock()
	defer km.Unlock()
	k, ok := km.keys[name]
	if ok {
		kb, ok := k.(*keyBool)
		if !ok {
			return nil, fmt.Errorf("key with name %v cannot be created/retrieved as type *keyBool. It was already registered as type %T", name, k)
		}
		return kb, nil
	}

	kb := &keyBool{
		name: name,
		id:   km.nextKeyID,
	}
	km.nextKeyID++
	km.keys[name] = kb
	return kb, nil
}

// CreateKeyInt64 creates or retrieves a key of type keyInt64 with name/ID set
// to the input argument name. Returns an error if a key with the same name
// exists and is of a different type.
func (km *keysManager) CreateKeyInt64(name string) (KeyInt64, error) {
	if !validateKeyName(name) {
		return nil, fmt.Errorf("key name %v is invalid", name)
	}
	km.Lock()
	defer km.Unlock()
	k, ok := km.keys[name]
	if ok {
		ki, ok := k.(*keyInt64)
		if !ok {
			return nil, fmt.Errorf("key with name %v cannot be created/retrieved as type *keyInt64. It was already registered as type %T", name, k)
		}
		return ki, nil
	}

	ki := &keyInt64{
		name: name,
		id:   km.nextKeyID,
	}
	km.nextKeyID++
	km.keys[name] = ki
	return ki, nil
}

func (km *keysManager) Count() int {
	return len(km.keys)
}

func (km *keysManager) Clear() {
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
