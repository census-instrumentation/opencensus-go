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
	"container/list"
	"errors"
	"fmt"
)

const (
	// KeyMaxSize is the maximum characters allowed in the key.
	KeyMaxSize = 256

	// ValueMaxSize is the maximum characters allowed in the value.
	ValueMaxSize = 256

	// MaxKeyValuePairs is the maximum number of key-value pairs allowed in the tracestate.
	MaxKeyValuePairs = 32
)

// Tracestate represent tracing-system specific context in a list of key-value pairs. Tracestate allows different
// vendors propagate additional information and inter-operate with their legacy Id formats.
type Tracestate struct {
	stateList *list.List
}

// Entry represent one key-value pair in a list of key-value pair of Tracestate.
// Key is an opaque string up to 256 characters printable. It MUST begin with a lowercase letter,
// and can only contain lowercase letters a-z, digits 0-9, underscores _, dashes -, asterisks *, and
// forward slashes /.
//
// Value is an opaque string up to 256 characters printable ASCII RFC0020 characters (i.e., the
// range 0x20 to 0x7E) except comma , and =.
type Entry struct {
	key   string
	value string
}

// Get retrieves value for a given key from Tracestate ts.
// If the key is not found then the value "" is returned.
// If the key is found then its value is returned.
func (ts *Tracestate) Get(key string) string {
	if ts.stateList == nil {
		return ""
	}
	var kv *Entry
	for e := ts.stateList.Front(); e != nil; e = e.Next() {
		kv = e.Value.(*Entry)
		if kv.key == key {
			return kv.value
		}
	}
	return ""
}

// GetEntries retrieves all key/value Entry from Tracestate ts.
func (ts *Tracestate) GetEntries() *list.List {
	return ts.stateList
}

// Remove removes the key/value Entry the Tracestate ts.
func (ts *Tracestate) Remove(key string) {
	if ts.stateList == nil {
		return
	}
	elemWithKey := ts.findElem(key)
	if elemWithKey != nil {
		ts.stateList.Remove(elemWithKey)
	}
}

// Set inserts a key/value Entry in the front if there isn't
// and existing Entry with the same key. If there exists an Entry
// with the same key then it is removed before inserting the new Entry.
// An error is returned if either the key or the value is invalid.
// An error is returned if the Tracestate object is invalid.
func (ts *Tracestate) Set(key, value string) error {
	if ts.stateList == nil {
		ts.stateList = list.New()
	}
	entry, err := createEntry(key, value)
	if err != nil {
		return err
	}
	ts.Remove(key)
	if ts.stateList.Len() >= MaxKeyValuePairs {
		return fmt.Errorf("Set failed: reached maximum key/value pairs limit of %d", MaxKeyValuePairs)
	}
	ts.stateList.PushFront(entry)
	return nil
}

func (ts *Tracestate) findElem(key string) *list.Element {
	var kv *Entry
	for e := ts.stateList.Front(); e != nil; e = e.Next() {
		kv = e.Value.(*Entry)
		if kv.key == key {
			return e
		}
	}
	return nil
}

func validateKey(key string) bool {
	keyRune := []rune(key)

	if len(key) > KeyMaxSize || key == "" || keyRune[0] < 'a' || keyRune[0] > 'z' {
		return false
	}
	for i := 1; i < len(key); i++ {
		c := keyRune[i]
		if !(c >= 'a' && c <= 'z') &&
			!(c >= '0' && c <= '9') &&
			c != '_' &&
			c != '-' &&
			c != '*' &&
			c != '/' {
			return false
		}
	}
	return true
}

func validateVal(value string) bool {
	valueRune := []rune(value)
	if value == "" || len(value) > ValueMaxSize || valueRune[len(value)-1] == ' ' /* '\u0020' */ {
		return false
	}
	for i := 0; i < len(value); i++ {
		c := valueRune[i]
		if c == ',' || c == '=' || c < ' ' /* '\u0020' */ || c > '~' /* '\u007E' */ {
			return false
		}
	}
	return true
}

func createEntry(key, value string) (*Entry, error) {
	if validateKey(key) && validateVal(value) {
		return &Entry{
			key:   key,
			value: value,
		}, nil
	}
	return nil, errors.New("invalid parameters")
}

// CreateTracestate creates a Tracestate object and adds a key-value pair to the list.
// If a non-empty parent is passed then key/value pair from the parent is copied
// to a newly created Tracestate object.
// If the key already exists in the parent then its value is replaced with the
// value passed to this function. The key is also moved to the front of
// the list. See Set func.
func CreateTracestate(parent *Tracestate, key, value string) (*Tracestate, error) {

	tracestate := Tracestate{
		stateList: list.New(),
	}

	// Insert all entries from the
	if parent != nil && parent.stateList != nil {
		tracestate.stateList.PushFrontList(parent.stateList)
	}

	err := tracestate.Set(key, value)
	if err != nil {
		return nil, err
	}
	return &tracestate, nil
}
