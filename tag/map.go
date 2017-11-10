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
	"bytes"
	"fmt"
	"sort"
)

// Tag is a key value pair that can be propagated on wire.
type Tag struct {
	Key   Key
	Value string
}

// Map is a map of tags. Use NewMap to build tag maps.
type Map struct {
	m map[Key]string
}

func (m *Map) String() string {
	var keys []Key
	for k := range m.m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Name() < keys[j].Name() })

	var buffer bytes.Buffer
	buffer.WriteString("{ ")
	for _, k := range keys {
		buffer.WriteString(fmt.Sprintf("{%v %v}", k.name, m.m[k]))
	}
	buffer.WriteString(" }")
	return buffer.String()
}

func (m *Map) insert(k Key, v string) {
	if _, ok := m.m[k]; ok {
		return
	}
	m.m[k] = v
}

func (m *Map) update(k Key, v string) {
	if _, ok := m.m[k]; ok {
		m.m[k] = v
	}
}

func (m *Map) upsert(k Key, v string) {
	m.m[k] = v
}

func (m *Map) delete(k Key) {
	delete(m.m, k)
}

func newMap(sizeHint int) *Map {
	return &Map{m: make(map[Key]string, sizeHint)}
}
