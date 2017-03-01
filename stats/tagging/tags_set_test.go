package tagging

import (
	"reflect"
	"testing"
)

func TestEncodeDecodeValuesSignature(t *testing.T) {
	type testData struct {
		tagsSet   []Tag
		keys      []Key
		wantSlice []Tag
	}

	k1, _ := DefaultKeyManager().CreateKeyString("k1")
	k2, _ := DefaultKeyManager().CreateKeyString("k2")
	k3, _ := DefaultKeyManager().CreateKeyString("k3")

	testSet := []testData{
		{
			[]Tag{},
			[]Key{k1},
			nil,
		},
		{
			[]Tag{},
			[]Key{k1},
			nil,
		},
		{
			[]Tag{k2.CreateTag("v2")},
			[]Key{k1},
			nil,
		},
		{
			[]Tag{k2.CreateTag("v2")},
			[]Key{k2},
			[]Tag{k2.CreateTag("v2")},
		},
		{
			[]Tag{k1.CreateTag("v1"), k2.CreateTag("v2")},
			[]Key{k1},
			[]Tag{k1.CreateTag("v1")},
		},
		{
			[]Tag{k2.CreateTag("v2"), k1.CreateTag("v1")},
			[]Key{k1},
			[]Tag{k1.CreateTag("v1")},
		},
		{
			[]Tag{k1.CreateTag("v1"), k2.CreateTag("v2"), k3.CreateTag("v3")},
			[]Key{k3, k1},
			[]Tag{k3.CreateTag("v3"), k1.CreateTag("v1")},
		},
	}

	for _, td := range testSet {
		ts := make(TagsSet)
		for _, t := range td.tagsSet {
			ts[t.Key()] = t
		}

		encoded := ts.TagsToValuesSignature(td.keys)

		decodedSlice, err := TagsFromValuesSignature(encoded, td.keys)
		if err != nil {
			t.Errorf("got error %v, want no error when decoding to slice encoded %v", err, td)
		}

		if !reflect.DeepEqual(decodedSlice, td.wantSlice) {
			t.Errorf("got %v, want %v when decoding to slice encoded %v", decodedSlice, td.wantSlice, td)
		}

		decodedMap, err := TagsSetFromValuesSignature(encoded, td.keys)
		if err != nil {
			t.Errorf("got error %v while decoding to map encoded %v, want no error", err, td)
		}

		if len(decodedSlice) != len(decodedMap) {
			t.Errorf("got len(decodedSlice) %v different than len(decodedMap) %v, want them equal when decoding %v", decodedSlice, decodedMap, td)
		}

		for _, tag := range decodedSlice {
			v, ok := decodedMap[tag.Key()]
			if !ok {
				t.Errorf("got key %v in decodedSlice not found in decodedMap %v , want them equivalent when decoding %v", tag.Key, decodedMap, td)
			}
			if v != tag {
				t.Errorf("got %v in decodedSlice different than in decodedMap %v for key %v, want the same when decoding %v", tag, v, tag.Key(), td)
			}
		}
	}
}

func TestEncodeDecodeFullSignature(t *testing.T) {
	type testData struct {
		tagsSet   []Tag
		wantSlice []Tag
	}

	k1, _ := DefaultKeyManager().CreateKeyString("k1")
	k2, _ := DefaultKeyManager().CreateKeyString("k2")
	k3, _ := DefaultKeyManager().CreateKeyString("k3")

	testSet := []testData{
		{
			[]Tag{},
			nil,
		},
		{
			[]Tag{k1.CreateTag("v1")},
			[]Tag{k1.CreateTag("v1")},
		},
		{
			[]Tag{k1.CreateTag("v1"), k2.CreateTag("v2")},
			[]Tag{k1.CreateTag("v1"), k2.CreateTag("v2")},
		},
		{
			[]Tag{k3.CreateTag("v3"), k2.CreateTag("v2"), k1.CreateTag("v1")},
			[]Tag{k1.CreateTag("v1"), k2.CreateTag("v2"), k3.CreateTag("v3")},
		},
	}

	for _, td := range testSet {
		ts := make(TagsSet)
		for _, t := range td.tagsSet {
			ts[t.Key()] = t
		}

		encoded := ts.TagsToFullSignature()

		decodedSlice, err := decodeFromFullSignatureToSlice([]byte(encoded))
		if err != nil {
			t.Errorf("got error %v, want no error when decoding to slice encoded %v", err, td)
		}

		if !reflect.DeepEqual(decodedSlice, td.wantSlice) {
			t.Errorf("got %v, want %v when decoding to slice encoded %v", decodedSlice, td.wantSlice, td)
		}

		decodedMap, err := decodeFromFullSignatureToMap([]byte(encoded))
		if err != nil {
			t.Errorf("got error %v while decoding to map encoded %v, want no error", err, td)
		}

		if len(decodedSlice) != len(decodedMap) {
			t.Errorf("got len(decodedSlice) %v different than len(decodedMap) %v, want them equal when decoding %v", decodedSlice, decodedMap, td)
		}

		for _, tag := range decodedSlice {
			v, ok := decodedMap[tag.Key]
			if !ok {
				t.Errorf("got key %v in decodedSlice not found in decodedMap %v , want them equivalent when decoding %v", tag.Key, decodedMap, td)
			}
			if v != tag.Value {
				t.Errorf("got %v in decodedSlice different than in decodedMap %v for key %v, want the same when decoding %v", tag.Value, v, tag.Key, td)
			}
		}
	}
}
