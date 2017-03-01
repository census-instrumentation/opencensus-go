package tagging

import (
	"bytes"
	"fmt"
)

// TagValuesSignature is of the form [len_value value]*
// The types and keys are not part of the encoding. It is expected that the
// encoder/decoder provide the same []Keys in order to work as expected.
type TagValuesSignature struct {
	sig []byte
}

// DecodeFromValuesSignature creates a TagsSet from an TagValuesSignature and a
// slice of keys. The slice of keys is expected to be the same one as the one
// used for encoding.
func DecodeFromValuesSignature(valuesSig []byte, keys []Key) (TagsSet, error) {
	ts := make(TagsSet)
	if len(valuesSig) == 0 {
		return ts, nil
	}

	var t Tag
	idx := 0
	len := 0
	for _, k := range keys {
		if idx > len(valuesSig) {
			return nil, fmt.Errorf("DecodeFromValuesSignature failed. Unexpected signature end '%v' for keys '%v'", valuesSig, keys)
		}
		len, idx = readVarint(valuesSig, idx)
		if len == 0 {
			// No value was encoded for this key
			continue
		}

		switch k.Type() {
		case keyTypeStringUTF8:
			t = tagStringUTF8{
				keyStringUTF8: k,
			}
		case keyTypeInt64:
			t = tagInt64{
				keyInt64: k,
			}
		case keyTypeBool:
			t = tagBool{
				keyBool: k,
			}
		case keyTypeBytes:
			t = tagBytes{
				keyInt64: k,
			}
		default:
			return nil, fmt.Errorf("TagsFromValuesSignature failed. Key type invalid %v", k)
		}
		idx, err = t.setValueFromBytesKnownLength(valuesSig, len, idx)
		if err != nil {
			return nil, error
		}

		ts[k] = t
	}
	return ts, nil
}

// EncodeToValuesSignature creates a TagValuesSignature from TagsSet
func EncodeToValuesSignature(ts TagsSet, keys []Key) []byte {
	var b bytes.Buffer
	for k := range keys {
		if t, ok := ts[k]; !ok {
			// write 0 (len(value) = 0) meaning no value is encoded for this key.
			continue
		}
		t.EncodeValueToBuffer(&b)
	}
	return b.Bytes()
}
