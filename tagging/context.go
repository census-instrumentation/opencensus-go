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
	"encoding/binary"
	"fmt"
	"sort"

	"golang.org/x/net/context"
)

type censusKey struct{}

// contextTags holds the census tags and values.
type contextTags map[string]string

func valid(s string) bool {
	return true
}

func validateTag(t Tag) error {
	// TODO(iamm2): Do validation checks. Length of key and value are
	// expected to be < 256 bytes, and can only contain printable
	// characters.
	if !valid(t.Key) || !valid(t.Value) {
		return fmt.Errorf("invalid census tag key: %q or value: %q", t.Key, t.Value)
	}
	return nil
}

// NewContextWithTags creates a new census.Context from context and adds the
// tags to it.
func NewContextWithTags(ctx context.Context, tags ...Tag) (context.Context, error) {
	parentTags, _ := ctx.Value(censusKey{}).(contextTags)

	newCt, err := newContextTags(parentTags, tags...)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, censusKey{}, newCt), nil
}

func newContextTags(old contextTags, newTags ...Tag) (contextTags, error) {
	newCt := make(contextTags)
	for k, v := range old {
		newCt[k] = v
	}

	for _, t := range newTags {
		if err := validateTag(t); err != nil {
			return nil, err
		}
		newCt[t.Key] = t.Value
	}
	return newCt, nil
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
