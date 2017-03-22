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
		k, _ := DefaultKeyManager().CreateKeyStringUTF8(fmt.Sprintf("%s%d", "long key name that is more than fifty characters for testing puposes", i))
		for j := 0; j < valuesPerKey; j++ {
			v := fmt.Sprintf("%s%d", longValue, j)
			tags = append(tags, &tagStringUTF8{k, v})
			muts = append(muts, &mutationStringUTF8{
				tag: &tagStringUTF8{
					k: k,
					v: v,
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
		ctx = ContextWithDerivedTagsSet(ctx, m)
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

	builder := &TagsSetBuilder{}
	for _, td := range testData {
		tags, muts := createMutations(td.keysCount, td.valuesPerKey)
		builder.StartFromEmpty()
		for _, t := range tags {
			builder.AddOrReplaceTag(t)
		}
		ctx, err := createNewContextWithMutations(muts)
		if err != nil {
			t.Fatal(err)
		}

		v := ctx.Value(ctxKey{})
		if v == nil {
			t.Error("context has no census value")
		}

		ts := builder.Build()
		if !reflect.DeepEqual(ts, v.(*TagsSet)) {
			t.Errorf("\ngot: %v\nwant: %v\n", ts, v.(*TagsSet))
		}
	}
}

// BenchmarkNewContextWithTag_When1TagPresent measures the performance of
// calling NewContextWithTag with a (key,value) tuple where key and value are
// each around 80 characters, and the context already carries 1 tag.
func BenchmarkNewContextWithTag_When1TagPresent(b *testing.B) {
	tags, muts := createMutations(1, 1)
	builder := &TagsSetBuilder{}
	builder.StartFromEmpty()
	for _, t := range tags {
		builder.AddOrReplaceTag(t)
	}
	ctx := ContextWithDerivedTagsSet(context.Background(), muts...)
	k, _ := DefaultKeyManager().CreateKeyStringUTF8(longKey + "255")
	mut := &mutationStringUTF8{
		tag: &tagStringUTF8{
			k: k,
			v: longValue + "255",
		},
		behavior: BehaviorAddOrReplace,
	}

	for i := 0; i < b.N; i++ {
		_ = ContextWithDerivedTagsSet(ctx, mut)
	}
}

// BenchmarkNewContextWithTag_When100TagsPresent measures the performance of
// calling NewContextWithTag with a (key,value) tuple where key and value are
// each around 80 characters, and the context already carries 100 tags.
func BenchmarkNewContextWithTag_When100TagsPresent(b *testing.B) {
	tags, muts := createMutations(100, 1)
	builder := &TagsSetBuilder{}
	builder.StartFromEmpty()
	for _, t := range tags {
		builder.AddOrReplaceTag(t)
	}
	ctx := ContextWithDerivedTagsSet(context.Background(), muts...)
	k, _ := DefaultKeyManager().CreateKeyStringUTF8(longKey + "255")
	mut := &mutationStringUTF8{
		tag: &tagStringUTF8{
			k: k,
			v: longValue + "255",
		},
		behavior: BehaviorAddOrReplace,
	}

	for i := 0; i < b.N; i++ {
		_ = ContextWithDerivedTagsSet(ctx, mut)
	}
}
