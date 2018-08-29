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
	"fmt"
	"testing"
)

var (
	emptyTracestate = &Tracestate{
		stateList: list.New(),
	}
)

func init() {
}

func checkFront(t *testing.T, tracestate *Tracestate, key, testname string) {
	entries := tracestate.GetEntries()
	entry := entries.Front().Value.(*Entry)
	got := entry.key
	if got != key {
		t.Errorf("Test:%s: key of the first entry in the list: got %q want %q", testname, got, key)
	}
}

func checkBack(t *testing.T, tracestate *Tracestate, key, testname string) {
	entries := tracestate.GetEntries()
	entry := entries.Back().Value.(*Entry)
	got := entry.key
	if got != key {
		t.Errorf("Test:%s: key of the last entry in the list: got %q want %q", testname, got, key)
	}
}

func checkSize(t *testing.T, tracestate *Tracestate, size int, testname string) {
	entries := tracestate.GetEntries()
	gotSize := entries.Len()
	if gotSize != size {
		t.Errorf("Test:%s: size of the list: got %q want %q", testname, gotSize, size)
	}
}

func checkKeyValue(t *testing.T, tracestate *Tracestate, key, value, testname string) {
	got := tracestate.Get(key)
	if got != value {
		t.Errorf("Test:%s: Get value for key=%s failed: got %q want %q", testname, key, got, value)
	}
}

func checkError(t *testing.T, tracestate *Tracestate, err error, testname, msg string) {
	if err != nil {
		t.Errorf("Test:%s: %s: tracestate=%q, error= %q", testname, msg, tracestate, err)
	}
}

func expectError(t *testing.T, tracestate *Tracestate, err error, testname, msg string) {
	if err == nil {
		t.Errorf("Test:%s: %s: tracestate=%q, error= %q", testname, msg, tracestate, err)
	}
}
func TestCreateWithNullParent(t *testing.T) {
	key1, value1 := "hello", "world"
	testname := "TestCreateWithNullParent"

	tracestate, err := CreateTracestate(nil, key1, value1)
	checkError(t, tracestate, err, testname, "CreateTracestate failed")
	checkKeyValue(t, tracestate, key1, value1, testname)
}

func TestSet(t *testing.T) {
	key1, value1, key2, value2 := "hello", "world", "foo", "bar"
	testname := "TestSet"

	tracestate, _ := CreateTracestate(nil, key1, value1)
	err := tracestate.Set(key2, value2)

	checkError(t, tracestate, err, testname, "Set failed")
	checkKeyValue(t, tracestate, key2, value2, testname)
}

func TestCreateWithParentWithSingleKey(t *testing.T) {
	key1, value1, key2, value2 := "hello", "world", "foo", "bar"
	testname := "TestCreateWithParentWithSingleKey"

	parent, _ := CreateTracestate(nil, key1, value1)
	tracestate, err := CreateTracestate(parent, key2, value2)

	checkError(t, tracestate, err, testname, "CreateTracestate failed from parent with single key")
	checkKeyValue(t, tracestate, key2, value2, testname)
	checkFront(t, tracestate, key2, testname)
	checkBack(t, tracestate, key1, testname)
}

func TestCreatWithParentWithDoubleKeys(t *testing.T) {
	key1, value1, key2, value2, key3, value3 := "hello", "world", "foo", "bar", "bar", "baz"
	testname := "TestCreatWithParentWithDoubleKeys"

	parent, _ := CreateTracestate(nil, key1, value1)
	parent, _ = CreateTracestate(parent, key2, value2)
	tracestate, err := CreateTracestate(parent, key3, value3)

	checkError(t, tracestate, err, testname, "CreateTracestate failed from parent with double key")
	checkKeyValue(t, tracestate, key3, value3, testname)
	checkFront(t, tracestate, key3, testname)
	checkBack(t, tracestate, key1, testname)
}

func TestCreateWithParentWithExistingKey(t *testing.T) {
	key1, value1, key2, value2, key3, value3 := "hello", "world", "foo", "bar", "hello", "baz"
	testname := "TestCreateWithParentWithExistingKey"

	parent, _ := CreateTracestate(nil, key1, value1)
	parent, _ = CreateTracestate(parent, key2, value2)
	tracestate, err := CreateTracestate(parent, key3, value3)

	checkError(t, tracestate, err, testname, "CreateTracestate failed with existing key")
	checkKeyValue(t, tracestate, key3, value3, testname)
	checkFront(t, tracestate, key3, testname)
	checkBack(t, tracestate, key2, testname)
	checkSize(t, tracestate, 2, testname)
}

