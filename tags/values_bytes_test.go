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
	"testing"
)

func Test_EncodeDecode_ValuesBytes(t *testing.T) {
	type testData struct {
		label     int
		m         *Map
		keys      []Key
		wantSlice map[Key][]byte
	}

	km := newKeysManager()
	k1, _ := km.newStringKey("k1")
	k2, _ := km.newStringKey("k2")
	k3, _ := km.newStringKey("k3")

	testSet := []testData{
		{
			0,
			&Map{
				map[Key][]byte{},
			},
			[]Key{k1},
			nil,
		},
		{
			1,
			&Map{
				map[Key][]byte{k2: []byte("v2")},
			},
			[]Key{},
			nil,
		},
		{
			3,
			&Map{
				map[Key][]byte{k2: []byte("v2")},
			},
			[]Key{k1},
			nil,
		},
		{
			4,
			&Map{
				map[Key][]byte{k2: []byte("v2")},
			},
			[]Key{k2},
			map[Key][]byte{
				k2: []byte("v2"),
			},
		},
		{
			5,
			&Map{
				map[Key][]byte{
					k1: []byte("v1"),
					k2: []byte("v2")},
			},
			[]Key{k1},
			map[Key][]byte{
				k1: []byte("v1"),
			},
		},
		{
			6,
			&Map{
				map[Key][]byte{
					k2: []byte("v2"),
					k1: []byte("v1")},
			},
			[]Key{k1, k2},
			map[Key][]byte{
				k1: []byte("v1"),
				k2: []byte("v2"),
			},
		},
		{
			7,
			&Map{
				map[Key][]byte{
					k1: []byte("v1"),
					k2: []byte("v2"),
					k3: []byte("v3")},
			},
			[]Key{k3, k1},
			map[Key][]byte{
				k1: []byte("v1"),
				k3: []byte("v3"),
			},
		},
	}

	for i, td := range testSet {
		ts := NewMap(td.m)

		vb := toValuesBytes(ts, td.keys)
		got := vb.toMap(td.keys)
		if len(got) != len(td.wantSlice) {
			t.Errorf("got len(decoded)=%v, want %v. Test case: %v", len(got), len(td.wantSlice), i)
		}

		for wantK, wantV := range td.wantSlice {
			v, ok := got[wantK]
			if !ok {
				t.Errorf("got key %v not found in decoded %v, want it found. Test case: %v", wantK.Name(), got, i)
			}
			if !reflect.DeepEqual(v, wantV) {
				t.Errorf("got tag %v in decoded, want %v. Test case: %v", v, wantV, i)
			}
		}
	}
}
