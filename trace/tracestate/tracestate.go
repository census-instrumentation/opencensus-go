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

package tracestate

import (
	"fmt"
	"regexp"
)

const (
	keyMaxSize       = 256
	valueMaxSize     = 256
	maxKeyValuePairs = 32
)

const (
	keyWithoutVendorFormat = `[a-z][_0-9a-z\-\*\/]{0,255}`
	keyWithVendorFormat    = `[a-z][_0-9a-z\-\*\/]{0,240}@[a-z][_0-9a-z\-\*\/]{0,13}`
	keyFormat              = `(` + keyWithoutVendorFormat + `)|(` + keyWithVendorFormat + `)`
	valueFormat            = `[\x20-\x2b\x2d-\x3c\x3e-\x7e]{0,255}[\x21-\x2b\x2d-\x3c\x3e-\x7e]`
)

var keyValidationRegExp = regexp.MustCompile(`^(` + keyFormat + `)$`)
var valueValidationRegExp = regexp.MustCompile(`^(` + valueFormat + `)$`)

// Tracestate represents tracing-system specific context in a list of key-value pairs. Tracestate allows different
// vendors propagate additional information and inter-operate with their legacy Id formats.
type Tracestate struct {
	entries []Entry
}

// Entry represents one key-value pair in a list of key-value pair of Tracestate.
// Key is an opaque string up to 256 characters printable. It MUST begin with a lowercase letter,
// and can only contain lowercase letters a-z, digits 0-9, underscores _, dashes -, asterisks *, and
// forward slashes /.
//
// Value is an opaque string up to 256 characters printable ASCII RFC0020 characters (i.e., the
// range 0x20 to 0x7E) except comma , and =.
type Entry struct {
	Key   string
	Value string
}

// Get retrieves value for a given key from Tracestate ts.
// If the key is not found then false is returned with the value "".
// If the key is found then true is returned with its value.
func (ts *Tracestate) Get(key string) (string, bool) {
	if len(ts.entries) == 0 {
		return "", false
	}
	for _, entry := range ts.entries {
		if entry.Key == key {
			return entry.Value, true
		}
	}
	return "", false
}

// Entries returns a slice of Entry.
func (ts *Tracestate) Entries() []Entry {
	return ts.entries
}

func (ts *Tracestate) remove(key string) *Entry {
	for index, entry := range ts.entries {
		if entry.Key == key {
			ts.entries = append(ts.entries[:index], ts.entries[index+1:]...)
			return &entry
		}
	}
	return nil
}

func (ts *Tracestate) add(entry Entry) error {
	ts.remove(entry.Key)
	if len(ts.entries) >= maxKeyValuePairs {
		return fmt.Errorf("reached maximum key/value pairs limit of %d", maxKeyValuePairs)
	}
	ts.entries = append([]Entry{entry}, ts.entries...)
	return nil
}

func isValid(entry Entry) bool {
	return keyValidationRegExp.MatchString(entry.Key) &&
		valueValidationRegExp.MatchString(entry.Value)
}

func containsDuplicateKey(entries []Entry) (string, bool) {
	keyMap := make(map[string]int)
	for _, entry := range entries {
		if _, ok := keyMap[entry.Key]; ok {
			return entry.Key, true
		}
		keyMap[entry.Key] = 1
	}
	return "", false
}

func areEntriesValid(entries []Entry) (*Entry, bool) {
	for _, entry := range entries {
		if !isValid(entry) {
			return &entry, false
		}
	}
	return nil, true
}

// NewFromEntries creates a Tracestate object from an array of key-value pair.
// nil is returned with an error if
//  1. the len of the entries > maxKeyValuePairs
//  2. the entries contain duplicate keys
//  3. one or more entries are invalid.
func NewFromEntries(entries []Entry) (*Tracestate, error) {
	if len(entries) == 0 || len(entries) > maxKeyValuePairs {
		return nil, fmt.Errorf("number of entries(%d) is larger than max (%d)", len(entries), maxKeyValuePairs)
	}
	if key, duplicate := containsDuplicateKey(entries); duplicate {
		return nil, fmt.Errorf("contains duplicate keys (%s)", key)
	}

	if entry, ok := areEntriesValid(entries); !ok {
		return nil, fmt.Errorf("key-value pair {%s, %s} is invalid", entry.Key, entry.Value)
	}
	tracestate := Tracestate{}

	if entries != nil {
		tracestate.entries = append([]Entry{}, entries...)
	}

	return &tracestate, nil
}

// NewFromParent creates a Tracestate object from a parent and an entry (key-value pair).
// Entries from the parent are copied if present and the entry passed to this function
// is inserted in front of the list. If there exists any entry with the same key it is
// removed. See add func.
// An error is returned with nil Tracestate if
//  1. the entry is invalid.
//  2. the number of entries exceeds maxKeyValuePairs.
func NewFromParent(parent *Tracestate, entry Entry) (*Tracestate, error) {
	if !isValid(entry) {
		return nil, fmt.Errorf("key-value pair {%s, %s} is invalid", entry.Key, entry.Value)
	}

	tracestate := Tracestate{}

	if parent != nil && len(parent.entries) > 0 {
		tracestate.entries = append([]Entry{}, parent.entries...)
	}

	err := tracestate.add(entry)
	if err != nil {
		return nil, err
	}
	return &tracestate, nil
}
