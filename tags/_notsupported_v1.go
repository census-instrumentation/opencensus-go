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
	"unsafe"
)

//-----------------------------------------------------------------------------
// Part of context.go
//-----------------------------------------------------------------------------

// ContextWithDerivedTagSet creates a new context from the old one replacing any
// existing TagSet. The new TagSet contains the tags already presents in the
// existing TagSet to which the mutations ms are applied
func ContextWithDerivedTagSet(ctx context.Context, tcs ...TagChange) context.Context {
	builder := &TagSetBuilder{}

	oldTs, ok := ctx.Value(ctxKey{}).(*TagSet)
	if !ok {
		builder.StartFromEmpty()
	} else {
		builder.StartFromTagSet(oldTs)
	}

	builder.Apply(tcs...)
	return context.WithValue(ctx, ctxKey{}, builder.Build())
}

//-----------------------------------------------------------------------------
// Part of key.go
//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------
// The methods below are related to tag and tag change creations and are not
// supported in v0.1. They are subject to change.

// CreateChange creates a change operation to a key.
func (k *KeyString) CreateChange(s string, op TagOp) TagChange {
	return &tagChange{
		k:  k,
		v:  []byte(s),
		op: op,
	}
}

// CreateTag creates a tag from a key.
func (k *KeyString) CreateTag(s string) *Tag {
	return &Tag{
		K: k,
		V: []byte(s),
	}
}

//-----------------------------------------------------------------------------
// The methods below are related to bool tag types are not supported in v0.1
// and are subject to change.

// KeyBool implements the Key interface and is used to represent keys for which
// the value type is a string.
type KeyBool struct {
	name string
	id   uint16
}

// CreateChange creates a change operation to a key.
func (k *KeyBool) CreateChange(b bool, op TagOp) TagChange {
	tc := &tagChange{
		k:  k,
		op: op,
	}
	if b {
		tc.v = []byte{1}
		return tc
	}
	tc.v = []byte{0}
	return tc
}

// CreateTag creates a tag from a key.
func (k *KeyBool) CreateTag(b bool) *Tag {
	t := &Tag{
		K: k,
	}
	if b {
		t.V = []byte{1}
		return t
	}
	t.V = []byte{0}
	return t
}

// Name returns the unique name of a key.
func (k *KeyBool) Name() string {
	return k.name
}

// ID returns the id of a key inside hte process.
func (k *KeyBool) ID() uint16 {
	return k.id
}

//-----------------------------------------------------------------------------
// The methods below are related to int64 tag types are not supported in v0.1
// and are subject to change.

// KeyInt64 implements the Key interface and is used to represent keys for
// which the value type is a int64.
type KeyInt64 struct {
	name string
	id   uint16
}

// CreateChange creates a change operation to a key.
func (k *KeyInt64) CreateChange(i int64, op TagOp) TagChange {
	tc := &tagChange{
		k:  k,
		op: op,
	}
	tc.v = make([]byte, 8)
	binary.LittleEndian.PutUint64(tc.v, uint64(i))
	return tc
}

// CreateTag creates a tag from a key.
func (k *KeyInt64) CreateTag(i int64) *Tag {
	t := &Tag{
		K: k,
		V: make([]byte, 8),
	}
	binary.LittleEndian.PutUint64(t.V, uint64(i))
	return t
}

// Name returns the unique name of a key.
func (k *KeyInt64) Name() string {
	return k.name
}

// ID returns the id of a key inside hte process.
func (k *KeyInt64) ID() uint16 {
	return k.id
}

func getKeyByID(id uint16) Key {
	if int(id) >= len(keys) {
		return nil
	}
	return keys[id]
}

//-----------------------------------------------------------------------------
// Part of keys_manager.go
//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------
// The methods below are related to bool and int64 tag types are not supported
// in v0.1 and are subject to change.

// CreateKeyBool creates/retrieves the *KeyBool identified by name.
var CreateKeyBool func(name string) (*KeyBool, error)

// CreateKeyInt64 creates/retrieves the *KeyInt64 identified by name.
var CreateKeyInt64 func(name string) (*KeyInt64, error)

