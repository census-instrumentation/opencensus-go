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
	"testing"
)

func init() {
}

func checkFront(t *testing.T, tracestate *Tracestate, key, testname string) {
	entries := tracestate.TraceEntries()
	entry := entries[0]
	got := entry.key
	if got != key {
		t.Errorf("Test:%s: key of the first entry in the list: got %q want %q", testname, got, key)
	}
}

func checkBack(t *testing.T, tracestate *Tracestate, key, testname string) {
	entries := tracestate.TraceEntries()
	entry := entries[len(entries)-1]
	got := entry.key
	if got != key {
		t.Errorf("Test:%s: key of the last entry in the list: got %q want %q", testname, got, key)
	}
}

func checkSize(t *testing.T, tracestate *Tracestate, size int, testname string) {
	tracestateEntries := tracestate.TraceEntries()
	gotSize := len(tracestateEntries)
	if gotSize != size {
		t.Errorf("Test:%s: size of the list: got %q want %q", testname, gotSize, size)
	}
}

func checkKeyValue(t *testing.T, tracestate *Tracestate, key, value, testname string) {
	wantOk := true
	if value == "" {
		wantOk = false
	}
	got, ok := tracestate.Get(key)
	if wantOk != ok || got != value {
		t.Errorf("Test:%s: Get value for key=%s failed: got %q want %q", testname, key, got, value)
	}
}

func checkError(t *testing.T, tracestate *Tracestate, err error, testname, msg string) {
	if err != nil {
		t.Errorf("Test:%s: %s: tracestate=%v, error= %v", testname, msg, tracestate, err)
	}
}

func expectError(t *testing.T, tracestate *Tracestate, err error, testname, msg string) {
	if err == nil {
		t.Errorf("Test:%s: %s: tracestate=%v, error=%v", testname, msg, tracestate, err)
	}
}

func TestCreateWithNullParent(t *testing.T) {
	key1, value1 := "hello", "world"
	testname := "TestCreateWithNullParent"

	tracestate, err := NewFromParent(nil, key1, value1)
	checkError(t, tracestate, err, testname, "Create Tracestate failed from null parent")
	checkKeyValue(t, tracestate, key1, value1, testname)
}

func TestCreateFromParentWithSingleKey(t *testing.T) {
	key1, value1, key2, value2 := "hello", "world", "foo", "bar"
	testname := "TestCreateFromParentWithSingleKey"

	parent, _ := NewFromParent(nil, key1, value1)
	tracestate, err := NewFromParent(parent, key2, value2)

	checkError(t, tracestate, err, testname, "Create Tracestate failed from parent with single key")
	checkKeyValue(t, tracestate, key2, value2, testname)
	checkFront(t, tracestate, key2, testname)
	checkBack(t, tracestate, key1, testname)
}

func TestCreateFromParentWithDoubleKeys(t *testing.T) {
	key1, value1, key2, value2, key3, value3 := "hello", "world", "foo", "bar", "bar", "baz"
	testname := "TestCreateFromParentWithDoubleKeys"

	entry1, _ := NewTracestateEntry(key1, value1)
	entry2, _ := NewTracestateEntry(key2, value2)
	entries := []*TracestateEntry{entry2, entry1}
	parent, _ := NewFromEntryArray(entries)
	tracestate, err := NewFromParent(parent, key3, value3)

	checkError(t, tracestate, err, testname, "Create Tracestate failed from parent with double keys")
	checkKeyValue(t, tracestate, key3, value3, testname)
	checkFront(t, tracestate, key3, testname)
	checkBack(t, tracestate, key1, testname)
}

func TestCreateFromParentWithExistingKey(t *testing.T) {
	key1, value1, key2, value2, key3, value3 := "hello", "world", "foo", "bar", "hello", "baz"
	testname := "TestCreateFromParentWithExistingKey"

	entry1, _ := NewTracestateEntry(key1, value1)
	entry2, _ := NewTracestateEntry(key2, value2)
	entries := []*TracestateEntry{entry1, entry2}
	parent, _ := NewFromEntryArray(entries)
	tracestate, err := NewFromParent(parent, key3, value3)

	checkError(t, tracestate, err, testname, "Create Tracestate failed with an existing key")
	checkKeyValue(t, tracestate, key3, value3, testname)
	checkFront(t, tracestate, key3, testname)
	checkBack(t, tracestate, key2, testname)
	checkSize(t, tracestate, 2, testname)
}

func TestImplicitImmutableTracestate(t *testing.T) {
	key1, value1, key2, value2, key3, value3 := "hello", "world", "hello", "bar", "foo", "baz"
	testname := "TestImplicitImmutableTracestate"

	parent, _ := NewFromParent(nil, key1, value1)
	tracestate, err := NewFromParent(parent, key2, value2)

	checkError(t, tracestate, err, testname, "Create Tracestate failed")
	checkKeyValue(t, tracestate, key2, value2, testname)
	checkKeyValue(t, parent, key2, value1, testname)

	// Get and update entries.
	entries := tracestate.TraceEntries()
	entry, _ := NewTracestateEntry(key3, value3)
	entries = append(entries, entry)

	// Check Tracestate does not have key3.
	checkKeyValue(t, tracestate, key3, "", testname)
}

