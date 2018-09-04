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
	"errors"
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

// Tracestate represent tracing-system specific context in a list of key-value pairs. Tracestate allows different
// vendors propagate additional information and inter-operate with their legacy Id formats.
type Tracestate struct {
	tracestateEntries []*TracestateEntry
}

// TracestateEntry represent one key-value pair in a list of key-value pair of Tracestate.
// Key is an opaque string up to 256 characters printable. It MUST begin with a lowercase letter,
// and can only contain lowercase letters a-z, digits 0-9, underscores _, dashes -, asterisks *, and
// forward slashes /.
//
// Value is an opaque string up to 256 characters printable ASCII RFC0020 characters (i.e., the
// range 0x20 to 0x7E) except comma , and =.
type TracestateEntry struct {
	key   string
	value string
}

// Key returns the key of TracestateEntry
func (te *TracestateEntry) Key() string {
	return te.key
}

// Value returns the value of TracestateEntry
func (te *TracestateEntry) Value() string {
	return te.value
}

// Get retrieves value for a given key from Tracestate ts.
// If the key is not found then false is returned with the value "" is returned.
// If the key is found then true is returned with its value.
func (ts *Tracestate) Get(key string) (string, bool) {
	if len(ts.tracestateEntries) == 0 {
		return "", false
	}
	for _, entry := range ts.tracestateEntries {
		if entry.key == key {
			return entry.value, true
		}
	}
	return "", false
}

// TraceEntries returns a slice of TracestateEntry.
func (ts *Tracestate) TraceEntries() []*TracestateEntry {
	return ts.tracestateEntries
}

func (ts *Tracestate) remove(key string) {
	for index, entry := range ts.tracestateEntries {
		if entry.key == key {
			ts.tracestateEntries = append(ts.tracestateEntries[:index], ts.tracestateEntries[index+1:]...)
			break
		}
	}
}

func (ts *Tracestate) add(entry *TracestateEntry) error {
	ts.remove(entry.Key())
	if len(ts.tracestateEntries) >= maxKeyValuePairs {
		return fmt.Errorf("Set failed: reached maximum key/value pairs limit of %d", maxKeyValuePairs)
	}
	ts.tracestateEntries = append([]*TracestateEntry{entry}, ts.tracestateEntries...)
	return nil
}

func isValid(key, value string) bool {
	return keyValidationRegExp.MatchString(key) &&
		valueValidationRegExp.MatchString(value)
}

// NewTracestateEntry creates a TracestateEntry object with given key and value.
// It returns error if either key or value is invalid.
func NewTracestateEntry(key, value string) (*TracestateEntry, error) {
	if isValid(key, value) {
		return &TracestateEntry{
			key:   key,
			value: value,
		}, nil
	}
	return nil, errors.New("invalid parameters")
}

func containsDuplicateKey(entries []*TracestateEntry) (string, bool) {
	keyMap := make(map[string]int)
	for _, entry := range entries {
		if _, ok := keyMap[entry.Key()]; ok {
			return entry.Key(), true
		}
		keyMap[entry.Key()] = 1
	}
	return "", false
}

// NewFromEntryArray creates a Tracestate object from an array of key-value pair.
// nil is returned with with an error if
//  1. If the len of the entries > maxKeyValuePairs
//  2. If the entries contain duplicate keys
func NewFromEntryArray(entries []*TracestateEntry) (*Tracestate, error) {

	if entries == nil {
		return nil, errors.New("Invalid parameter")
	}
	if len(entries) == 0 || len(entries) > maxKeyValuePairs {
		return nil, fmt.Errorf("Invalid number of tracestateEntry (%d)", len(entries))
	}
	if key, duplicate := containsDuplicateKey(entries); duplicate {
		return nil, fmt.Errorf("Contains duplicate keys (%s)", key)
	}

	tracestate := Tracestate{}

	if entries != nil {
		tracestate.tracestateEntries = append([]*TracestateEntry{}, entries...)
	}

	return &tracestate, nil
}

// NewFromParent creates a Tracestate object and adds a key-value pair to the list.
// If a non-empty parent is passed then key/value pair from the parent is copied
// to a newly created Tracestate object.
// If the key already exists in the parent then its value is replaced with the
// value passed to this function. The key is also moved to the front of
// the list. See add func.
func NewFromParent(parent *Tracestate, key, value string) (*Tracestate, error) {

	tracestate := Tracestate{}

	if parent != nil && len(parent.tracestateEntries) > 0 {
		tracestate.tracestateEntries = append([]*TracestateEntry{}, parent.tracestateEntries...)
	}

	entry, err := NewTracestateEntry(key, value)
	if err != nil {
		return nil, err
	}

	err = tracestate.add(entry)
	if err != nil {
		return nil, err
	}
	return &tracestate, nil
}
