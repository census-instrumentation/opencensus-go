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

// DecodeFromFullSignatureToSlice creates a []Tag] from an encodded []byte.
func DecodeFromFullSignatureToSlice(fullSig []byte) ([]Tag, error) {
	var ts []Tag
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
			return nil, fmt.Errorf("TagsFromValuesSignature failed. Key type invalid %v", typ)
		}

		idx, err = t.setKeyFromBytes(fullSig, idx)
		if err != nil {
			return nil, err
		}
		idx, err = t.setValueFromBytes(fullSig, idx)
		if err != nil {
			return nil, err
		}

		ts = append(ts, t)
	}
	return ts, nil
}

// DecodeFromFullSignatureToTagsSet creates a TagsSet from an encodded []byte.
func DecodeFromFullSignatureToTagsSet(fullSig []byte) (*TagsSet, error) {
	ts := &TagsSet{
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
			return nil, fmt.Errorf("TagsFromValuesSignature failed. Key type invalid %v", typ)
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

// EncodeToFullSignature creates a full signature []byte from TagsSet
func EncodeToFullSignature(ts *TagsSet) []byte {
	var b bytes.Buffer
	for _, t := range ts.m {
		b.WriteByte(byte(t.Key().Type()))
		t.encodeKeyToBuffer(&b)
		t.encodeValueToBuffer(&b)
	}

	return b.Bytes()
}
