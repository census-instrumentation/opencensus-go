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
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

//-----------------------------------------------------------------------------
// Part of context_test.go
//-----------------------------------------------------------------------------

const longKey = "long tag key name that is more than fifty characters for testing puposes"
const longValue = "long tag value name that is more than fifty characters for testing puposes"

func createTagChange(keysCount int) (*TagSet, []TagChange) {
	var changes []TagChange
	ts := newTagSet(0)
	for i := 0; i < keysCount; i++ {
		k, _ := DefaultKeyManager().CreateKeyString(fmt.Sprintf("%s%d", longKey, i))
		ts.upsertBytes(k, []byte(longValue))
		changes = append(changes, &tagChange{
			k:  k,
			v:  []byte(longValue),
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
		k:  k,
		v:  []byte(longValue + "255"),
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
		k:  k,
		v:  []byte(longValue + "255"),
		op: TagOpUpsert,
	}

	for i := 0; i < b.N; i++ {
		_ = ContextWithDerivedTagSet(ctx, c)
	}
}

//-----------------------------------------------------------------------------
// Part of keys_manager_test.go
//-----------------------------------------------------------------------------

func Test_KeysManager_NoErrors2(t *testing.T) {
	type testData struct {
		createCommands      []func(km *keysManager) (Key, error)
		wantCount           int
		wantCountAfterClear int
	}

	testSet := []testData{
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k2") },
				func(km *keysManager) (Key, error) { return km.createKeyInt64("k3") },
				func(km *keysManager) (Key, error) { return km.createKeyBool("k4") },
			},
			4,
			0,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
			},
			1,
			0,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyBool("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyBool("k1") },
			},
			1,
			0,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyInt64("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyInt64("k1") },
			},
			1,
			0,
		},
		{
			[]func(km *keysManager) (Key, error){},
			0,
			0,
		},
	}

	for i, td := range testSet {
		km := newKeysManager()
		for j, f := range td.createCommands {
			_, err := f(km)
			if err != nil {
				t.Errorf("got error %v, want no error calling keysManager.createKeyXYZ(...). Test case: %v, function: %v", err, i, j)
			}
		}
		gotCount := km.count()
		if gotCount != td.wantCount {
			t.Errorf("got keys count %v, want keys count %v", gotCount, td.wantCount)
		}

		km.clear()
		gotCountAfterClear := km.count()
		if gotCountAfterClear != td.wantCountAfterClear {
			t.Errorf("got keys count %v, want keys count %v after clear()", gotCountAfterClear, td.wantCountAfterClear)
		}
	}
}

func Test_KeysManager_ExpectErrors2(t *testing.T) {
	type testData struct {
		createCommands []func() (Key, error)
		wantErrCount   int
	}

	testSet := []testData{
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyInt64("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyBool("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
			},
			2,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyInt64("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyBool("k1") },
			},
			1,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyInt64("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
			},
			1,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyBool("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
			},
			1,
		},
	}

	for i, td := range testSet {
		km := newKeysManager()
		gotErrCount := 0
		for _, f := range td.createCommands {
			_, err := f(km)
			if err != nil {
				gotErrCount++
			}
		}

		if gotErrCount != td.wantErrCount {
			t.Errorf("got errors count %v, want errors count %v. Test case %v", gotErrCount, td.wantErrCount, i)
		}
	}
}

//-----------------------------------------------------------------------------
// Part of values_bytes_test.go
//-----------------------------------------------------------------------------

func Test_EncodeDecode_ValuesBytes2(t *testing.T) {
	type testData struct {
		tagsSet   *TagSet
		keys      []Key
		wantSlice map[Key][]byte
	}

	km := newKeysManager()
	k1, _ := km.createKeyString("k1")
	k2, _ := km.createKeyString("k2")
	k3, _ := km.createKeyString("k3")
	k4, _ := km.createKeyInt64("k4")
	k5, _ := km.createKeyBool("k5")

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
				map[Key][]byte{k2: []byte("v2")},
			},
			[]Key{k1},
			nil,
		},
		{
			&TagSet{
				map[Key][]byte{k2: []byte("v2")},
			},
			[]Key{k2},
			map[Key][]byte{
				k2: []byte("v2"),
			},
		},
		{
			&TagSet{
				map[Key][]byte{
					k1: []byte("v1"),
					k2: []byte("v2")},
			},
			[]Key{k1},
			map[Key][]byte{
				k1: []byte("v1"),
				k2: []byte("v2"),
			},
		},
		{
			&TagSet{
				map[Key][]byte{
					k2: []byte("v2"),
					k1: []byte("v1")},
			},
			[]Key{k1},
			map[Key][]byte{
				k1: []byte("v1"),
			},
		},
		{
			&TagSet{
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
		{
			&TagSet{
				map[Key][]byte{
					k1: []byte("v1"),
					k4: func() {
						v := make([]byte, 8)
						binary.LittleEndian.PutUint64(v, 10)
						return v
					}(),
					k5: byte(1), // true
				},
			},
			//[]Key{k3, k4, k5},
			[]Key{k3},
			map[Key][]byte{
				k1: []byte("v1"),
				k4: func() {
					v := make([]byte, 8)
					binary.LittleEndian.PutUint64(v, 10)
					return v
				}(),
				k5: byte(1), //true
			},
		},
	}

	builder := &TagSetBuilder{}
	for i, td := range testSet {
		builder.StartFromTagSet(td.tagsSet)
		ts := builder.Build()

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
	for k := range ts.m {
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
	for k := range ts.m {
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
	for k := range ts.m {
		keys = append(keys, k)
	}
	vb := ts.toValuesBytes(keys)

	for i := 0; i < b.N; i++ {
		_ = vb.toMap(keys)
	}
}

//-----------------------------------------------------------------------------
// Part of grpc_codec_test.go
//-----------------------------------------------------------------------------

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
