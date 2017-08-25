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

// Key is the interface for all key types.
type Key interface {
	Name() string
	ID() uint16
}

// KeyString implements the Key interface and is used to represent keys for
// which the value type is a string.
type KeyString struct {
	name string
	id   uint16
}

// Name returns the unique name of a key.
func (k *KeyString) Name() string {
	return k.name
}

// ID returns the id of a key inside hte process.
func (k *KeyString) ID() uint16 {
	return k.id
}

func (k *KeyString) String() string {
	return fmt.Sprintf("%v", k.Name())
}

// CreateKeyString creates/retrieves the *KeyString identified by name.
var CreateKeyString func(name string) (*KeyString, error)
