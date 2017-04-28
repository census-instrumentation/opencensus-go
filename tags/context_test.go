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

package tags

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/net/context"
)

const longKey = "long tag key name that is more than fifty characters for testing puposes"
const longValue = "long tag value name that is more than fifty characters for testing puposes"

func createMutations(keysCount int) (*TagSet, []Mutation) {
	var muts []Mutation
	ts := &TagSet{
		m: make(map[Key]Tag),
	}
	for i := 0; i < keysCount; i++ {
		k, _ := DefaultKeyManager().CreateKeyStringUTF8(fmt.Sprintf("%s%d", longKey, i))
		ts.m[k] = &tagStringUTF8{k, longValue}
		muts = append(muts, &mutationStringUTF8{
			tag: &tagStringUTF8{
				k: k,
				v: longValue,
			},
			behavior: BehaviorAddOrReplace,
		})
	}
	return ts, muts
}

func Test_Context_WithDerivedTagSet_WhenNoTagPresent(t *testing.T) {
	testData := []int{1, 100}

	for _, i := range testData {
		want, muts := createMutations(i)

		ctx := ContextWithDerivedTagSet(context.Background(), muts...)
		v := ctx.Value(ctxKey{})
		if v == nil {
			t.Error("context has no *TagSet value")
		}

		if !reflect.DeepEqual(v.(*TagSet), want) {
			t.Errorf("\ngot: %v\nwant: %v\n", v.(*TagSet), want)
		}
	}
}

// BenchmarkContext_WithDerivedTagSet_When1TagPresent measures the performance
// of calling ContextWithDerivedTagSet with a (key,value) tuple where key and
// value are each around 80 characters, and the context already carries 1 tag.
func Benchmark_Context_WithDerivedTagSet_When1TagPresent(b *testing.B) {
	_, muts := createMutations(1)
	ctx := ContextWithDerivedTagSet(context.Background(), muts...)

	k, _ := DefaultKeyManager().CreateKeyStringUTF8(longKey + "255")
	mut := &mutationStringUTF8{
		tag: &tagStringUTF8{
			k: k,
			v: longValue + "255",
		},
		behavior: BehaviorAddOrReplace,
	}

	for i := 0; i < b.N; i++ {
		_ = ContextWithDerivedTagSet(ctx, mut)
	}
}

// BenchmarkContext_WithDerivedTagSet_When100TagsPresent measures the
// performance of calling ContextWithDerivedTagSet with a (key,value) tuple
// where key and value are each around 80 characters, and the context already
// carries 100 tags.
func Benchmark_Context_WithDerivedTagSet_When100TagsPresent(b *testing.B) {
	_, muts := createMutations(100)
	ctx := ContextWithDerivedTagSet(context.Background(), muts...)

	k, _ := DefaultKeyManager().CreateKeyStringUTF8(longKey + "255")
	mut := &mutationStringUTF8{
		tag: &tagStringUTF8{
			k: k,
			v: longValue + "255",
		},
		behavior: BehaviorAddOrReplace,
	}

	for i := 0; i < b.N; i++ {
		_ = ContextWithDerivedTagSet(ctx, mut)
	}
}
