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

import "unsafe"

var sizeOfUint16 = (int)(unsafe.Sizeof(uint16(0)))

type valuesBytes struct {
	buf        []byte
	wIdx, rIdx int
}

func (vb *valuesBytes) growIfRequired(expected int) {
	if len(vb.buf)-vb.wIdx < expected {
		tmp := make([]byte, 2*(len(vb.buf)+1)+expected)
		copy(tmp, vb.buf)
		vb.buf = tmp
	}
}

func (vb *valuesBytes) writeValue(v []byte) {
	length := len(v)
	vb.growIfRequired(sizeOfUint16 + length)

	// writing length of v
	bytes := *(*[2]byte)(unsafe.Pointer(&length))
	vb.buf[vb.wIdx] = bytes[0]
	vb.wIdx++
	vb.buf[vb.wIdx] = bytes[1]
	vb.wIdx++

	if length == 0 {
		// No value was encoded for this key
		return
	}

	// writing v
	copy(vb.buf[vb.wIdx:], v)
	vb.wIdx += length
}

// readValue is the helper method to read the values when decoding valuesBytes to a map[Key][]byte.
// It is meant to be used by toMap(...) only.
func (vb *valuesBytes) readValue() []byte {
	// read length of v
	length := (int)(*(*uint16)(unsafe.Pointer(&vb.buf[vb.rIdx])))
	vb.rIdx += sizeOfUint16
	if length == 0 {
		// No value was encoded for this key
		return nil
	}

	// read value of v
	v := make([]byte, length)
	endIdx := vb.rIdx + length
	copy(v, vb.buf[vb.rIdx:endIdx])
	vb.rIdx = endIdx
	return v
}

func (vb *valuesBytes) toMap(ks []Key) map[Key][]byte {
	m := make(map[Key][]byte, len(ks))
	for _, k := range ks {
		v := vb.readValue()
		if v != nil {
			m[k] = v
		}
	}
	vb.rIdx = 0
	return m
}

func (vb *valuesBytes) Bytes() []byte {
	return vb.buf[:vb.wIdx]
}

func toValuesBytes(ts *TagSet, ks []Key) *valuesBytes {
	vb := &valuesBytes{
		buf: make([]byte, len(ks)),
	}
	for _, k := range ks {
		v := ts.m[k]
		vb.writeValue(v)
	}
	return vb
}