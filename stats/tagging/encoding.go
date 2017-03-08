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
)

// decodeVarintString read the length of a string encoded as varint in btags,
// then reads the string itself from btags. It ensures that all reads are
// within the boundaries of the slice to avoid a panic. Returns
func decodeVarintString(fullSig []byte, idx int) (string, int, error) {
	b, valueEnd, err := decodeVarintBytes(fullSig, idx)
	if err != nil {
		return "", 0, err
	}
	return string(b), valueEnd, nil
}

func decodeVarintBytes(fullSig []byte, idx int) ([]byte, int, error) {
	if idx > len(fullSig) {
		return nil, 0, fmt.Errorf("unexpected end while decodeVarintBytes '%x' starting at idx '%v'", fullSig, idx)
	}
	length, valueStart := binary.Varint(fullSig[idx:])
	if valueStart <= 0 {
		return nil, 0, fmt.Errorf("unexpected end while decodeVarintBytes '%x' starting at idx '%v'", fullSig, idx)
	}

	valueStart += idx
	valueEnd := valueStart + int(length)
	if valueEnd > len(fullSig) || length < 0 {
		return nil, 0, fmt.Errorf("malformed encoding: length:%v, upper%v, maxLength:%v", length, valueEnd, len(fullSig))
	}
	return fullSig[valueStart:valueEnd], valueEnd, nil
}

func decodeVarintInt64(fullSig []byte, idx int) (int64, int, error) {
	if idx > len(fullSig) {
		return 0, -1, fmt.Errorf("unexpected end while decodeVarintInt64 '%x' starting at idx '%v'", fullSig, idx)
	}
	length, readBytes := binary.Varint(fullSig[idx:])
	if readBytes <= 0 {
		return 0, -1, fmt.Errorf("unexpected end while decodeVarintInt64 '%x' starting at idx '%v'", fullSig, idx)
	}

	valueStart := readBytes + idx
	valueEnd := valueStart + int(length)
	if valueEnd > len(fullSig) || length < 0 {
		return 0, -1, fmt.Errorf("malformed encoding: length:%v, upper%v, maxLength:%v", length, valueEnd, len(fullSig))
	}

	i, readBytes := binary.Varint(fullSig[valueStart:])
	if valueStart+readBytes != valueEnd {
		return 0, 1, fmt.Errorf("unexpected end while decodeVarintInt64 '%x' starting at idx '%v'", fullSig, idx)
	}

	return i, valueEnd, nil
}

func decodeVarint(sig []byte, idx int) (l int, newIdx int, err error) {
	if idx >= len(sig) {
		return 0, 0, fmt.Errorf("unexpected end while decodeVarint '%x' starting at idx '%v'", sig, idx)
	}
	length, valueStart := binary.Varint(sig[idx:])
	if valueStart <= 0 {
		return 0, 0, fmt.Errorf("unexpected end while decodeVarint '%x' starting at idx '%v'", sig, idx)
	}
	return int(length), valueStart + idx, nil
}

func encodeVarintString(dst *bytes.Buffer, s string) {
	encodeVarint(dst, int16(len(s)))
	dst.Write([]byte(s))
}

func encodeVarintBytes(dst *bytes.Buffer, b []byte) {
	encodeVarint(dst, int16(len(b)))
	dst.Write(b)
}

func encodeVarintInt64(dst *bytes.Buffer, i int64) {
	tmp := make([]byte, binary.MaxVarintLen64)
	varIntSize := binary.PutVarint(tmp, i)
	encodeVarint(dst, int16(varIntSize))
	dst.Write(tmp[:varIntSize])
}

func encodeVarint(dst *bytes.Buffer, i int16) {
	tmp := make([]byte, binary.MaxVarintLen16)
	varIntSize := binary.PutVarint(tmp, int64(i))
	dst.Write(tmp[:varIntSize])
}
