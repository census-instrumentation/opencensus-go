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

func createTagChange(keysCount int) (*TagSet, []TagChange) {
	var changes []TagChange
	ts := newTagSet(0)
	for i := 0; i < keysCount; i++ {
		k, _ := DefaultKeyManager().CreateKeyString(fmt.Sprintf("%s%d", longKey, i))
		ts.upsertBytes(k, []byte(longValue))
		changes = append(changes, &tagChange{
			k: k,
			v: []byte(longValue),
			op: TagOpUpsert,
		})
	}
	return ts, changes
}

func Test_Context_WithDerivedTagSet_WhenNoTagPresent(t *testing.T) {
	testData := []int{1, 100}

	for _, i := range testData {
		want, changes := createTagChange(i)

		ctx := ContextWithDerivedTagSet(context.Background(), changes...)
		ts := FromContext(ctx)
		if len(ts.m) == 0 {
			t.Error("context has no *TagSet value")
		}

		if !reflect.DeepEqual(ts, want) {
			t.Errorf("\ngot: %v\nwant: %v\n", ts, want)
		}
	}
}

// BenchmarkContext_WithDerivedTagSet_When1TagPresent measures the performance
// of calling ContextWithDerivedTagSet with a (key,value) tuple where key and
// value are each around 80 characters, and the context already carries 1 tag.
func Benchmark_Context_WithDerivedTagSet_When1TagPresent(b *testing.B) {
	_, changes := createTagChange(1)
	ctx := ContextWithDerivedTagSet(context.Background(), changes...)

	k, _ := DefaultKeyManager().CreateKeyString(longKey + "255")
	c := &tagChange{
			k: k,
			v: []byte(longValue + "255"),
			op: TagOpUpsert,
		}

	for i := 0; i < b.N; i++ {
		_ = ContextWithDerivedTagSet(ctx, c)
	}
}

// BenchmarkContext_WithDerivedTagSet_When100TagsPresent measures the
// performance of calling ContextWithDerivedTagSet with a (key,value) tuple
// where key and value are each around 80 characters, and the context already
// carries 100 tags.
func Benchmark_Context_WithDerivedTagSet_When100TagsPresent(b *testing.B) {
	_, changes := createTagChange(100)
	ctx := ContextWithDerivedTagSet(context.Background(), changes...)

	k, _ := DefaultKeyManager().CreateKeyString(longKey + "255")
	c := &tagChange{
			k: k,
			v: []byte(longValue + "255"),
			op: TagOpUpsert,
		}

	for i := 0; i < b.N; i++ {
		_ = ContextWithDerivedTagSet(ctx, c)
	}
}
