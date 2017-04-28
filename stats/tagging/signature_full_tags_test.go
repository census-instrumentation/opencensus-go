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

func Test_EncodeDecode_FullSignature(t *testing.T) {
	type testData struct {
		tagsSlice []Tag
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

	builder := &TagsSetBuilder{}
	for i, td := range testSet {
		builder.StartFromTags(td.tagsSlice)
		ts := builder.Build()

		encoded := EncodeToFullSignature(ts)

		decoded, err := DecodeFromFullSignatureToTagsSet(encoded)
		if err != nil {
			t.Errorf("got error '%v', want no error when decoding. Test case: '%v'", err, i)
			continue
		}

		if len(decoded.m) != len(td.wantSlice) {
			t.Errorf("got len(decoded)=%v, want %v. Test case: %v", len(decoded.m), len(td.wantSlice), i)
			continue
		}

		for _, tag := range td.wantSlice {
			v, ok := decoded.m[tag.Key()]
			if !ok {
				t.Errorf("got key %v not found in decoded %v, want it found. Test case: %v", tag.Key().Name, decoded, i)
				continue
			}
			if !reflect.DeepEqual(v, tag) {
				t.Errorf("got tag %v in decoded, want %v. Test case: %v", v, tag, i)
			}
		}
	}
}

func Test_EncodeDecode_FullSignature_When100TagsPresent(t *testing.T) {
	ts, _ := createMutations(100)
	encoded := EncodeToFullSignature(ts)
	decoded, err := DecodeFromFullSignatureToTagsSet(encoded)
	if err != nil {
		t.Fatalf("got error %v, want no error when decoding", err)
	}

	if len(decoded.m) != len(ts.m) {
		t.Fatalf("got len(decoded)=%v, want %vv", len(decoded.m), len(ts.m))
	}

	if !reflect.DeepEqual(decoded.m, ts.m) {
		t.Fatalf("got %v in decoded, want %v", decoded.m, ts.m)
	}
}

// Benchmark_Encode_FullSignature_When1TagPresent measures the performance of
// calling EncodeToFullSignature a context with 1 tag where its key and value
// are around 80 characters each.
func Benchmark_Encode_FullSignature_When1TagPresent(b *testing.B) {
	ts, _ := createMutations(1)
	for i := 0; i < b.N; i++ {
		_ = EncodeToFullSignature(ts)
	}
}

// Benchmark_Decode_FullSignature_When1TagPresent measures the performance of
// calling DecodeFromFullSignatureToTagsSet when signature has 1 tag and its
// key and value are around 80 characters each.
func Benchmark_Decode_FullSignature_When1TagPresent(b *testing.B) {
	ts, _ := createMutations(1)
	encoded := EncodeToFullSignature(ts)

	for i := 0; i < b.N; i++ {
		_, err := DecodeFromFullSignatureToTagsSet(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark_Encode_FullSignature_When100TagsPresent measures the performance
// of calling EncodeToFullSignature a context with 100 tags where each tag key
// and value are around 80 characters each.
func Benchmark_Encode_FullSignature_When100TagsPresent(b *testing.B) {
	ts, _ := createMutations(100)

	for i := 0; i < b.N; i++ {
		_ = EncodeToFullSignature(ts)
	}
}

// Benchmark_Decode_FullSignature_When100TagsPresent measures the performance
// of calling DecodeFromFullSignatureToTagsSet when signature has 100 tags and
// each tag key and value are around 80 characters each.
func Benchmark_Decode_FullSignature_When100TagsPresent(b *testing.B) {
	ts, _ := createMutations(100)
	encoded := EncodeToFullSignature(ts)

	for i := 0; i < b.N; i++ {
		_, err := DecodeFromFullSignatureToTagsSet([]byte(encoded))
		if err != nil {
			b.Fatal(err)
		}
	}
}
