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

// Key represents a tag key.
type Key struct {
	name string
}

// NewKey creates or retrieves a string key identified by name.
// Calling NewKey consequently with the same name returns the same key.
func NewKey(name string) (Key, error) {
	if !checkKeyName(name) {
		return Key{}, errInvalidKeyName
	}
	return Key{name: name}, nil
}

// Name returns the name of the key.
func (k Key) Name() string {
	return k.name
}

// Extract produces a dimension of the key name with a value associated with
// this Key in the given Map. Internal use only.
func (k Key) Extract(m *Map) (val string, ok bool) {
	return m.Value(k)
}

// AliasedKey represents a Key that should be exported under a different alias.
type AliasedKey struct {
	k     Key
	alias Key
}

// As produces a AliasedKey representing this key aliased.
func (k Key) As(alias Key) AliasedKey {
	return AliasedKey{k: k, alias: alias}
}

// Extract produces a dimension of the alias with a value associated with the
// underlying key. Internal use only.
func (r AliasedKey) Extract(m *Map) (val string, ok bool) {
	return r.k.Extract(m)
}

// Name returns the alias of this key.
func (r AliasedKey) Name() string {
	return r.alias.name
}

// TODO(ramonza): add Key.WhitelistValues(...)