//-----------------------------------------------------------------------------
// Part of tag_change.go
//-----------------------------------------------------------------------------

// TagOp defines the types of operations allowed.
type TagOp byte

const (
	// TagOpInvalid is not a valid operation. It is here just to detect that a TagOp isn't set.
	TagOpInvalid TagOp = iota

	// TagOpInsert adds the (key, value) to a set if the set doesn't already
	// contain a tag with the same key. Otherwise it is a no-op.
	TagOpInsert

	// TagOpUpdate replaces the (key, value) in a set if the set contains a
	// (key, value) pair with the same key. Otherwise it is a no-op.
	TagOpUpdate

	// TagOpUpsert adds the (key, value) to a set regardless if the set does
	// contain or doesn't contain a (key, value) pair with the same key.
	TagOpUpsert

	// TagOpDelete deletes the (key, value) from a set if it contain a pair
	// with the same key. Otherwise it is a no-op.
	TagOpDelete
)

// TagChange is the interface for tag changes. It is not expected to have
// multiple types implement it. Its main purpose is to only allow read
// operations on its fields and hide its the write operations.
type TagChange interface {
	Key() Key
	Value() []byte
	Op() TagOp
}

// tagChange implements TagChange
type tagChange struct {
	k  Key
	v  []byte
	op TagOp
}

func (tc *tagChange) Key() Key {
	return tc.k
}

func (tc *tagChange) Value() []byte {
	return tc.v
}

func (tc *tagChange) Op() TagOp {
	return tc.op
}

//-----------------------------------------------------------------------------
// Part of tag_set_builder.go
//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------
// The methods below are related to int64 tag types are not supported in v0.1
// and are subject to change.

// InsertInt64 inserts an int64 value 'i' associated with the the key 'k' in
// the tags set being built. If a tag with the same key already exists in the
// tags set being built then this is a no-op.
func (tb *TagSetBuilder) InsertInt64(k *KeyInt64, i int64) *TagSetBuilder {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, uint64(i))
	tb.insertBytes(k, v)
	return tb
}

// UpdateInt64 updates an int64 value 'i' associated with the the key 'k' in
// the tags set being built. If a no tag with the same key is already present
// in the tags set being built then this is a no-op.
func (tb *TagSetBuilder) UpdateInt64(k *KeyInt64, i int64) *TagSetBuilder {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, uint64(i))
	tb.updateBytes(k, v)
	return tb
}

// UpsertInt64 updates or insert an int64 value 'i' associated with the key 'k'
// in the tags set being built.
func (tb *TagSetBuilder) UpsertInt64(k *KeyInt64, i int64) *TagSetBuilder {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, uint64(i))
	tb.upsertBytes(k, v)
	return tb
}

//-----------------------------------------------------------------------------
// The methods below are related to bool tag types are not supported in v0.1
// and are subject to change.

// UpsertBool updates or insert a bool value 'b' associated with the key 'k' in
// the tags set being built.
func (tb *TagSetBuilder) UpsertBool(k *KeyBool, b bool) *TagSetBuilder {
	v := make([]byte, 1)
	if b {
		v[0] = 1
	} else {
		v[1] = 0
	}
	tb.upsertBytes(k, v)
	return tb
}

// UpdateBool updates a bool value 'b' associated with the the key 'k' in the
// tags set being built. If a no tag with the same key is already present in
// the tags set being built then this is a no-op.
func (tb *TagSetBuilder) UpdateBool(k *KeyBool, b bool) *TagSetBuilder {
	v := make([]byte, 1)
	if b {
		v[0] = 1
	} else {
		v[1] = 0
	}
	tb.updateBytes(k, v)
	return tb
}

// InsertBool inserts an bool value 'b' associated with the the key 'k' in the
// tags set being built. If a tag with the same key already exists in the tags
// set being built then this is a no-op.
func (tb *TagSetBuilder) InsertBool(k *KeyBool, b bool) *TagSetBuilder {
	v := make([]byte, 1)
	if b {
		v[0] = 1
	} else {
		v[1] = 0
	}
	tb.insertBytes(k, v)
	return tb
}

