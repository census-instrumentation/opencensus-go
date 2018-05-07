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

// Key represents a tag key. Tags are propagated along RPC boundaries,
// unless created by NewLocalKey.
type Key struct {
	name       string
	propagated bool
}

// NewKey creates a Key with the given name. The returned Key will be propagated
// on RPC boundaries by default. Use NewLocalKey if you only need the tags with
// this Key to be available in the current process.
func NewKey(name string) (Key, error) {
	if !checkKeyName(name) {
		return Key{}, errInvalidKeyName
	}
	return Key{name: name, propagated: true}, nil
}

// NewLocalKey creates a new Key local to the current process. Tags with local
// keys will not be propagated in outbound RPCs.
func NewLocalKey(name string) (Key, error) {
	if !checkKeyName(name) {
		return Key{}, errInvalidKeyName
	}
	return Key{name: name, propagated: false}, nil
}

// Name returns the name of the key.
func (k Key) Name() string {
	return k.name
}
