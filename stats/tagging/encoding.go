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
		return nil, 0, fmt.Errorf("unexpected end while decodeVarintString '%x' starting at idx '%v'", fullSig, idx)
	}
	length, valueStart := binary.Varint(fullSig[idx:])
	if valueStart <= 0 {
		return nil, 0, fmt.Errorf("unexpected end while decodeVarintString '%x' starting at idx '%v'", fullSig, idx)
	}

	valueStart += idx
	valueEnd := valueStart + int(length)
	if valueEnd > len(fullSig) || length < 0 {
		return nil, 0, fmt.Errorf("malformed encoding: length:%v, upper%v, maxLength:%v", length, valueEnd, len(fullSig))
	}
	return fullSig[valueStart:valueEnd], valueEnd, nil
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

func encodeVarint(dst *bytes.Buffer, i int16) {
	tmp := make([]byte, binary.MaxVarintLen16)
	varIntSize := binary.PutVarint(tmp, int64(i))
	dst.Write(tmp[:varIntSize])
}

/*
func decodeInt64(fullSig []byte, idx int) (int64, error) {
	if len(bytes) < 8 {
		return 0, fmt.Errorf("[]bytes not large enough to decode int64FromBytes: %v", bytes)
	}
	return int64(binary.LittleEndian.Uint64(bytes)), nil
}

func stringFromBytes(bytes []byte, length int32) (string, error) {
	if int32(len(bytes)) < length {
		return "", fmt.Errorf("[]bytes not large enough to decode stringFromBytes: %v", bytes)
	}
	return string(bytes[:length]), nil
}


func lengthFromBytes(bytes []byte) (int32, error) {
	return int32FromBytes(bytes)
}

func stringFromBytes(bytes []byte, length int32) (string, error) {
	if int32(len(bytes)) < length {
		return "", fmt.Errorf("[]bytes not large enough to decode stringFromBytes: %v", bytes)
	}
	return string(bytes[:length]), nil
}

func boolFromBytes(bytes []byte) (bool, error) {
	if len(bytes) < 1 {
		return false, errors.New("[]bytes not large enough to decode boolFromBytes")
	}
	return bytes[0] == 1, nil
}

func typeFromBytes(bytes []byte) (keyType, error) {
	if len(bytes) < 1 {
		return keyTypeStringUTF8, errors.New("[]bytes not large enough to decode typeFromBytes")
	}
	switch keyType(bytes[0]) {
	case keyTypeStringUTF8, keyTypeBool, keyTypeInt64:
		return keyType(bytes[0]), nil
	default:
		return keyType(bytes[0]), fmt.Errorf("unknow keyType: %v", bytes[0])
	}
}

func int32FromBytes(bytes []byte) (int32, error) {
	if len(bytes) < 4 {
		return 0, fmt.Errorf("[]bytes not large enough to decode int32FromBytes: %v", bytes)
	}
	return int32(binary.LittleEndian.Uint32(bytes)), nil
}

func int64FromBytes(bytes []byte) (int64, error) {
	if len(bytes) < 8 {
		return 0, fmt.Errorf("[]bytes not large enough to decode int64FromBytes: %v", bytes)
	}
	return int64(binary.LittleEndian.Uint64(bytes)), nil
}

func float64FromBytes(bytes []byte) (float64, error) {
	if len(bytes) < 8 {
		return 0, fmt.Errorf("[]bytes not large enough to decode float64FromBytes: %v", bytes)
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(bytes)), nil
}

func stringToBytes(s string) []byte {
	return []byte(s)
}

func boolToByte(b bool) byte {
	if b {
		return byte(1)
	}
	return byte(0)
}

func typeToByte(kt keyType) byte {
	return byte(kt)
}

func int32ToBytes(i int) []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(i))
	return bytes
}

func int64ToBytes(i int64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(i))
	return bytes
}

func float64ToBytes(f float64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, math.Float64bits(f))
	return bytes
}
*/