//-----------------------------------------------------------------------------
// The methods below are related to TagChange are not supported in v0.1 and are
// subject to change

// Apply applies a set of changes to the tags set being built.
func (tb *TagSetBuilder) Apply(tcs ...TagChange) *TagSetBuilder {
	for _, tc := range tcs {
		switch tc.Op() {
		case TagOpInsert:
			tb.ts.insertBytes(tc.Key(), tc.Value())
		case TagOpUpdate:
			tb.ts.updateBytes(tc.Key(), tc.Value())
		case TagOpUpsert:
			tb.ts.upsertBytes(tc.Key(), tc.Value())
		case TagOpDelete:
			tb.ts.delete(tc.Key())
		default:
			continue
		}
	}
	return tb
}

//-----------------------------------------------------------------------------
// Part of tag_set.go
//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------
// The methods below are related to int64 tag types are not supported in v0.1
// and are subject to change.

// ValueAsInt64 returns the int64 associated with a specified key.
func (ts *TagSet) ValueAsInt64(k Key) (int64, error) {
	if _, ok := k.(*KeyInt64); !ok {
		return 0, fmt.Errorf("values of key '%v' are not of type int64", k.Name())
	}

	b, ok := ts.m[k]
	if !ok {
		return 0, fmt.Errorf("no value assigned to tag key '%v'", k.Name())
	}
	return int64(binary.LittleEndian.Uint64(b)), nil
}

//-----------------------------------------------------------------------------
// The methods below are related to bool tag types are not supported in v0.1
// and are subject to change.

// ValueAsBool returns the bool associated with a specified key.
func (ts *TagSet) ValueAsBool(k Key) (bool, error) {
	if _, ok := k.(*KeyBool); !ok {
		return false, fmt.Errorf("values of key '%v' are not of type bool", k.Name())
	}

	b, ok := ts.m[k]
	if !ok {
		return false, fmt.Errorf("no value assigned to tag key '%v'", k.Name())
	}
	if b[0] == 1 {
		return true, nil
	}
	return false, nil
}

