package tags

import "fmt"

// TagSet is the object holding the tags stored in context.
type TagSet struct {
	m map[Key][]byte
}

func (ts *TagSet) Tags() []Tag {
}

func newTagSet(size int) *TagSet {
	return &TagSet{
		m: make(map[Key][]byte, size),
	}
}

// DecodeFromValuesSignatureToTagSet creates a TagSet from an encoded []byte
// and a slice of keys. The slice of keys is expected to be the same one as the
// one used for encoding.
// This method is intended to be used by the package instrumentation/stats
// library.
func DecodeFromValuesSignature(valuesSig []byte, keys []Key) (*TagSet, error) {
	ts := &TagSet{
		m: make(map[Key]Tag),
	}
	if len(valuesSig) == 0 {
		return ts, nil
	}

	var (
		t      Tag
		err    error
		idx    int
		length int
	)
	for _, k := range keys {
		if idx > len(valuesSig) {
			return nil, fmt.Errorf("DecodeFromValuesSignatureToTagSet failed. Unexpected signature end '%v' for keys '%v'", valuesSig, keys)
		}
		if length, idx, err = decodeVarint(valuesSig, idx); err != nil {
			return nil, err
		}

		if length == 0 {
			// No value was encoded for this key
			continue
		}

		switch typ := k.(type) {
		case *keyStringUTF8:
			t = &tagStringUTF8{
				k: typ,
			}
		case *keyInt64:
			t = &tagInt64{
				k: typ,
			}
		case *keyBool:
			t = &tagBool{
				k: typ,
			}
		case *keyBytes:
			t = &tagBytes{
				k: typ,
			}
		default:
			return nil, fmt.Errorf("DecodeFromValuesSignatureToTagSet failed. Key type invalid %v", k)
		}
		idx, err = t.setValueFromBytesKnownLength(valuesSig, idx, length)
		if err != nil {
			return nil, err
		}

		ts.m[k] = t
	}
	return ts, nil
}

// EncodeToValuesSignature creates an encoded []byte from TagSet and keys.
// This method is intended to be used by the package instrumentation/stats
// library.
func EncodeToValuesSignature(ts *TagSet, keys []Key) []byte {
	b := &buffer{
		bytes: make([]byte, 10*len(keys)),
	}
	for _, k := range keys {
		t, ok := ts.m[k]
		if !ok {
			// write 0 (len(value) = 0) meaning no value is encoded for this key.
			b.writeZero()
			continue
		}
		t.encodeValueToBuffer(b)
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
		case keyTypeStringUTF8:
			t = &tagStringUTF8{}
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

func (ts *TagSet) insertBytes(k Key, b []byte) bool {}

func (ts *TagSet) updateBytes(k Key, b []byte) bool {}

func (ts *TagSet) upsertBytes(k Key, b []byte) {}

func (ts *TagSet) deleteBytes(k Key, b []byte) bool {}

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
