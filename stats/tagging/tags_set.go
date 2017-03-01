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

// TagsSet is the tags set representation in the context.
type TagsSet map[Key]Tag

// ApplyMutation applies a single mutation to the TagsSet
func (ts TagsSet) ApplyMutation(m Mutation) {
	t := m.Tag()
	k := t.Key()
	switch m.Behavior() {
	case BehaviorReplace:
		if _, ok := ts[k]; ok {
			ts[k] = t
		}
	case BehaviorAdd:
		if _, ok := ts[k]; !ok {
			ts[k] = t
		}
	case BehaviorAddOrReplace:
		ts[k] = t
	default:
		panic(fmt.Sprintf("mutation type is %v. This is a bug and should never happen.", m.Behavior()))
	}
}

// ApplyMutations applies multiple mutations to the TagsSet
func (ts TagsSet) ApplyMutations(ms ...Mutation) {
	for _, m := range ms {
		ts.ApplyMutation(m)
	}
}


// TagsFromValuesSignature decodes a []byte signature to a []Tag when the keys
// are not part of the encoding.
func TagsFromValuesSignature(bytes TagValuesSignature, keys []Key) ([]Tag, error) {
	var tags []Tag
	if len(bytes) == 0 {
		return tags, nil
	}
	idx := int32(0)
	for _, k := range keys {
		len, err := lengthFromBytes(bytes[idx:])
		if err != nil {
			return nil, err
		}
		idx += 4
		if len == 0 {
			continue
		}

		switch typ := k.(type) {
		case *keyString:
			v, err := stringFromBytes(bytes[idx:], len)
			if err != nil {
				return nil, err
			}
			idx += len
			tags = append(tags, typ.CreateTag(v))
		case *keyInt64:
			v, err := int64FromBytes(bytes[idx:])
			if err != nil {
				return nil, err
			}
			idx += 8
			tags = append(tags, typ.CreateTag(v))
		case *keyBool:
			v, err := boolFromBytes(bytes[idx:])
			if err != nil {
				return nil, err
			}
			idx++
			tags = append(tags, typ.CreateTag(v))
		default:
			return nil, fmt.Errorf("TagsFromValuesSignature failed. Key type invalid %v", k)
		}
	}
	return tags, nil
}

// TagsSetFromValuesSignature decodes a []byte signature to a []Tag when the keys
// are not part of the encoding.
func TagsSetFromValuesSignature(bytes TagValuesSignature, keys []Key) (TagsSet, error) {
	var ts TagsSet
	if len(bytes) == 0 {
		return ts, nil
	}
	ts = make(TagsSet)
	idx := int32(0)
	for _, k := range keys {
		len, err := lengthFromBytes(bytes[idx:])
		if err != nil {
			return nil, err
		}
		idx += 4
		if len == 0 {
			continue
		}

		switch typ := k.(type) {
		case *keyString:
			v, err := stringFromBytes(bytes[idx:], len)
			if err != nil {
				return nil, err
			}
			idx += len
			ts[typ] = typ.CreateTag(v)
		case *keyInt64:
			v, err := int64FromBytes(bytes[idx:])
			if err != nil {
				return nil, err
			}
			idx += 8
			ts[typ] = typ.CreateTag(v)
		case *keyBool:
			v, err := boolFromBytes(bytes[idx:])
			if err != nil {
				return nil, err
			}
			idx++
			ts[typ] = typ.CreateTag(v)
		default:
			return nil, fmt.Errorf("TagsFromValuesSignature failed. Key type invalid %v", k)
		}
	}
	return ts, nil
}

// TagsToValuesSignature encodes the subset of values corresponding to the
// keys. Values are encoded wihtout their keys and without any type
// information. To decode it is expected to provide the list of keys in the
// same orders as the one used for encoding.
func (ts TagsSet) TagsToValuesSignature(keys []Key) TagValuesSignature {
	var b bytes.Buffer
	for _, k := range keys {
		t, ok := ts[k]
		if !ok {
			b.Write(int32ToBytes(0))
			continue
		}
		t.WriteValueToBuffer(&b)
	}
	return b.Bytes()
}

/*
func TagsFromFullSignature(fullsig TagsFullSignature) ([]Tag, error) {
	return nil, nil
}
*/

func (ts TagsSet) TagsToFullSignature() TagsFullSignature {
	var b bytes.Buffer
	for _, k := range keys {
		t, ok := ts[k]
		if !ok {
			WriteKeyToBuffer(k, &b)
			b.Write(int32ToBytes(0))
			continue
		}
		t.WriteKeyValueToBuffer(&b)
	}
	return b.Bytes()
}

