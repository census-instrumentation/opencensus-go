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

import "encoding/binary"

type buffer struct {
	bytes    []byte
	writeIdx int
}

func (b * buffer) Read()
func (b *buffer) writeMetadataTypeStringUTF8() {
	b.growIfRequired(1)
	b.bytes[b.writeIdx] = 0
	b.writeIdx++
}

func (b *buffer) writeMetadataTypeInt64() {
	b.growIfRequired(1)
	b.bytes[b.writeIdx] = 1
	b.writeIdx++
}

func (b *buffer) writeMetadataTypeBool() {
	b.growIfRequired(1)
	b.bytes[b.writeIdx] = 2
	b.writeIdx++
}

func (b *buffer) writeMetadataTypeBytes() {
	b.growIfRequired(1)
	b.bytes[b.writeIdx] = 3
	b.writeIdx++
}

// writeValueFalse writes the value when it is a true bool.
func (b *buffer) writeValueTrue() {
	b.growIfRequired(2)
	// 2 in next line is the varint encoding for 1. Equivalent to calling
	// binary.PutVarint(b.bytes[b.writeIdx:], int64(1))
	b.bytes[b.writeIdx] = 2
	b.writeIdx++
	b.bytes[b.writeIdx] = byte(1)
	b.writeIdx++
}

// writeValueFalse writes the value when it is a false bool.
func (b *buffer) writeValueFalse() {
	b.growIfRequired(2)

	// 2 in next line is the varint encoding for 1. Equivalent to calling
	// binary.PutVarint(b.bytes[b.writeIdx:], int64(1))
	b.bytes[b.writeIdx] = 2
	b.writeIdx++
	b.bytes[b.writeIdx] = 0
	b.writeIdx++
}

// writeBytes writes the value when it is a []byte
func (b *buffer) writeBytes(u []byte) {
	b.growIfRequired(binary.MaxVarintLen64 + len(u))
	bytesWritten := binary.PutVarint(b.bytes[b.writeIdx:], int64(len(u)))
	b.writeIdx += bytesWritten
	copy(b.bytes[b.writeIdx:], u)
	b.writeIdx += len(u)
}

func (b *buffer) writeByte(i byte) {
	b.growIfRequired(1)
	b.bytes[b.writeIdx] = i
	b.writeIdx++
}

func (b *buffer) writeZero() {
	b.growIfRequired(2)
	b.bytes[b.writeIdx] = uint16(0)
	b.writeIdx++
}

// writeStringUTF8 writes the key (always a string) or value (when it is a
// stringUTF8)
func (b *buffer) writeStringUTF8(s string) {
	b.growIfRequired(binary.MaxVarintLen64 + len(s))
	bytesWritten := binary.PutVarint(b.bytes[b.writeIdx:], int64(len(s)))
	b.writeIdx += bytesWritten
	copy(b.bytes[b.writeIdx:], s)
	b.writeIdx += len(s)
}

// writeValueInt64 writes the value when it is an int64. It encodes it as a
// varint64. It is preceded with 1 byte representing the length of encoded
// varint64. The 1 byte length is itself varint16 encoded.
func (b *buffer) writeValueInt64(i int64) {
	b.growIfRequired(binary.MaxVarintLen64 + 1)
	// we skip 1 byte because we need to write in it the length of the encoded
	// "i". This length is not known until after the encoding occurs.
	bytesWritten := binary.PutVarint(b.bytes[b.writeIdx+1:], i)
	//we write the length of the encoded "i" in hte skipped byte.
	binary.PutVarint(b.bytes[b.writeIdx:], int64(bytesWritten))
	b.writeIdx += bytesWritten + 1
}

func (b *buffer) growIfRequired(expected int) {
	for len(b.bytes)-b.writeIdx < expected {
		tmp := make([]byte, len(b.bytes)*2)
		copy(tmp, b.bytes)
		b.bytes = tmp
	}
}