/*

func (bc *bytesCodec) ReadBytes() ([]byte, err) {

	endIdx := bc.ridx+sizeOfUint16
	if end > len(bc.b) {
		return nil, fmt.Errorf("ReadBytes() failed. endIdx=%v, bytes=%v", endIdx, bc.b)
	}

	length :=  binary.LittleEndian.Uint16(valuesSig[idx:])
	idx += sizeOfUint16

	if idx+length > len(valuesSig) {
		return nil, fmt.Errorf("DecodeFromValuesSignature failed. Unexpected signature end '%v' for keys '%v'", valuesSig, keys)
	}

	if length == 0 {
		// No value was encoded for this key
		continue
	}

	ts.m[k] = valuesSig[idx:idx+length]
	idx += length

}

// DecodeFromValuesSignatureToTagSet creates a TagSet from an encoded []byte
// and a slice of keys. The slice of keys is expected to be the same one as the
// one used for encoding.
// This method is intended to be used by the package instrumentation/stats
// library.
func DecodeFromValuesSignature(valuesSig []byte, keys []Key) (*TagSet, error) {
	ts := &TagSet{
		m: make(map[Key][]byte),
	}
	if len(valuesSig) == 0 {
		return ts, nil
	}

	br := bytesReader{valuesSig}
	for _, k := range keys {
		bytes, err := bc.ReadBytes()
		if err != nil {
			return nil, err
		}

		if len(bytes) > 0 {
			// No value was encoded for this key
			continue
		}

		ts.m[k] = bytes
	}

	return ts, nil
}

// EncodeToValuesSignature creates an encoded []byte from TagSet and keys.
// This method is intended to be used by the package instrumentation/stats
// library.
func EncodeToValuesSignature(ts *TagSet, keys []Key) []byte {
	var b buffer
	for _, k := range keys {
		v, ok := ts.m[k]
		if !ok {
			// write 0 (len(value) == 0) meaning no value is encoded for this key.
			b.WriteUint16(0)
			continue
		}
		b.WriteUint16(len(v))
		b.WriteBytes(v)
	}
	return b.bytes
}

// DecodeFromFullSignature creates a TagSet from an encodded []byte. This
// method is intended to be used by the package instrumentation/stats library.
func DecodeFromFullSignature(fullSig []byte) (*TagSet, error) {
	ts := &TagSet{
		m: make(map[Key]Tag),
	}
	if len(fullSig) == 0 {
		return ts, nil
	}

	var t Tag
	var err error
	idx := 0
	for idx < len(fullSig) {
		typ := keyType(fullSig[idx])
		idx++

		switch typ {
		case keyTypeString:
			t = &tagString{}
		case keyTypeInt64:
			t = &tagInt64{}
		case keyTypeBool:
			t = &tagBool{}
		case keyTypeBytes:
			t = &tagBytes{}
		default:
			return nil, fmt.Errorf("DecodeFromFullSignatureToTagSet failed. Key type invalid %v", typ)
		}

		idx, err = t.setKeyFromBytes(fullSig, idx)
		if err != nil {
			return nil, err
		}
		idx, err = t.setValueFromBytes(fullSig, idx)
		if err != nil {
			return nil, err
		}

		ts.m[t.Key()] = t
	}
	return ts, nil
}

// EncodeToFullSignature creates a full signature []byte from TagSet.
// This method is intended to be used by the package instrumentation/stats
// library.
func EncodeToFullSignature(ts *TagSet) []byte {
	b := &buffer{
		bytes: make([]byte, 25*len(ts.m)),
	}
	for _, t := range ts.m {
		b.writeByte(byte(t.Key().Type()))
		t.encodeKeyToBuffer(b)
		t.encodeValueToBuffer(b)
	}
	return b.bytes[:b.writeIdx]
}

func (ts *TagSet) GetTagString(k KeyString) (string, err) {}

func (ts *TagSet) GetTagInt64(k KeyInt64) (int64, err) {}

func (ts *TagSet) GetTagBool(k KeyBool) (bool, err) {}

func (tb *TagSet) insertString(k KeyString, s string) bool {}

func (tb *TagSet) updateString(k KeyString, s string) bool {}

func (tb *TagSet) upsertString(k KeyString, s string) {}

func (tb *TagSet) deleteString(k KeyString, s string) bool {}

func (tb *TagSet) insertInt64(k KeyInt64, i int64) bool {}

func (tb *TagSet) updateInt64(k KeyInt64, i int64) bool {}

func (tb *TagSet) upsertInt64(k KeyInt64, i int64) {}

func (tb *TagSet) deleteInt64(k KeyInt64, i int64) bool {}

func (tb *TagSet) insertBool(k KeyBool, b bool) bool {}

func (tb *TagSet) updateBool(k KeyBool, b bool) bool {}

func (tb *TagSet) upsertBool(k KeyBool, b bool) {}

func (tb *TagSet) deleteBool(k KeyBool, b bool) bool {}







func tagSetFromValuesBytes(vs []byte, ks []Key) *TagSet {
	ts := &TagSet{
		m : make(map[Key][]byte),
	}

	for _, k := range ks {
		v = vs.readValue()
		vs = vs[len(v)+2:]
		if v != nil {
			ts.m[k] = v
		}
	}
	return ts
}




func tagSetFromKeyValuesBytes(kvs keyValueSet) *TagSet {
	ts := &TagSet{
		m : make(map[Key][]byte),
	}

	ks := kvs.keySet
	vs := kvs.valueSet

	for ;len(ks) > 0; {
		k := ks.readValue()
		ks = ks[2:]
		v:= vs.readValue()
		vs = vs[len(v)+2:]
		if bytes != nil {
			ts.m[k] bytes
		}
	}
	return ts
}

type keyValueSet struct {
	keySet []byte
	valueSet []byte
}


func readValue(bytes []byte) []byte {}


func readKey(bytes []byte) key {
	id := *(*uint16)(unsafe.Pointer(&bytes[0]))
	return getKeyByID(id)
}

func writeKeyID(bytes []byte, k key) []byte{
	tmp := *(*[2]byte)(unsafe.Pointer(&k.id))
	copy(bytes[len(bytes), tmp)
	return bytes
}
*/

