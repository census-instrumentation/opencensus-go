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
	"fmt"
)

// TagValuesSignature is of the form [len_value value]*
// The types and keys are not part of the encoding. It is expected that the
// encoder/decoder provide the same []Keys in order to work as expected.
type TagValuesSignature struct {
	sig []byte
}

// DecodeFromValuesSignatureToSlice creates a []Tag from an encodded []byte and
// a slice of keys. The slice of keys is expected to be the same one as the one
// used for encoding.
func DecodeFromValuesSignatureToSlice(valuesSig []byte, keys []Key) ([]Tag, error) {
	var ts []Tag
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
			return nil, fmt.Errorf("DecodeFromValuesSignatureToSlice failed. Unexpected signature end '%v' for keys '%v'", valuesSig, keys)
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
			return nil, fmt.Errorf("DecodeFromValuesSignatureToSlice failed. Key type invalid %v", k)
		}
		idx, err = t.setValueFromBytesKnownLength(valuesSig, idx, length)
		if err != nil {
			return nil, err
		}

		ts = append(ts, t)
	}
	return ts, nil
}

// DecodeFromValuesSignatureToTagSet creates a TagSet from an encodded []byte
// and a slice of keys. The slice of keys is expected to be the same one as the
// one used for encoding.
func DecodeFromValuesSignatureToTagSet(valuesSig []byte, keys []Key) (*TagSet, error) {
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

// EncodeToValuesSignature creates a TagValuesSignature from TagSet
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
