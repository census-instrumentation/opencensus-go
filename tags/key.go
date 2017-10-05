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

import "fmt"

var keys []Key

// Key represents a tag key.
type Key interface {
	Name() string
	ID() uint16
	ValueAsString(b []byte) string
}

// KeyString is a Key and represents string keys.
type KeyString struct {
	name string
	id   uint16
}

// TODO(jbd): Raname KeyString to StringKey?
// TODO(jbd): What is ID? Should we export an accessor for it?

// Name returns the name of the key.
func (k *KeyString) Name() string {
	return k.name
}

// ID returns the ID of the key.
func (k *KeyString) ID() uint16 {
	return k.id
}

// ValueAsString encodes the given values represented in binary to string.
func (k *KeyString) ValueAsString(v []byte) string {
	return string(v)
}

func (k *KeyString) String() string {
	return fmt.Sprintf("%v", k.Name())
}

// KeyStringByName creates or retrieves a *KeyString identified by name.
func KeyStringByName(name string) (*KeyString, error) {
	return createKeyString(name)
}

var createKeyString func(name string) (*KeyString, error)
