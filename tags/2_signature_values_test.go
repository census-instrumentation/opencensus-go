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
	"reflect"
	"testing"
)

func Test_EncodeDecode_ValuesBytes(t *testing.T) {
	type testData struct {
		tagsSet   *TagSet
		keys      []Key
		wantSlice []*Tag
	}

	DefaultKeyManager().Clear()
	k1, _ := DefaultKeyManager().CreateKeyString("k1")
	k2, _ := DefaultKeyManager().CreateKeyString("k2")
	k3, _ := DefaultKeyManager().CreateKeyString("k3")
	k4, _ := DefaultKeyManager().CreateKeyInt64("k4")
	k5, _ := DefaultKeyManager().CreateKeyBool("k5")

	testSet := []testData{
		{
			&TagSet{
				map[Key][]byte{},
			},
			[]Key{k1},
			nil,
		},
		{
			&TagSet{
				map[Key][]byte{},
			},
			[]Key{k2},
			nil,
		},
		{
			&TagSet{
				map[Key][]byte{ k2:k2.CreateTag("v2").V},
			},		
			[]Key{k1},
			nil,
		},
		{
			&TagSet{
				map[Key][]byte{ k2:k2.CreateTag("v2").V},
			},		
			[]Key{k2},
			[]*Tag{k2.CreateTag("v2")},
		},
		{
			&TagSet{
				map[Key][]byte{ 
					k1:k1.CreateTag("v1").V,
					k2:k2.CreateTag("v2").V},
			},
			[]Key{k1},
			[]*Tag{k1.CreateTag("v1")},
		},
		{
			&TagSet{
				map[Key][]byte{ 
					k2:k2.CreateTag("v2").V,
					k1:k1.CreateTag("v1").V},
			},
			[]Key{k1},
			[]*Tag{k1.CreateTag("v1")},
		},
		{
			&TagSet{
				map[Key][]byte{ 
					k1:k1.CreateTag("v1").V,
					k2:k2.CreateTag("v2").V,
					k3:k3.CreateTag("v3").V},
			},			
			[]Key{k3, k1},
			[]*Tag{k3.CreateTag("v3"), k1.CreateTag("v1")},
		},
		{
			&TagSet{
				map[Key][]byte{ 
					k1:k1.CreateTag("v1").V,
					k4:k4.CreateTag(10).V,
					k5:k5.CreateTag(true).V},
			},			
			[]Key{k3, k4, k5},
			[]*Tag{k4.CreateTag(10), k5.CreateTag(true)},
		},
	}

	builder := &TagSetBuilder{}
	for i, td := range testSet {
		builder.StartFromTagSet(td.tagsSet)
		ts := builder.Build()

		vb := ts.toValuesBytes(td.keys)
t.Logf("------------\n%v\n", vb.buf)
		got := vb.toMap(td.keys)

		if len(got) != len(td.wantSlice) {
			t.Errorf("got len(decoded)=%v, want %v. Test case: %v", len(got), len(td.wantSlice), i)
		}

		for _, tag := range td.wantSlice {
			v, ok := got[tag.K]
			if !ok {
				t.Errorf("got key %v not found in decoded %v, want it found. Test case: %v", tag.K.Name(), got, i)
			}
			if !reflect.DeepEqual(v, tag.V) {
				t.Errorf("got tag %v in decoded, want %v. Test case: %v", v, tag.V, i)
			}
		}
	}
}

// Benchmark_Encode_ValuesBytes_When1TagPresent measures the performance of
// calling EncodeToValuesBytes a context with 1 tag where its key and value
// are around 80 characters each.
func Benchmark_Encode_ValuesBytes_When1TagPresent(b *testing.B) {
	ts, _ := createTagChange(1)
	var keys []Key
	for k, _ := range ts.m {
		keys = append(keys, k)
	}

	for i := 0; i < b.N; i++ {
		_ = ts.toValuesBytes(keys)
	}
}

// Benchmark_Decode_ValuesBytes_When1TagPresent measures the performance of
// calling DecodeFromValuesBytesToTagSet when signature has 1 tag and its
// key and value are around 80 characters each.
func Benchmark_Decode_ValuesBytes_When1TagPresent(b *testing.B) {
	ts, _ := createTagChange(1)
	var keys []Key
	for k, _ := range ts.m {
		keys = append(keys, k)
	}
	vb := ts.toValuesBytes(keys)

	for i := 0; i < b.N; i++ {
		_ = vb.toMap(keys)
	}
}

// Benchmark_Encode_ValuesBytes_When100TagsPresent measures the performance
// of calling EncodeToValuesBytes a context with 100 tags where each tag
// key and value are around 80 characters each.
func Benchmark_Encode_ValuesBytes_When100TagsPresent(b *testing.B) {
	ts, _ := createTagChange(100)
	var keys []Key
	for k, _ := range ts.m {
		keys = append(keys, k)
	}

	for i := 0; i < b.N; i++ {
		_ = ts.toValuesBytes(keys)
	}
}

// Benchmark_Decode_ValuesBytes_When100TagsPresent measures the performance
// of calling DecodeFromValuesBytesToTagSet when signature has 100 tags
// and each tag key and value are around 80 characters each.
func Benchmark_Decode_ValuesBytes_When100TagsPresent(b *testing.B) {
	ts, _ := createTagChange(100)
	var keys []Key
	for k, _ := range ts.m {
		keys = append(keys, k)
	}
	vb := ts.toValuesBytes(keys)

	for i := 0; i < b.N; i++ {
		_ = vb.toMap(keys)
	}
}
