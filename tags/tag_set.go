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

// TagSet is the object holding the tags stored in context. It is not meant to
// be created manually by code outside the library. It should only be created
// using the TagSetBuilder class.
type TagSet struct {
	m map[Key][]byte
}

func newTagSet(size int) *TagSet {
	return &TagSet{
		m: make(map[Key][]byte, size),
	}
}

func (ts *TagSet) toValuesBytes(ks []Key) *valuesBytes {
	vb := &valuesBytes{
		buf: make([]byte, len(ks)),
	}
	for _, k := range ks {
		v := ts.m[k]
		vb.writeValue(v)
	}
	return vb
}

func (ts *TagSet) insertBytes(k Key, b []byte) bool {
	if _, ok := ts.m[k]; ok {
		return false
	}
	ts.m[k] = b
	return true
}

func (ts *TagSet) updateBytes(k Key, b []byte) bool {
	if _, ok := ts.m[k]; !ok {
		return false
	}
	ts.m[k] = b
	return true
}

func (ts *TagSet) upsertBytes(k Key, b []byte) {
	ts.m[k] = b
}

func (ts *TagSet) delete(k Key) {
	delete(ts.m, k)
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
