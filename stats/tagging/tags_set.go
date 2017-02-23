package tagging

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
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

// TagValuesSignature is of the form (type_value len_value value)*
type TagValuesSignature struct {
	bytes []byte
}

// TagsFullSignature  is of the form (type_value len_value value)*
type TagsFullSignature struct {
	bytes []byte
}

// TagsFromValuesSignature decodes a []byte signature to a []Tag when the keys
// are not part of the encoding.
func TagsFromValuesSignature(valuesSig TagValuesSignature, keys []Key) ([]Tag, error) {
	var tags []Tag
	idx := 0
	for _, k := range keys {
		l := int32FromBytes(valuesSig[idx:])
		switch typ := k.(type) {
		case *keyString:
		case *keyInt64:
		case *keyFloat64:
		case *keyBool:

		}
	}

		t, err := k.createTag(valuesSig[idx:])
		if err != nil {
			return nil, err
		}
		idx += t.Vlength()
		v, idx, err := readVarintString(valuesSig)
		if err != nil {
			return nil, err
		}
		valuesSig = valuesSig[idx:]
		if len(v) == 0 {
			continue
		}

		tags = append(tags, Key.createKey(v))
	}
	return tags, nil
}

func TagsFromFullSignature(fullsig TagsFullSignature) ([]Tag, error) {
	return nil, nil
}

func TagsToValuesSignature(tags []Tag, keys []Key) (TagValuesSignature, error) {
}

func TagsToFullSignature(tags []Tag) (TagsFullSignature, error) {
}

func float64FromBytes(bytes []byte) float64 {
	f := math.Float64frombits(binary.LittleEndian.Uint64(bytes))
	return f
}

func float64ToBytes(f float64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, math.Float64bits(f))
	return bytes
}

func int64FromBytes(bytes []byte) int64 {
	return binary.LittleEndian.Uint64(bytes)
}

func int64ToBytes(i int64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, i)
	return bytes
}

func int32FromBytes(bytes []byte) int32 {
	return binary.LittleEndian.Uint32(bytes)
}

func int32ToBytes(i int32) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.Uint32(bytes, i)
	return bytes
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
