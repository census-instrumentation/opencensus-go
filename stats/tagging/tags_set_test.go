// Copyright 2017 Google Inc.
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

package tagging

import (
	"reflect"
	"sort"
	"testing"
)

func TestEncodeDecodeValuesSignature(t *testing.T) {
	type testData struct {
		tagsSet   []Tag
		keys      []Key
		wantSlice []Tag
	}

	DefaultKeyManager().Clear()
	k1, _ := DefaultKeyManager().CreateKeyStringUTF8("k1")
	k2, _ := DefaultKeyManager().CreateKeyStringUTF8("k2")
	k3, _ := DefaultKeyManager().CreateKeyStringUTF8("k3")
	k4, _ := DefaultKeyManager().CreateKeyInt64("k4")
	k5, _ := DefaultKeyManager().CreateKeyBool("k5")

	testSet := []testData{
		{
			[]Tag{},
			[]Key{k1},
			nil,
		},
		{
			[]Tag{},
			[]Key{k2},
			nil,
		},
		{
			[]Tag{k2.CreateTag("v2")},
			[]Key{k1},
			nil,
		},
		{
			[]Tag{k2.CreateTag("v2")},
			[]Key{k2},
			[]Tag{k2.CreateTag("v2")},
		},
		{
			[]Tag{k1.CreateTag("v1"), k2.CreateTag("v2")},
			[]Key{k1},
			[]Tag{k1.CreateTag("v1")},
		},
		{
			[]Tag{k2.CreateTag("v2"), k1.CreateTag("v1")},
			[]Key{k1},
			[]Tag{k1.CreateTag("v1")},
		},
		{
			[]Tag{k1.CreateTag("v1"), k2.CreateTag("v2"), k3.CreateTag("v3")},
			[]Key{k3, k1},
			[]Tag{k3.CreateTag("v3"), k1.CreateTag("v1")},
		},
		{
			[]Tag{k1.CreateTag("v1"), k4.CreateTag(10), k5.CreateTag(true)},
			[]Key{k3, k4, k5},
			[]Tag{k4.CreateTag(10), k5.CreateTag(true)},
		},
	}

	for i, td := range testSet {
		ts := make(TagsSet)
		for _, t := range td.tagsSet {
			ts[t.Key()] = t
		}

		encoded := EncodeToValuesSignature(ts, td.keys)

		decodedSlice, err := DecodeFromValuesSignatureToSlice(encoded, td.keys)
		if err != nil {
			t.Errorf("got error %v, want no error when decoding to slice. Test case: %v", err, i)
		}

		sort.Sort(tagSliceByName(decodedSlice))
		sort.Sort(tagSliceByName(td.wantSlice))
		if !reflect.DeepEqual(decodedSlice, td.wantSlice) {
			t.Errorf("got %v, want %v when decoding to slice . Test case: %v", decodedSlice, td.wantSlice, i)
		}

		decodedMap, err := DecodeFromValuesSignatureToTagsSet(encoded, td.keys)
		if err != nil {
			t.Errorf("got error %v, want no error when decoding to map. Test case: %v", err, i)
		}

		if len(decodedSlice) != len(decodedMap) {
			t.Errorf("got len(decodedSlice) %v different than len(decodedMap) %v, want them equal when decoding. Test case: %v", decodedSlice, decodedMap, i)
		}

		for _, tag := range decodedSlice {
			v, ok := decodedMap[tag.Key()]
			if !ok {
				t.Errorf("got key %v in decodedSlice not found in decodedMap %v , want them equivalent when decoding. Test case: %v", tag.Key().Name, decodedMap, i)
			}
			if !reflect.DeepEqual(v, tag) {
				t.Errorf("got %v in decodedSlice different than in decodedMap %v for key %v, want the same when decoding. Test case: %v", tag, v, tag.Key().Name, i)
			}
		}
	}
}

func TestEncodeDecodeFullSignature(t *testing.T) {
	type testData struct {
		tagsSet   []Tag
		wantSlice []Tag
	}

	DefaultKeyManager().Clear()
	k1, _ := DefaultKeyManager().CreateKeyStringUTF8("k1")
	k2, _ := DefaultKeyManager().CreateKeyStringUTF8("k2")
	k3, _ := DefaultKeyManager().CreateKeyInt64("k3")
	k4, _ := DefaultKeyManager().CreateKeyBool("k4")

	testSet := []testData{
		{
			[]Tag{},
			nil,
		},
		{
			[]Tag{k1.CreateTag("v1")},
			[]Tag{k1.CreateTag("v1")},
		},
		{
			[]Tag{k1.CreateTag("v1"), k2.CreateTag("v2")},
			[]Tag{k1.CreateTag("v1"), k2.CreateTag("v2")},
		},
		{
			[]Tag{k3.CreateTag(100), k2.CreateTag("v2"), k1.CreateTag("v1")},
			[]Tag{k3.CreateTag(100), k2.CreateTag("v2"), k1.CreateTag("v1")},
		},
		{
			[]Tag{k4.CreateTag(true), k3.CreateTag(100), k2.CreateTag("v2"), k1.CreateTag("v1")},
			[]Tag{k4.CreateTag(true), k3.CreateTag(100), k2.CreateTag("v2"), k1.CreateTag("v1")},
		},
	}

	for _, td := range testSet {
		ts := make(TagsSet)
		for _, t := range td.tagsSet {
			ts[t.Key()] = t
		}

		encoded := EncodeToFullSignature(ts)

		decodedSlice, err := DecodeFromFullSignatureToSlice(encoded)
		if err != nil {
			t.Errorf("got error %v, want no error when decoding to slice encoded %v", err, td)
			continue
		}
		sort.Sort(tagSliceByName(decodedSlice))
		sort.Sort(tagSliceByName(td.wantSlice))

		if !reflect.DeepEqual(decodedSlice, td.wantSlice) {
			t.Errorf("got %v, want %v when decoding to slice encoded %v", decodedSlice, td.wantSlice, td)
			continue
		}

		decodedMap, err := DecodeFromFullSignatureToTagsSet(encoded)
		if err != nil {
			t.Errorf("got error %v while decoding to map encoded %v, want no error", err, td)
			continue
		}

		if len(decodedSlice) != len(decodedMap) {
			t.Errorf("got len(decodedSlice) %v different than len(decodedMap) %v, want them equal when decoding %v", decodedSlice, decodedMap, td)
			continue
		}

		for _, tag := range decodedSlice {
			v, ok := decodedMap[tag.Key()]
			if !ok {
				t.Errorf("got key %v in decodedSlice not found in decodedMap %v , want them equivalent when decoding %v", tag.Key().Name(), decodedMap, td)
				continue
			}
			if !reflect.DeepEqual(v, tag) {
				t.Errorf("got %v in decodedSlice different than in decodedMap %v for key %v, want the same when decoding %v", tag, v, tag.Key().Name(), td)
			}
		}
	}
}
