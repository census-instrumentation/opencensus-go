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
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/net/context"
)

const longKey = "long tag key name that is more than fifty characters for testing puposes"
const longValue = "long tag value name that is more than fifty characters for testing puposes"

func createMutations(keysCount, valuesPerKey int) (tags []Tag, muts []Mutation) {
	for i := 0; i < keysCount; i++ {
		k, _ := DefaultKeyManager().CreateKeyString(fmt.Sprintf("%s%d", "long key name that is more than fifty characters for testing puposes", i))
		for j := 0; j < valuesPerKey; j++ {
			v := fmt.Sprintf("%s%d", longValue, j)
			tags = append(tags, &tagString{k, v})
			muts = append(muts, &mutationString{
				tagString: &tagString{
					keyString: k,
					v:         v,
				},
				behavior: BehaviorAddOrReplace,
			})
		}
	}
	return
}

func createNewContextWithMutations(muts []Mutation) (context.Context, error) {
	ctx := context.Background()
	for _, m := range muts {
		ctx = NewContextWithMutations(ctx, m)
	}
	return ctx, nil
}

func TestNewContextWithMutations(t *testing.T) {
	type newContextTestData struct {
		keysCount, valuesPerKey int
	}
	testData := []newContextTestData{}
	testData = append(testData, newContextTestData{1, 1})
	testData = append(testData, newContextTestData{100, 1})

	for _, td := range testData {
		tags, muts := createMutations(td.keysCount, td.valuesPerKey)
		ts := make(TagsSet)
		for _, t := range tags {
			ts[t.Key()] = t
		}
		ctx, err := createNewContextWithMutations(muts)
		if err != nil {
			t.Fatal(err)
		}

		v := ctx.Value(ctxKey{})
		if v == nil {
			t.Error("context has no census value")
		}

		if !reflect.DeepEqual(ts, v.(TagsSet)) {
			t.Errorf("\ngot: %v\nwant: %v\n", ts, v.(TagsSet))
		}
	}
}