//-----------------------------------------------------------------------------
// Part of tag.go
//-----------------------------------------------------------------------------

// Tag is the tuple (key, value) used only when extracting []Tag from a TagSet.
type Tag struct {
	K Key
	V []byte
}

//-----------------------------------------------------------------------------
// Part of keys_manager.go
//-----------------------------------------------------------------------------

// CreateKeyInt64 creates or retrieves a key of type keyInt64 with name/ID set
// to the input argument name. Returns an error if a key with the same name
// exists and is of a different type.
// Not supported in v0.1 and are subject to change
func (km *keysManager) createKeyInt64(name string) (*KeyInt64, error) {
	if !validateKeyName(name) {
		return nil, fmt.Errorf("key name %v is invalid", name)
	}
	km.Lock()
	defer km.Unlock()
	k, ok := km.keys[name]
	if ok {
		ki, ok := k.(*KeyInt64)
		if !ok {
			return nil, fmt.Errorf("key with name %v cannot be created/retrieved as type *keyInt64. It was already registered as type %T", name, k)
		}
		return ki, nil
	}

	ki := &KeyInt64{
		name: name,
		id:   km.nextKeyID,
	}
	km.nextKeyID++
	km.keys[name] = ki
	return ki, nil
}

// CreateKeyBool creates or retrieves a key of type keyBool with name/ID set to
// the input argument name. Returns an error if a key with the same name exists
// and is of a different type.
// Not supported in v0.1 and are subject to change
func (km *keysManager) createKeyBool(name string) (*KeyBool, error) {
	if !validateKeyName(name) {
		return nil, fmt.Errorf("key name %v is invalid", name)
	}
	km.Lock()
	defer km.Unlock()
	k, ok := km.keys[name]
	if ok {
		kb, ok := k.(*KeyBool)
		if !ok {
			return nil, fmt.Errorf("key with name %v cannot be created/retrieved as type *keyBool. It was already registered as type %T", name, k)
		}
		return kb, nil
	}

	kb := &KeyBool{
		name: name,
		id:   km.nextKeyID,
	}
	km.nextKeyID++
	km.keys[name] = kb
	return kb, nil
}

//-----------------------------------------------------------------------------
// Part of grpc_codec.go
//-----------------------------------------------------------------------------

type keyType byte

const (
	keyTypeString keyType = iota
	keyTypeInt64
	keyTypeTrue
	keyTypeFalse
)

const (
	tagsVersionID = byte(0)

	srvStatsVersionFieldID   = byte(0)
	srvStatsLatencyNsFieldID = byte(0)

	traceVersionID                     = byte(0)
	traceFieldID                       = byte(0)
	traceSpanFieldID                   = byte(1)
	traceOptionsFieldID                = byte(2)
	traceMaskFieldID                   = byte(0xFC)
	traceInvSamplingProbabilityFieldID = byte(0xFD)
	traceParenSpanFieldID              = byte(0xFE)
)

type GRPCCodec struct {
	buf        []byte
	wIdx, rIdx int
}

func (gc *GRPCCodec) EncodeTagSet(ts *TagSet) []byte {
	gc.buf = make([]byte, len(ts.m))
	gc.wIdx, gc.rIdx = 0, 0

	gc.writeByte(byte(tagsVersionID))
	for k, v := range ts.m {
		if k.Name() == "method" || k.Name() == "caller" {
			continue
		}

		switch k.(type) {
		case *KeyString:
			gc.writeByte(byte(keyTypeString))
			gc.writeStringWithVarintLen(k.Name())
			gc.writeBytesWithVarintLen(v)
			break
		case *KeyInt64:
			gc.writeByte(byte(keyTypeInt64))
			gc.writeStringWithVarintLen(k.Name())
			gc.writeBytes(v)
			break
		case *KeyBool:
			if v[0] == 1 {
				gc.writeTagTrue(k.Name())
			} else {
				gc.writeTagFalse(k.Name())
			}
			break
		default:
			continue
		}
	}

	return gc.buf[:gc.wIdx]
}