func TestKeyWithValidChar(t *testing.T) {
	testname := "TestKeyWithValidChar"

	arrayRune := []rune("")
	for c := 'a'; c <= 'z'; c++ {
		arrayRune = append(arrayRune, c)
	}
	for c := '0'; c <= '9'; c++ {
		arrayRune = append(arrayRune, c)
	}
	arrayRune = append(arrayRune, '_')
	arrayRune = append(arrayRune, '-')
	arrayRune = append(arrayRune, '*')
	arrayRune = append(arrayRune, '/')
	key := string(arrayRune)
	tracestate, err := NewFromParent(nil, key, "world")

	checkError(t, tracestate, err, testname, "Create Tracestate failed with all valid characters in key")
}

func TestKeyWithInvalidChar(t *testing.T) {
	testname := "TestKeyWithInvalidChar"

	keys := []string{"1ab", "1ab2", "Abc", " abc", "a=b"}

	for _, key := range keys {
		tracestate, err := NewFromParent(nil, key, "world")
		expectError(t, tracestate, err, testname, fmt.Sprintf(
			"Create Tracesate did not err with invalid key=%q", key))
	}
}

func TestNilKey(t *testing.T) {
	testname := "TestNilKey"

	tracestate, err := NewFromParent(nil, "", "world")
	expectError(t, tracestate, err, testname, "Create Tracesate did not err with nil key (\"\")")
}

func TestValueWithInvalidChar(t *testing.T) {
	testname := "TestNilKey"

	// Invalid characters comma, equal, slash, star
	// Invalid trailing space
	keys := []string{"A=B", "A,B", "AB "}

	for _, value := range keys {
		tracestate, err := NewFromParent(nil, "hello", value)
		expectError(t, tracestate, err, testname,
			fmt.Sprintf("Create Tracesate did not err with invalid value=%q", value))
	}
}

func TestNilValue(t *testing.T) {
	testname := "TestNilValue"

	tracestate, err := NewFromParent(nil, "hello", "")
	expectError(t, tracestate, err, testname, "Create Tracesate did not err with nil value (\"\")")
}

func TestInvalidKeyLen(t *testing.T) {
	testname := "TestInvalidKeyLen"

	arrayRune := []rune("")
	for i := 0; i <= keyMaxSize+1; i++ {
		arrayRune = append(arrayRune, 'a')
	}
	key := string(arrayRune)
	tracestate, err := NewFromParent(nil, key, "world")

	expectError(t, tracestate, err, testname, "Create Tracestate did not err with invalid key length")
}

func TestInvalidValueLen(t *testing.T) {
	testname := "TestInvalidValueLen"

	arrayRune := []rune("")
	for i := 0; i <= valueMaxSize+1; i++ {
		arrayRune = append(arrayRune, 'a')
	}
	value := string(arrayRune)

	tracestate, err := NewFromParent(nil, "hello", value)
	expectError(t, tracestate, err, testname, "Create Tracestate did not err with invalid value length")
}

func TestCreateFromArrayWithOverLimitKVPairs(t *testing.T) {
	testname := "TestCreateFromArrayWithOverLimitKVPairs"

	entries := []*TracestateEntry{}
	for i := 0; i <= maxKeyValuePairs; i++ {
		key := fmt.Sprintf("a%db", i)
		entry, _ := NewTracestateEntry(key, "world")
		entries = append(entries, entry)
	}
	tracestate, err := NewFromEntryArray(entries)
	expectError(t, tracestate, err, testname,
		"Create did not err when attempted to exceed number of key-value pair limit")
}

func TestCreateFromArrayZeroKeys(t *testing.T) {
	testname := "TestCreateFromArrayZeroKeys"

	entries := []*TracestateEntry{}
	tracestate, err := NewFromEntryArray(entries)
	expectError(t, tracestate, err, testname,
		"Create did not err with zero key-value pair")
}

func TestCreateFromParentWithOverLimitKVPairs(t *testing.T) {
	testname := "TestCreateFromParentWithOverLimitKVPairs"

	entries := []*TracestateEntry{}
	for i := 0; i < maxKeyValuePairs; i++ {
		key := fmt.Sprintf("a%db", i)
		entry, _ := NewTracestateEntry(key, "world")
		entries = append(entries, entry)
	}
	parent, err := NewFromEntryArray(entries)

	checkError(t, parent, err, testname, fmt.Sprintf("Create failed to add %d key-value pair", maxKeyValuePairs))

	// Add one more to go over limit
	key := fmt.Sprintf("a%d", maxKeyValuePairs)
	tracestate, err := NewFromParent(parent, key, "world")
	expectError(t, tracestate, err, testname,
		"Create did not err when attempted to exceed number of key-value pair limit")
}

func TestCreateFromArrayWithDuplicateKeys(t *testing.T) {
	key1, value1, key2, value2, key3, value3 := "hello", "world", "foo", "bar", "hello", "baz"
	testname := "TestCreateFromArrayWithDuplicateKeys"

	entry1, _ := NewTracestateEntry(key1, value1)
	entry2, _ := NewTracestateEntry(key2, value2)
	entry3, _ := NewTracestateEntry(key3, value3)
	entries := []*TracestateEntry{entry1, entry2, entry3}
	tracestate, err := NewFromEntryArray(entries)

	expectError(t, tracestate, err, testname,
		"Create did not err when attempted to create from array with duplicate keys")
}