/*

// BenchmarkNewContextWithTag_When1TagPresent measures the performance of
// calling NewContextWithTag with a (key,value) tuple where key and value are
// each around 80 characters, and the context already carries 1 tag.
func BenchmarkNewContextWithTag_When1TagPresent(b *testing.B) {
	tags := createTagInstances(1, 1)
	ctx, err := createNewContextWithTags(tags)
	if err != nil {
		b.Error(err)
	}
	tag := Tag{longKey + "255", longValue + "255"}

	for i := 0; i < b.N; i++ {
		_, err := NewContextWithTags(ctx, tag)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkNewContextWithTag_When100TagsPresent measures the performance of
// calling NewContextWithTag with a (key,value) tuple where key and value are
// each around 80 characters, and the context already carries 100 tags.
func BenchmarkNewContextWithTag_When100TagsPresent(b *testing.B) {
	tags := createTagInstances(100, 1)
	ctx, err := createNewContextWithTags(tags)
	if err != nil {
		b.Error(err)
	}
	tag := Tag{longKey + "255", longValue + "255"}

	for i := 0; i < b.N; i++ {
		_, err := NewContextWithTags(ctx, tag)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncodeToFullSignature_When1TagPresent measures the performance of
// calling encodeToFullSignature a context with 1 tag where its key and
// value are around 80 characters each.
func BenchmarkEncodeToFullSignature_When1TagPresent(b *testing.B) {
	tags := createTagInstances(1, 1)
	ct, err := newContextTags(nil, tags...)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_ = ct.encodeToFullSignature()
	}
}

// BenchmarkDecodeFromFullSignatureToSlice_When1TagPresent measures the
// performance of calling decodeFromFullSignatureToSlice when signature has 1
// tag and its key and value are around 80 characters each.
func BenchmarkDecodeFromFullSignatureToSlice_When1TagPresent(b *testing.B) {
	tags := createTagInstances(1, 1)
	ct, err := newContextTags(nil, tags...)
	if err != nil {
		b.Error(err)
	}
	encoded := ct.encodeToFullSignature()
	for i := 0; i < b.N; i++ {
		_, err := decodeFromFullSignatureToSlice([]byte(encoded))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncodeToFullSignature_When100TagsPresent measures the performance
// of calling encodeToFullSignature a context with 100 tags where each tag key
// and value are around 80 characters each.
func BenchmarkEncodeToFullSignature_When100TagsPresent(b *testing.B) {
	tags := createTagInstances(100, 1)
	ct, err := newContextTags(nil, tags...)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		_ = ct.encodeToFullSignature()
	}
}

// BenchmarkDecodeFromFullSignatureToSlice_When100TagsPresent measures the
// performance of calling decodeFromFullSignatureToSlice when signature has 100
// tags and each tag key and value are around 80 characters each.
func BenchmarkDecodeFromFullSignatureToSlice_When100TagsPresent(b *testing.B) {
	tags := createTagInstances(100, 1)
	ct, err := newContextTags(nil, tags...)
	if err != nil {
		b.Error(err)
	}
	encoded := ct.encodeToFullSignature()
	for i := 0; i < b.N; i++ {
		_, err := decodeFromFullSignatureToSlice([]byte(encoded))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncodeToValuesSignature_When1TagPresent measures the performance of
// calling encodeToValuesSignature a context with 1 tag where its key and
// value are around 80 characters each.
func BenchmarkEncodeToValuesSignature_When1TagPresent(b *testing.B) {
	tags := createTagInstances(1, 1)
	ct, err := newContextTags(nil, tags...)
	if err != nil {
		b.Error(err)
	}
	var keys []string
	for _, tag := range tags {
		keys = append(keys, tag.Key)
	}

	for i := 0; i < b.N; i++ {
		_ = ct.encodeToValuesSignature(keys)
	}
}

// BenchmarkDecodeFromValuesSignatureToSlice_When1TagPresent measures the
// performance of calling decodeFromValuesSignatureToSlice when signature has 1
// tag and its key and value are around 80 characters each.
func BenchmarkDecodeFromValuesSignatureToSlice_When1TagPresent(b *testing.B) {
	tags := createTagInstances(1, 1)
	ct, err := newContextTags(nil, tags...)
	if err != nil {
		b.Error(err)
	}
	var keys []string
	for _, tag := range tags {
		keys = append(keys, tag.Key)
	}

	encoded := ct.encodeToValuesSignature(keys)
	for i := 0; i < b.N; i++ {
		_, err := decodeFromValuesSignatureToSlice([]byte(encoded), keys)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEncodeToValuesSignature_When100TagsPresent measures the performance
// of calling encodeToValuesSignature a context with 100 tags where each tag
// key and value are around 80 characters each.
func BenchmarkEncodeToValuesSignature_When100TagsPresent(b *testing.B) {
	tags := createTagInstances(100, 1)
	ct, err := newContextTags(nil, tags...)
	if err != nil {
		b.Error(err)
	}
	var keys []string
	for _, tag := range tags {
		keys = append(keys, tag.Key)
	}

	for i := 0; i < b.N; i++ {
		_ = ct.encodeToValuesSignature(keys)
	}
}

// BenchmarkDecodeFromValuesSignatureToSlice_When100TagsPresent measures the
// performance of calling decodeFromValuesSignatureToSlice when signature has
// 100 tags and each tag key and value are around 80 characters each.
func BenchmarkDecodeFromValuesSignatureToSlice_When100TagsPresent(b *testing.B) {
	tags := createTagInstances(100, 1)
	ct, err := newContextTags(nil, tags...)
	if err != nil {
		b.Error(err)
	}
	var keys []string
	for _, tag := range tags {
		keys = append(keys, tag.Key)
	}

	encoded := ct.encodeToValuesSignature(keys)
	for i := 0; i < b.N; i++ {
		_, err := decodeFromValuesSignatureToSlice([]byte(encoded), keys)
		if err != nil {
			b.Fatal(err)
		}
	}
}
*/
