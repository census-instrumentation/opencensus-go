package tagging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

type Tags map[Key]Mutation

func (t Tags) ApplyMutation(m Mutation) {
	k := m.Key()
	switch m.Behavior() {
	case BehaviorReplace:
		if v, ok := t[k]; ok {
			t[k] = v
		}
	case BehaviorAdd:
		if v, ok := t[k]; !ok {
			t[k] = v
		}
	case BehaviorAddOrReplace:
		t[k] = m
	default:
		panic(fmt.Sprintf("mutation type is %v. This is a bug and should never happen.", m.Behavior()))
	}
}

func (t Tags) ApplyMutations(ms ...Mutation) {
	for _, m := range ms {
		t.ApplyMutation(m)
	}
}

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

// decodeFromValuesSignatureToSlice decodes a []byte signature to a []Tag when
// the keys are not part of the encoding.
func decodeFromValuesSignatureToSlice(valuesSig []byte, keys []string) ([]Tag, error) {
	var tags []Tag
	for _, k := range keys {
		v, idx, err := readVarintString(valuesSig)
		if err != nil {
			return nil, err
		}
		valuesSig = valuesSig[idx:]
		if len(v) == 0 {
			continue
		}

		tags = append(tags, Tag{k, v})
	}
	return tags, nil
}

// readVarintString read the length of a string encoded as varint in btags,
// then reads the string itself from btags. It ensures that all reads are
// within the boundaries of the slice to avoid a panic. Returns
func readVarintString(btags []byte) (string, int, error) {
	if len(btags) == 0 {
		return "", 0, errors.New("btags is empty")
	}

	length, valueStart := binary.Varint(btags)
	valueEnd := valueStart + int(length)
	if valueEnd > len(btags) || length < 0 {
		return "", 0, fmt.Errorf("malformed encoding: length:%v, upper%v, maxLength:%v", length, valueEnd, len(btags))
	}

	value := btags[valueStart:valueEnd]
	return string(value), valueEnd, nil
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