func TestSetWithExistingKey(t *testing.T) {
	key1, value1, key2, value2, key3, value3 := "hello", "world", "foo", "bar", "hello", "baz"
	testname := "TestSetWithExistingKey"

	parent, _ := CreateTracestate(nil, key1, value1)
	tracestate, _ := CreateTracestate(parent, key2, value2)
	err := tracestate.Set(key3, value3)

	checkError(t, tracestate, err, testname, "Set failed with existing key")
	checkKeyValue(t, tracestate, key3, value3, testname)
	checkFront(t, tracestate, key3, testname)
	checkBack(t, tracestate, key2, testname)
	checkSize(t, tracestate, 2, testname)
}

func TestRemove(t *testing.T) {
	key1, value1, key2, value2, key3, value3 := "hello", "world", "foo", "bar", "bar", "baz"
	testname := "TestRemove"

	parent, _ := CreateTracestate(nil, key1, value1)
	tracestate, _ := CreateTracestate(parent, key2, value2)
	tracestate.Set(key3, value3)
	tracestate.Remove(key3)

	checkSize(t, tracestate, 2, testname)
	checkKeyValue(t, tracestate, key3, "", testname)
}

func TestRemoveNonExistent(t *testing.T) {
	key1, value1, key2, value2, key3 := "hello", "world", "foo", "bar", "bar"
	testname := "TestRemoveNonExistent"

	parent, _ := CreateTracestate(nil, key1, value1)
	tracestate, _ := CreateTracestate(parent, key2, value2)
	tracestate.Remove(key3)

	checkSize(t, tracestate, 2, testname)
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
	tracestate, err := CreateTracestate(emptyTracestate, key, "world")

	checkError(t, tracestate, err, testname, "Set failed with all valid characters in key")
}

func TestKeyWithInvalidFirstAndLastChar(t *testing.T) {
	testname := "TestKeyWithInvalidFirstAndLastChar"

	keys := []string{"1AB", "AB2", "1AB2"}

	for _, key := range keys {
		tracestate, err := CreateTracestate(emptyTracestate, key, "world")
		expectError(t, tracestate, err, testname, fmt.Sprintf("Set did not err with invalid key=%q", key))
	}
}

func TestNilKey(t *testing.T) {
	testname := "TestNilKey"

	tracestate, err := CreateTracestate(emptyTracestate, "", "world")
	expectError(t, tracestate, err, testname, "Set did not err with nil key (\"\")")
}

func TestValueWithInvalidChar(t *testing.T) {
	testname := "TestNilKey"

	// Invalid characters comma, equal, slash, star
	// Invalid trailing space
	keys := []string{"A=B", "A,B", "AB "}

	for _, value := range keys {
		tracestate, err := CreateTracestate(emptyTracestate, "hello", value)
		expectError(t, tracestate, err, testname, fmt.Sprintf("Set did not err with invalid value=%q", value))
	}
}

func TestNilValue(t *testing.T) {
	testname := "TestNilValue"

	tracestate, err := CreateTracestate(emptyTracestate, "hello", "")
	expectError(t, tracestate, err, testname, "Set did not err with nil value (\"\")")
}

func TestInvalidKeyLen(t *testing.T) {
	testname := "TestInvalidKeyLen"

	arrayRune := []rune("")
	for i := 0; i <= KeyMaxSize+1; i++ {
		arrayRune = append(arrayRune, 'a')
	}
	key := string(arrayRune)
	tracestate, err := CreateTracestate(emptyTracestate, key, "world")

	expectError(t, tracestate, err, testname, "CreateTracestate did not err with invalid key length")
}

func TestInvalidValueLen(t *testing.T) {
	testname := "TestInvalidValueLen"

	arrayRune := []rune("")
	for i := 0; i <= ValueMaxSize+1; i++ {
		arrayRune = append(arrayRune, 'a')
	}
	value := string(arrayRune)

	tracestate, err := CreateTracestate(emptyTracestate, "hello", value)
	expectError(t, tracestate, err, testname, "CreateTracestate did not err with invalid value length")
}

func TestMaxKeyValuePairs(t *testing.T) {
	testname := "TestMaxKeyValuePairs"

	tracestate := &Tracestate{}
	for i := 0; i < MaxKeyValuePairs; i++ {
		key := fmt.Sprintf("a%db", i)
		err := tracestate.Set(key, "c")
		checkError(t, tracestate, err, testname, fmt.Sprintf("Set failed to add %d key-value pair", i+1))

	}
	// Add one more
	key := fmt.Sprintf("a%d", MaxKeyValuePairs)
	err := tracestate.Set(key, "c")
	expectError(t, tracestate, err, testname,
		"Set did not err when attempted to exceed number of key-value pair limit")
}
