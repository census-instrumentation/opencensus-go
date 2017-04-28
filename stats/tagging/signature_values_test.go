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
	"testing"
)

func Test_EncodeDecode_ValuesSignature(t *testing.T) {
	type testData struct {
		tagsSlice []Tag
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

	builder := &TagsSetBuilder{}
	for i, td := range testSet {
		builder.StartFromTags(td.tagsSlice)
		ts := builder.Build()

		encoded := EncodeToValuesSignature(ts, td.keys)

		decoded, err := DecodeFromValuesSignatureToTagsSet(encoded, td.keys)
		if err != nil {
			t.Errorf("got error %v, want no error when decoding. Test case: %v", err, i)
		}

		if len(decoded.m) != len(td.wantSlice) {
			t.Errorf("got len(decoded)=%v, want %v. Test case: %v", len(decoded.m), len(td.wantSlice), i)
		}

		for _, tag := range td.wantSlice {
			v, ok := decoded.m[tag.Key()]
			if !ok {
				t.Errorf("got key %v not found in decoded %v, want it found. Test case: %v", tag.Key().Name, decoded, i)
			}
			if !reflect.DeepEqual(v, tag) {
				t.Errorf("got tag %v in decoded, want %v. Test case: %v", v, tag, i)
			}
		}
	}
}

// Benchmark_Encode_ValuesSignature_When1TagPresent measures the performance of
// calling EncodeToValuesSignature a context with 1 tag where its key and value
// are around 80 characters each.
func Benchmark_Encode_ValuesSignature_When1TagPresent(b *testing.B) {
	ts, _ := createMutations(1)
	var keys []Key
	for _, t := range ts.m {
		keys = append(keys, t.Key())
	}

	for i := 0; i < b.N; i++ {
		_ = EncodeToValuesSignature(ts, keys)
	}
}

// Benchmark_Decode_ValuesSignature_When1TagPresent measures the performance of
// calling DecodeFromValuesSignatureToTagsSet when signature has 1 tag and its
// key and value are around 80 characters each.
func Benchmark_Decode_ValuesSignature_When1TagPresent(b *testing.B) {
	ts, _ := createMutations(1)
	var keys []Key
	for _, t := range ts.m {
		keys = append(keys, t.Key())
	}
	encoded := EncodeToValuesSignature(ts, keys)

	for i := 0; i < b.N; i++ {
		_, err := DecodeFromValuesSignatureToTagsSet([]byte(encoded), keys)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark_Encode_ValuesSignature_When100TagsPresent measures the performance
// of calling EncodeToValuesSignature a context with 100 tags where each tag
// key and value are around 80 characters each.
func Benchmark_Encode_ValuesSignature_When100TagsPresent(b *testing.B) {
	ts, _ := createMutations(100)
	var keys []Key
	for _, t := range ts.m {
		keys = append(keys, t.Key())
	}

	for i := 0; i < b.N; i++ {
		_ = EncodeToValuesSignature(ts, keys)
	}
}

// Benchmark_Decode_ValuesSignature_When100TagsPresent measures the performance
// of calling DecodeFromValuesSignatureToTagsSet when signature has 100 tags
// and each tag key and value are around 80 characters each.
func Benchmark_Decode_ValuesSignature_When100TagsPresent(b *testing.B) {
	ts, _ := createMutations(100)
	var keys []Key
	for _, t := range ts.m {
		keys = append(keys, t.Key())
	}
	encoded := EncodeToValuesSignature(ts, keys)

	for i := 0; i < b.N; i++ {
		_, err := DecodeFromValuesSignatureToTagsSet([]byte(encoded), keys)
		if err != nil {
			b.Fatal(err)
		}
	}
}