/*
func validateTag(t Tag) error {
	// TODO(iamm2): Do validation checks. Length of key and value are
	// expected to be < 256 bytes, and can only contain printable
	// characters.
	if !valid(t.Key) || !valid(t.Value) {
		return fmt.Errorf("invalid census tag key: %q or value: %q", t.Key, t.Value)
	}
	return nil
}


// decodeFromValuesSignatureToMap decodes a []byte signature to a contextTags
// when the keys are not part of the encoding.
func decodeFromValuesSignatureToMap(valuesSig []byte, keys []string) (contextTags, error) {
	ct := make(contextTags)
	for _, k := range keys {
		v, idx, err := readVarintString(valuesSig)
		if err != nil {
			return nil, err
		}
		valuesSig = valuesSig[idx:]
		if len(v) == 0 {
			continue
		}

		ct[k] = v
	}
	return ct, nil
}

// decodeFromFullSignatureToMap decodes a []byte signature to a contextTags
// when the keys are part of the encoding.
func decodeFromFullSignatureToMap(fullSig []byte) (contextTags, error) {
	ct := make(contextTags)

	for len(fullSig) > 0 {
		key, idx, err := readVarintString(fullSig)
		if err != nil {
			return nil, err
		}
		fullSig = fullSig[idx:]

		val, idx, err := readVarintString(fullSig)
		if err != nil {
			return nil, err
		}
		fullSig = fullSig[idx:]

		ct[key] = val
	}
	return ct, nil
}

// encodeToFullSignatureWithPrefix is used to encode the contextTags
// (map[string]string) to the wire format that is used in the protobuf to pass
// context information between remote tasks. This is the same format used by
// the other languages (Java, C, C++...)
func (ct contextTags) encodeToFullSignatureWithPrefix() []byte {
	var keys []string
	for k := range ct {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	tmp := make([]byte, binary.MaxVarintLen64)

	varIntSize := binary.PutVarint(tmp, int64(len(ct)))
	buf.Write(tmp[:varIntSize]) // writing number of tags as prefix

	for _, k := range keys {
		v := ct[k]
		varIntSize = binary.PutVarint(tmp, int64(len(k)))
		buf.Write(tmp[:varIntSize]) // writing keyLen
		buf.WriteString(k)          // keyLen

		varIntSize = binary.PutVarint(tmp, int64(len(v)))
		buf.Write(tmp[:varIntSize]) // valLen
		buf.WriteString(v)          // writing value
	}
	return buf.Bytes()
}

// decodeFromFullSignatureWithPrefixToMap decodes a []byte signature to a contextTags
// when the keys are part of the encoding as well as the number of tags encoded.
func decodeFromFullSignatureWithPrefixToMap(fullSig []byte) (contextTags, error) {
	tmp := fullSig

	if len(fullSig) == 0 {
		return nil, nil
	}

	count, idx := binary.Varint(fullSig)
	if count < 0 || (count > 0 && idx >= len(fullSig)) {
		return nil, fmt.Errorf("malformed encoding: count:%v, idx%v, len(fullSig):%v", count, idx, len(fullSig))
	}

	ct := make(contextTags, count)

	fullSig = fullSig[idx:]
	for len(fullSig) > 0 {
		key, idx, err := readVarintString(fullSig)
		if err != nil {
			return nil, err
		}
		fullSig = fullSig[idx:]

		val, idx, err := readVarintString(fullSig)
		if err != nil {
			return nil, err
		}
		fullSig = fullSig[idx:]

		ct[key] = val
	}

	if len(ct) != int(count) {
		return nil, fmt.Errorf("malformed encoding. got %v tags, want %v tags (sig: %v)", len(ct), count, tmp)
	}
	return ct, nil
}

/*
// A Tag is the (key,value) pair that the client code uses to tag a
// measurement.
type Tag struct {
	Key, Value string
}

func TagsFromSignature(signature []byte, keys []string) ([]Tag, error) {
	if len(keys) == 0 {
		return decodeFromFullSignatureToSlice(signature)
	}
	return decodeFromValuesSignatureToSlice(signature, keys)
}

// decodeFromFullSignatureToSlice decodes a []byte signature to a []Tag when
// the keys are part of the encoding.
func decodeFromFullSignatureToSlice(fullSig []byte) ([]Tag, error) {
	var tags []Tag

	for len(fullSig) > 0 {
		key, idx, err := readVarintString(fullSig)
		if err != nil {
			return nil, err
		}
		fullSig = fullSig[idx:]

		val, idx, err := readVarintString(fullSig)
		if err != nil {
			return nil, err
		}
		fullSig = fullSig[idx:]

		tags = append(tags, Tag{key, val})
	}
	return tags, nil
}




// TagsSet holds the census tags and values.
type TagsSet map[string]string

// EncodeToValuesSignature is used in the usageCollection to convert the
// TagsSet (map[string]string) to a string that can be used as map keys. It
// is used by for views wher ethe list of keys is known before hand (all views
// except the "all tags views"). It is twice as fast as EncodeToFullSignature.
func (ts TagsSet) EncodeToValuesSignature(specificKeys []string) string {
	var buf bytes.Buffer
	tmp := make([]byte, binary.MaxVarintLen64)
	for _, k := range specificKeys {
		v, ok := ts[k]
		if !ok {
			varIntSize := binary.PutVarint(tmp, 0)
			buf.Write(tmp[:varIntSize])
			continue
		}
		varIntSize := binary.PutVarint(tmp, int64(len(v)))
		buf.Write(tmp[:varIntSize])
		buf.WriteString(v)
	}
	return buf.String()
}

// EncodeToFullSignature is used in the usageCollection to convert the
// TagsSet (map[string]string) to a string that can be used as map keys.
// It is only used for the "all tags views" as the keys are not known ahead
// of time. The encoding is very similar to the on-wire encoding used between
// tasks. The format is: [key_len key_bytes value_len value_bytes]*, where
// key_len and value_len are varint encoded.
func (ts TagsSet) EncodeToFullSignature() string {
	var keys []string
	for k := range ts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	tmp := make([]byte, binary.MaxVarintLen64)
	for _, k := range keys {
		v := ts[k]
		varIntSize := binary.PutVarint(tmp, int64(len(k)))
		buf.Write(tmp[:varIntSize]) // writing keyLen
		buf.WriteString(k)          // keyLen

		varIntSize = binary.PutVarint(tmp, int64(len(v)))
		buf.Write(tmp[:varIntSize]) // valLen
		buf.WriteString(v)          // writing value
	}
	return buf.String()
}
*/