func (gc *GRPCCodec) DecodeTagSet(bytes []byte) (*TagSet, error) {
	gc.buf = bytes
	gc.wIdx, gc.rIdx = 0, 0
	ts := newTagSet(0)

	if len(gc.buf) == 0 {
		return ts, nil
	}

	version, err := gc.readByte()
	if err != nil {
		return nil, err
	}
	if version > tagsVersionID {
		return nil, fmt.Errorf("GRPCCodec.DecodeTagSet() doesn't support version %v. Supports only up to: %v", version, tagsVersionID)
	}

	var k Key
	var v []byte

	for !gc.readEnded() {
		typByte, err := gc.readByte()
		if err != nil {
			continue
		}
		typ := keyType(typByte)
		kName, err := gc.readStringWithVarintLen()
		if err != nil {
			continue
		}
		switch typ {
		case keyTypeString:
			v, err = gc.readBytesWithVarintLen()
			if err != nil {
				continue
			}
			k, err = DefaultKeyManager().CreateKeyString(kName)
			if err != nil {
				continue
			}
			break
		case keyTypeInt64:
			v, err = gc.readBytes(8)
			if err != nil {
				continue
			}
			k, err = DefaultKeyManager().CreateKeyInt64(kName)
			if err != nil {
				continue
			}
			break
		case keyTypeFalse:
			v = []byte{0}
			k, err = DefaultKeyManager().CreateKeyBool(kName)
			if err != nil {
				continue
			}
			break
		case keyTypeTrue:
			v = []byte{1}
			k, err = DefaultKeyManager().CreateKeyBool(kName)
			if err != nil {
				continue
			}
			break
		default:
			return nil, fmt.Errorf("toStubbyTagsFormat failed. Key type invalid %v", typ)
		}
		ts.upsertBytes(k, v)
	}
	return ts, nil
}

func (gc *GRPCCodec) growIfRequired(expected int) {
	if len(gc.buf)-gc.wIdx < expected {
		tmp := make([]byte, 2*(len(gc.buf)+1)+expected)
		copy(tmp, gc.buf)
		gc.buf = tmp
	}
}

func (gc *GRPCCodec) writeTagString(k, v string) {
	gc.writeByte(byte(keyTypeString))
	gc.writeStringWithVarintLen(k)
	gc.writeStringWithVarintLen(v)
}

func (gc *GRPCCodec) writeTagUint64(k string, i uint64) {
	gc.writeByte(byte(keyTypeInt64))
	gc.writeStringWithVarintLen(k)
	gc.writeUint64(i)
}

func (gc *GRPCCodec) writeTagTrue(k string) {
	gc.writeByte(byte(keyTypeTrue))
	gc.writeStringWithVarintLen(k)
}

func (gc *GRPCCodec) writeTagFalse(k string) {
	gc.writeByte(byte(keyTypeFalse))
	gc.writeStringWithVarintLen(k)
}

func (gc *GRPCCodec) writeTraceID(low, high uint64) {
	gc.growIfRequired(17)
	gc.writeByte(traceFieldID)
	gc.writeUint64(low)
	gc.writeUint64(high)
}

func (gc *GRPCCodec) writeTraceSpanID(i uint64) {
	gc.growIfRequired(9)
	gc.writeByte(traceSpanFieldID)
	gc.writeUint64(i)
}

func (gc *GRPCCodec) writeTraceOptions(i byte) {
	gc.growIfRequired(2)
	gc.writeByte(traceOptionsFieldID)
	gc.writeByte(i)
}

func (gc *GRPCCodec) writeTraceMask(i uint32) {
	gc.growIfRequired(5)
	gc.writeByte(traceMaskFieldID)
	gc.writeUint32(i)
}

func (gc *GRPCCodec) writeTraceInverseSamplingProbability(f float64) {
	g := float32(f)
	i := *(*uint32)(unsafe.Pointer(&g))
	gc.growIfRequired(5)
	gc.writeByte(traceInvSamplingProbabilityFieldID)
	gc.writeUint32(i)
}

func (gc *GRPCCodec) writeTraceParenSpanID(i uint64) {
	gc.growIfRequired(9)
	gc.writeByte(traceParenSpanFieldID)
	gc.writeUint64(i)
}

