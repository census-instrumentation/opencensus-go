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

import "testing"

// BenchmarkEncodeToValuesSignature_When1TagPresent measures the performance of
// calling encodeToValuesSignature a context with 1 tag where its key and
// value are around 80 characters each.
func BenchmarkEncodeToValuesSignature_When1TagPresent(b *testing.B) {
	tags, _ := createMutations(1, 1)
	ts := make(TagsSet)
	var keys []Key
	for _, t := range tags {
		ts[t.Key()] = t
		keys = append(keys, t.Key())
	}

	for i := 0; i < b.N; i++ {
		_ = EncodeToValuesSignature(ts, keys)
	}
}

// BenchmarkDecodeFromValuesSignatureToSlice_When1TagPresent measures the
// performance of calling decodeFromValuesSignatureToSlice when signature has 1
// tag and its key and value are around 80 characters each.
func BenchmarkDecodeFromValuesSignatureToSlice_When1TagPresent(b *testing.B) {
	tags, _ := createMutations(1, 1)
	ts := make(TagsSet)
	var keys []Key
	for _, t := range tags {
		ts[t.Key()] = t
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

// BenchmarkEncodeToValuesSignature_When100TagsPresent measures the performance
// of calling encodeToValuesSignature a context with 100 tags where each tag
// key and value are around 80 characters each.
func BenchmarkEncodeToValuesSignature_When100TagsPresent(b *testing.B) {
	tags, _ := createMutations(100, 1)
	ts := make(TagsSet)
	var keys []Key
	for _, t := range tags {
		ts[t.Key()] = t
		keys = append(keys, t.Key())
	}

	for i := 0; i < b.N; i++ {
		_ = EncodeToValuesSignature(ts, keys)
	}
}

// BenchmarkDecodeFromValuesSignatureToSlice_When100TagsPresent measures the
// performance of calling decodeFromValuesSignatureToSlice when signature has
// 100 tags and each tag key and value are around 80 characters each.
func BenchmarkDecodeFromValuesSignatureToSlice_When100TagsPresent(b *testing.B) {
	tags, _ := createMutations(100, 1)
	ts := make(TagsSet)
	var keys []Key
	for _, t := range tags {
		ts[t.Key()] = t
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
