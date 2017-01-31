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

package stats

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/net/context"
)

const longKey = "long tag key name that is more than fifty characters for testing puposes"
const longValue = "long tag value name that is more than fifty characters for testing puposes"

func createTagInstances(keysCount, valuesPerKey int) (tags []Tag) {
	for i := 0; i < keysCount; i++ {
		k := fmt.Sprintf("%s%d", "long key name that is more than fifty characters for testing puposes", i)
		for j := 0; j < valuesPerKey; j++ {
			v := fmt.Sprintf("%s%d", longValue, j)
			tags = append(tags, Tag{k, v})
		}
	}
	return
}

func createNewContextWithTags(tags []Tag) (context.Context, error) {
	ctx := context.Background()
	for _, tag := range tags {
		c, err := NewContextWithTags(ctx, tag)
		if err != nil {
			return nil, err
		}
		ctx = c
	}
	return ctx, nil
}

func TestNewContextWithTag(t *testing.T) {
	type newContextTestData struct {
		keysCount, valuesPerKey int
	}
	testData := []newContextTestData{}
	testData = append(testData, newContextTestData{1, 1})
	testData = append(testData, newContextTestData{100, 1})

	for _, td := range testData {
		tags := createTagInstances(td.keysCount, td.valuesPerKey)
		tagmap := make(contextTags)
		for _, t := range tags {
			tagmap[t.Key] = t.Value
		}

		ctx, err := createNewContextWithTags(tags)
		if err != nil {
			t.Fatal(err)
		}

		v := ctx.Value(censusKey{})
		if v == nil {
			t.Error("context has no census value")
		}

		if !reflect.DeepEqual(tagmap, v.(contextTags)) {
			t.Errorf("got: %v. Want: %v", tagmap, v.(contextTags))
		}
	}
}

func TestEncodeDecodeValuesSignature(t *testing.T) {
	type testData struct {
		ctxTagSet []Tag
		keys      []string
		wantSlice []Tag
	}

	testSet := []testData{
		{
			[]Tag{},
			[]string{},
			nil,
		},
		{
			[]Tag{},
			[]string{"k1"},
			nil,
		},
		{
			[]Tag{{"k2", "v2"}},
			[]string{"k1"},
			nil,
		},
		{
			[]Tag{{"k2", "v2"}},
			[]string{"k2"},
			[]Tag{{"k2", "v2"}},
		},
		{
			[]Tag{{"k1", "v1"}, {"k2", "v2"}},
			[]string{"k1"},
			[]Tag{{"k1", "v1"}},
		},
		{
			[]Tag{{"k2", "v2"}, {"k1", "v1"}},
			[]string{"k1"},
			[]Tag{{"k1", "v1"}},
		},
		{
			[]Tag{{"k1", "v1"}, {"k2", "v2"}, {"k3", "v3"}},
			[]string{"k3", "k1"},
			[]Tag{{"k3", "v3"}, {"k1", "v1"}},
		},
	}

	for _, td := range testSet {

		ct, err := newContextTags(nil, td.ctxTagSet...)
		if err != nil {

		}

		encoded := ct.encodeToValuesSignature(td.keys)

		decodedSlice, err := decodeFromValuesSignatureToSlice([]byte(encoded), td.keys)
		if err != nil {
			t.Errorf("got error %v, want no error when decoding to slice encoded %v", err, td)
		}

		if !reflect.DeepEqual(decodedSlice, td.wantSlice) {
			t.Errorf("got %v, want %v when decoding to slice encoded %v", decodedSlice, td.wantSlice, td)
		}

		decodedMap, err := decodeFromValuesSignatureToMap([]byte(encoded), td.keys)
		if err != nil {
			t.Errorf("got error %v while decoding to map encoded %v, want no error", err, td)
		}

		if len(decodedSlice) != len(decodedMap) {
			t.Errorf("got len(decodedSlice) %v different than len(decodedMap) %v, want them equal when decoding %v", decodedSlice, decodedMap, td)
		}

		for _, tag := range decodedSlice {
			v, ok := decodedMap[tag.Key]
			if !ok {
				t.Errorf("got key %v in decodedSlice not found in decodedMap %v , want them equivalent when decoding %v", tag.Key, decodedMap, td)
			}
			if v != tag.Value {
				t.Errorf("got %v in decodedSlice different than in decodedMap %v for key %v, want the same when decoding %v", tag.Value, v, tag.Key, td)
			}
		}
	}
}

func TestEncodeDecodeFullSignature(t *testing.T) {
	type testData struct {
		ctxTagSet []Tag
		wantSlice []Tag
	}

	testSet := []testData{
		{
			[]Tag{},
			nil,
		},
		{
			[]Tag{{"k1", "v1"}},
			[]Tag{{"k1", "v1"}},
		},
		{
			[]Tag{{"k1", "v1"}, {"k2", "v2"}},
			[]Tag{{"k1", "v1"}, {"k2", "v2"}},
		},
		{
			[]Tag{{"k3", "v3"}, {"k2", "v2"}, {"k1", "v1"}},
			[]Tag{{"k1", "v1"}, {"k2", "v2"}, {"k3", "v3"}},
		},
	}

	for _, td := range testSet {

		ct, err := newContextTags(nil, td.ctxTagSet...)
		if err != nil {

		}

		encoded := ct.encodeToFullSignature()

		decodedSlice, err := decodeFromFullSignatureToSlice([]byte(encoded))
		if err != nil {
			t.Errorf("got error %v, want no error when decoding to slice encoded %v", err, td)
		}

		if !reflect.DeepEqual(decodedSlice, td.wantSlice) {
			t.Errorf("got %v, want %v when decoding to slice encoded %v", decodedSlice, td.wantSlice, td)
		}

		decodedMap, err := decodeFromFullSignatureToMap([]byte(encoded))
		if err != nil {
			t.Errorf("got error %v while decoding to map encoded %v, want no error", err, td)
		}

		if len(decodedSlice) != len(decodedMap) {
			t.Errorf("got len(decodedSlice) %v different than len(decodedMap) %v, want them equal when decoding %v", decodedSlice, decodedMap, td)
		}

		for _, tag := range decodedSlice {
			v, ok := decodedMap[tag.Key]
			if !ok {
				t.Errorf("got key %v in decodedSlice not found in decodedMap %v , want them equivalent when decoding %v", tag.Key, decodedMap, td)
			}
			if v != tag.Value {
				t.Errorf("got %v in decodedSlice different than in decodedMap %v for key %v, want the same when decoding %v", tag.Value, v, tag.Key, td)
			}
		}
	}
}

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
