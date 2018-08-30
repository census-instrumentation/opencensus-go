// Copyright 2018, OpenCensus Authors
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

package trace

import (
	"fmt"
	"regexp"
)

const (
	KEY_WITHOUT_VENDOR_FORMAT = `[a-z][_0-9a-z\-\*\/]{0,255}`
	KEY_WITH_VENDOR_FORMAT = `[a-z][_0-9a-z\-\*\/]{0,240}@[a-z][_0-9a-z\-\*\/]{0,13}`
	KEY_FORMAT = `(` + KEY_WITHOUT_VENDOR_FORMAT + `)|(` + KEY_WITH_VENDOR_FORMAT + `)`
	VALUE_FORMAT = `[\x20-\x2b\x2d-\x3c\x3e-\x7e]{0,255}[\x21-\x2b\x2d-\x3c\x3e-\x7e]`
)

var KEY_VALIDATION_RE = regexp.MustCompile(`^(` + KEY_FORMAT + `)$`)
var VALUE_VALIDATION_RE = regexp.MustCompile(`^(` + VALUE_FORMAT + `)$`)

type TracestateEntry struct {
	key string
	value string
}

func (ts *TracestateEntry) IsValid() bool {
	return KEY_VALIDATION_RE.MatchString(ts.key) &&
		VALUE_VALIDATION_RE.MatchString(ts.value)
}

// TraceState is a tracing-system specific context in a list of key-value pairs.
// TraceState allows different vendors propagate additional information and
// inter-operate with their legacy Id formats.
type Tracestate struct {
	entries []TracestateEntry
}

func (ts *Tracestate) IsValid() bool {
	if len(ts.entries) == 0 || len(ts.entries) > 32 {
		return false
	}
	for _, entry := range ts.entries {
		if !entry.IsValid() {
			return false
		}
	}
	return true
}

func (ts *Tracestate) Fork() *Tracestate {
	retval := Tracestate{entries: ts.entries}
	return &retval
}

func (ts *Tracestate) Get(key string) string {
	for _, entry := range ts.entries {
		if entry.key == key {
			return entry.value
		}
	}
	return ""
}

func (ts *Tracestate) Remove(key string) string {
	return ts.Set(key, "")
}

func (ts *Tracestate) Set(key string, value string) string {
	retval := ""
	newEntry := TracestateEntry{key: key, value: value}
	if !newEntry.IsValid() && value != "" {
		return retval
	}
	for index, entry := range ts.entries {
		if entry.key == key {
			ts.entries = append(ts.entries[:index], ts.entries[index + 1:]...)
			retval = entry.value
			break
		}
	}
	if value != "" {
		ts.entries = append([]TracestateEntry{newEntry}, ts.entries...)
	}
	return retval
}

func (ts *Tracestate) String() string {
	return fmt.Sprintf("tracestate%s", ts.entries)
}
