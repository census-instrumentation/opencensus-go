// Copyright 2017, OpenCensus Authors
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
	"reflect"
	"sort"
	"testing"
	"unsafe"
)

func Test_EncodeDecode_TagSet(t *testing.T) {
	k1, _ := KeyStringByName("k1")
	k2, _ := KeyStringByName("k2")
	k3, _ := KeyStringByName("k3 is very weird <>.,?/'\";:`~!@#$%^&*()_-+={[}]|\\")
	k4, _ := KeyStringByName("k4")

	type pair struct {
		k *KeyString
		v string
	}

	type testCase struct {
		label string
		pairs []pair
	}

	testCases := []testCase{
		{
			"0",
			[]pair{},
		},
		{
			"1",
			[]pair{
				{k1, "v1"},
			},
		},
		{
			"2",
			[]pair{
				{k1, "v1"},
				{k2, "v2"},
			},
		},
		{
			"3",
			[]pair{
				{k1, "v1"},
				{k2, "v2"},
				{k3, "v3"},
			},
		},
		{
			"4",
			[]pair{
				{k1, "v1"},
				{k2, "v2"},
				{k3, "v3"},
				{k4, "v4 is very weird <>.,?/'\";:`~!@#$%^&*()_-+={[}]|\\"},
			},
		},
	}

	for _, tc := range testCases {
		tsb := NewTagSetBuilder(nil)
		for _, pair := range tc.pairs {
			tsb.UpsertString(pair.k, pair.v)
		}
		ts := tsb.Build()

		encoded := Encode(ts)
		decoded, err := Decode(encoded)

		if err != nil {
			t.Errorf("Test case '%v'. Decoding encoded tagSet failed. %v", tc.label, err)
		}

		got := make([]pair, 0)
		for k, v := range decoded.m {
			ks, ok := k.(*KeyString)
			if !ok {
				t.Errorf("Test case '%v'. Wrong key type. got %T, want *KeyString", tc.label, k)
			}
			got = append(got, pair{ks, string(v)})
		}
		want := tc.pairs

		sort.Slice(got, func(i, j int) bool { return uintptr(unsafe.Pointer(got[i].k)) < uintptr(unsafe.Pointer(got[j].k)) })
		sort.Slice(want, func(i, j int) bool { return uintptr(unsafe.Pointer(want[i].k)) < uintptr(unsafe.Pointer(want[j].k)) })

		if !reflect.DeepEqual(got, tc.pairs) {
			t.Errorf("Test case '%v'. Decoded tagSet is wrong. Got %v, want %v", tc.label, got, tc.pairs)
		}

	}
}
