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

func Test_EncodeDecode_GRPCSignature(t *testing.T) {
	type testData struct {
		want *TagSet
	}

	DefaultKeyManager().Clear()
	k1, _ := DefaultKeyManager().CreateKeyString("k1")
	k2, _ := DefaultKeyManager().CreateKeyString("k2")
	k3, _ := DefaultKeyManager().CreateKeyInt64("k3")
	k4, _ := DefaultKeyManager().CreateKeyBool("k4")

	testSet := []testData{
		{
			&TagSet{
				map[Key][]byte{},
			},
		},
		{
			&TagSet{
				map[Key][]byte{k1: k1.CreateTag("v1").V},
			},
		},
		{
			&TagSet{
				map[Key][]byte{
					k1: k1.CreateTag("v1").V,
					k2: k2.CreateTag("v2").V},
			},
		},
		{
			&TagSet{
				map[Key][]byte{
					k3: k3.CreateTag(100).V,
					k2: k2.CreateTag("v2").V,
					k1: k1.CreateTag("v1").V},
			},
		},
		{
			&TagSet{
				map[Key][]byte{
					k4: k4.CreateTag(true).V,
					k3: k3.CreateTag(100).V,
					k2: k2.CreateTag("v2").V,
					k1: k1.CreateTag("v1").V},
			},
		},
	}

	builder := &TagSetBuilder{}
	for i, td := range testSet {
		builder.StartFromTagSet(td.want)
		ts := builder.Build()

		gc := &GRPCCodec{}
		encoded := gc.EncodeTagSet(ts)
		got, err := gc.DecodeTagSet(encoded)
		if err != nil {
			t.Errorf("got error '%v', want no error when decoding. Test case: '%v'", err, i)
			continue
		}

		if len(got.m) != len(td.want.m) {
			t.Errorf("got len(decoded)=%v, want %v. Test case: %v", len(got.m), len(td.want.m), i)
			continue
		}

		for k, v := range td.want.m {
			gotV, ok := got.m[k]
			if !ok {
				t.Errorf("got TagSet not containing key %v, want it found. Test case: %v", k.Name(), i)
				continue
			}
			if !reflect.DeepEqual(gotV, v) {
				t.Errorf("got tag value %v in decoded, want %v. Test case: %v", gotV, v, i)
			}
		}
	}
}
func Test_EncodeDecode_GRPCSignature_When100TagsPresent(t *testing.T) {
	ts, _ := createTagChange(100)
	gc := &GRPCCodec{}
	encoded := gc.EncodeTagSet(ts)
	decoded, err := gc.DecodeTagSet(encoded)
	if err != nil {
		t.Fatalf("got error %v, want no error when decoding", err)
	}

	if len(decoded.m) != len(ts.m) {
		t.Fatalf("got len(decoded)=%v, want %vv", len(decoded.m), len(ts.m))
	}

	if !reflect.DeepEqual(decoded.m, ts.m) {
		t.Fatalf("got %v in decoded, want %v", decoded.m, ts.m)
	}
}

// Benchmark_Encode_GRPCSignature_When1TagPresent measures the performance of
// calling EncodeTagSet a context with 1 tag where its key and value are around
// 80 characters each.
func Benchmark_Encode_GRPCSignature_When1TagPresent(b *testing.B) {
	ts, _ := createTagChange(1)
	gc := &GRPCCodec{}
	for i := 0; i < b.N; i++ {
		_ = gc.EncodeTagSet(ts)
	}
}

// Benchmark_Decode_GRPCSignature_When1TagPresent measures the performance of
// calling DecodeTagSet when signature has 1 tag and its key and value are
// around 80 characters each.
func Benchmark_Decode_GRPCSignature_When1TagPresent(b *testing.B) {
	ts, _ := createTagChange(1)
	gc := &GRPCCodec{}
	encoded := gc.EncodeTagSet(ts)

	for i := 0; i < b.N; i++ {
		_, err := gc.DecodeTagSet(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark_Encode_GRPCSignature_When100TagsPresent measures the performance
// of calling EncodeTagSet a context with 100 tags where each tag key and value
// are around 80 characters each.
func Benchmark_Encode_GRPCSignature_When100TagsPresent(b *testing.B) {
	ts, _ := createTagChange(100)
	gc := &GRPCCodec{}
	for i := 0; i < b.N; i++ {
		_ = gc.EncodeTagSet(ts)
	}
}

// Benchmark_Decode_GRPCSignature_When100TagsPresent measures the performance
// of calling DecodeTagSet when signature has 100 tags and each tag key and
// value are around 80 characters each.
func Benchmark_Decode_GRPCSignature_When100TagsPresent(b *testing.B) {
	ts, _ := createTagChange(100)
	gc := &GRPCCodec{}
	encoded := gc.EncodeTagSet(ts)

	for i := 0; i < b.N; i++ {
		_, err := gc.DecodeTagSet(encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}