func (gc *GRPCCodec) writeServerStatsLatencyNs(i uint64) {
	gc.growIfRequired(9)
	gc.writeByte(srvStatsLatencyNsFieldID)
	gc.writeUint64(i)
}

func (gc *GRPCCodec) writeBytesWithVarintLen(bytes []byte) {
	length := len(bytes)

	gc.growIfRequired(binary.MaxVarintLen64 + length)
	gc.wIdx += binary.PutUvarint(gc.buf[gc.wIdx:], uint64(length))
	copy(gc.buf[gc.wIdx:], bytes)
	gc.wIdx += length
}

func (gc *GRPCCodec) writeStringWithVarintLen(s string) {
	length := len(s)

	gc.growIfRequired(binary.MaxVarintLen64 + length)
	gc.wIdx += binary.PutUvarint(gc.buf[gc.wIdx:], uint64(length))
	copy(gc.buf[gc.wIdx:], s)
	gc.wIdx += length
}

func (gc *GRPCCodec) writeByte(v byte) {
	gc.growIfRequired(1)
	gc.buf[gc.wIdx] = v
	gc.wIdx++
}

func (gc *GRPCCodec) writeBytes(bytes []byte) {
	length := len(bytes)
	gc.growIfRequired(length)
	copy(gc.buf[gc.wIdx:], bytes)
	gc.wIdx += length
}

func (gc *GRPCCodec) writeUint32(i uint32) {
	gc.growIfRequired(4)
	binary.LittleEndian.PutUint32(gc.buf[gc.wIdx:], i)
	gc.wIdx += 4
}

func (gc *GRPCCodec) writeUint64(i uint64) {
	gc.growIfRequired(8)
	binary.LittleEndian.PutUint64(gc.buf[gc.wIdx:], i)
	gc.wIdx += 8
}

func (gc *GRPCCodec) readByte() (byte, error) {
	if len(gc.buf) < gc.rIdx+1 {
		return 0, fmt.Errorf("unexpected end while readByte '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	b := gc.buf[gc.rIdx]
	gc.rIdx++
	return b, nil
}

func (gc *GRPCCodec) readBytes(count int) ([]byte, error) {
	if len(gc.buf) < gc.rIdx+count {
		return nil, fmt.Errorf("unexpected end while readBytes '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	endIdx := gc.rIdx + count
	tmp := gc.buf[gc.rIdx:endIdx]
	gc.rIdx = endIdx
	return tmp, nil
}

func (gc *GRPCCodec) readUint32() (uint32, error) {
	if len(gc.buf) < gc.rIdx+4 {
		return 0, fmt.Errorf("unexpected end while readUint32 '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	i := binary.LittleEndian.Uint32(gc.buf[gc.rIdx:])
	gc.rIdx += 4
	return i, nil
}

func (gc *GRPCCodec) readUint64() (uint64, error) {
	if len(gc.buf) < gc.rIdx+8 {
		return 0, fmt.Errorf("unexpected end while readUint64 '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	i := binary.LittleEndian.Uint64(gc.buf[gc.rIdx:])
	gc.rIdx += 8
	return i, nil
}

func (gc *GRPCCodec) readBytesWithVarintLen() ([]byte, error) {
	if gc.readEnded() {
		return nil, fmt.Errorf("unexpected end while readBytesWithVarintLen '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	length, valueStart := binary.Uvarint(gc.buf[gc.rIdx:])
	if valueStart <= 0 {
		return nil, fmt.Errorf("unexpected end while readBytesWithVarintLen '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}

	valueStart += gc.rIdx
	valueEnd := valueStart + int(length)
	if valueEnd > len(gc.buf) || length < 0 {
		return nil, fmt.Errorf("malformed encoding: length:%v, upper%v, maxLength:%v", length, valueEnd, len(gc.buf))
	}

	gc.rIdx = valueEnd
	return gc.buf[valueStart:valueEnd], nil
}

func (gc *GRPCCodec) readStringWithVarintLen() (string, error) {
	bytes, err := gc.readBytesWithVarintLen()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (gc *GRPCCodec) readEnded() bool {
	return gc.rIdx >= len(gc.buf)
}
