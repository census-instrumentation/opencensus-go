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

package tag

import (
	"reflect"
	"sort"
	"testing"

	"golang.org/x/net/context"
)

var keys []Key

type pair struct {
	k Key
	v string
}

func Test_EncodeDecode_Set(t *testing.T) {
	k1, _ := NewKey("k1")
	k2, _ := NewKey("k2")
	k3, _ := NewKey("k3 is very weird <>.,?/'\";:`~!@#$%^&*()_-+={[}]|\\")
	k4, _ := NewKey("k4")

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
		mods := make([]Mutator, len(tc.pairs))
		for i, pair := range tc.pairs {
			mods[i] = Upsert(pair.k, pair.v)
		}
		ts, err := NewMap(context.Background(), mods...)
		if err != nil {
			t.Errorf("%v: NewMap = %v", tc.label, err)
		}
		encoded := Encode(ts)
		decoded, err := Decode(encoded)

		if err != nil {
			t.Errorf("%v: decoding encoded tag map failed: %v", tc.label, err)
		}

		got := make([]pair, 0)
		for k, v := range decoded.m {
			got = append(got, pair{k, string(v)})
		}
		want := tc.pairs

		sort.Sort(pairs(got))
		sort.Sort(pairs(want))

		if !reflect.DeepEqual(got, want) {
			t.Errorf("%v: decoded tag map = %#v; want %#v", tc.label, got, want)
		}
	}
}

type pairs []pair

func (ps pairs) Len() int           { return len(ps) }
func (ps pairs) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }
func (ps pairs) Less(i, j int) bool { return ps[i].k.name < ps[j].k.name }

var _ sort.Interface = (pairs)(nil)
