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
	"context"
	"reflect"
	"testing"
)

func Test_ContextWithNewTagSet_Add_Retrieve(t *testing.T) {
	ts1 := newTagSet(2)
	ts1.upsertBytes(&KeyString{"k1", 1}, []byte("v1"))
	ts1.upsertBytes(&KeyString{"k2", 1}, []byte("v2"))
	ctx := ContextWithNewTagSet(context.Background(), ts1)
	got := FromContext(ctx)

	if !reflect.DeepEqual(got, ts1) {
		t.Errorf("got tag set %v, want tag set %v", got, ts1)
	}
}

func Test_ContextWithNewTagSet_Add_Replace_Retrieve(t *testing.T) {
	ts1 := newTagSet(1)
	ts1.upsertBytes(&KeyString{"k1", 1}, []byte("v1"))
	ctx1 := ContextWithNewTagSet(context.Background(), ts1)

	ts2 := newTagSet(1)
	ts2.upsertBytes(&KeyString{"k2", 1}, []byte("v2"))
	ctx2 := ContextWithNewTagSet(ctx1, ts2)

	got1 := FromContext(ctx1)
	got2 := FromContext(ctx2)

	if !reflect.DeepEqual(got1, ts1) {
		t.Errorf("got tag set %v, want tag set %v", got1, ts1)
	}

	if !reflect.DeepEqual(got2, ts2) {
		t.Errorf("got tag set %v, want tag set %v", got2, ts2)
	}
}
