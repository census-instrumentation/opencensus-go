package tagging

import (
	"bytes"
	"fmt"
)

// TagsFullSignature  is the encoding used for serialization for interprocess
// communication. It is of the form:
// [tag_type key_len key_bytes value_len value_bytes]*
// Where:
//  * tag_type is one byte, and is used to describe the format of value_bytes.
//    In particular, the low 2 bits of this byte are used as follows:
//    00 (value 0): string (UTF-8) encoding
//    01 (value 1): integer (varint int64 encoding). See
//      https://developers.google.com/protocol-buffers/docs/encoding#varints
//      for documentation on the varint format.
//    10 (value 2): boolean format. In this case value_len should equal 1, and
//       the value_bytes will be a single byte containing either 0 (false) or
//       1 (true).
//    11 (value 3): byte sequence. Arbitrary uninterpreted bytes.
//  * The key_len and value_len fields are represented using a varint, with a
//    maximum value of 16383 bytes (this value is guaranteed to fit in at most
//    2 bytes). Zero length keys or values are not allowed.
//  * The value in key_bytes is a US-ASCII format string.
type TagsFullSignature struct {
	sig []byte
}

// DecodeFromFullSignature creates a TagsSet from TagsFullSignature
func DecodeFromFullSignature(fullSig []byte) (TagsSet, error) {
	ts := make(TagsSet)
	if len(fullSig) == 0 {
		return ts, nil
	}

	var k Key
	var t Tag
	idx := int32(0)
	for idx < len(fullSig) {
		typ := keyType(fullSig[idx])

		switch typ {
		case keyTypeStringUTF8:
			k = keyStringUTF8{}
			t = tagStringUTF8{
				keyStringUTF8: &k,
			}
		case keyTypeInt64:
			k = keyInt64{}
			t = tagInt64{
				keyInt64: &k,
			}
		case keyTypeBool:
			k = keyBool{}
			t = tagBool{
				keyBool: &k,
			}
		case keyTypeBytes:
			k = keyBytes{}
			t = tagBytes{
				keyInt64: &k,
			}
		default:
			return nil, fmt.Errorf("TagsFromValuesSignature failed. Key type invalid %v", k)
		}

		idx, err = k.setFromBytes(fullSig, idx)
		if err != nil {
			return nil, err
		}
		idx, err = t.setValueFromBytes(fullSig, idx)
		if err != nil {
			return nil, err
		}

		ts[k] = t
	}
	return ts, nil
}

// EncodeToFullSignature creates a full signature []byte from TagsSet
func EncodeToFullSignature(ts TagsSet) []byte {
	var b bytes.Buffer
	for _, t := range ts {
		t.EncodeFullTagToBuffer(&b)
	}

	return b.Bytes()
}
