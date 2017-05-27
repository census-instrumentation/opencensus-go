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

import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

type keyType byte

const (
	keyTypeString keyType = iota
	keyTypeInt64
	keyTypeTrue
	keyTypeFalse
)

const (
	tagsVersionID = byte(0)

	srvStatsVersionFieldID   = byte(0)
	srvStatsLatencyNsFieldID = byte(0)

	traceVersionID                     = byte(0)
	traceFieldID                       = byte(0)
	traceSpanFieldID                   = byte(1)
	traceOptionsFieldID                = byte(2)
	traceMaskFieldID                   = byte(0xFC)
	traceInvSamplingProbabilityFieldID = byte(0xFD)
	traceParenSpanFieldID              = byte(0xFE)
)

type GRPCCodec struct {
	buf        []byte
	wIdx, rIdx int
}

func (gc *GRPCCodec) EncodeTagSet(ts *TagSet) []byte {
	gc.buf = make([]byte, len(ts.m))
	gc.wIdx, gc.rIdx = 0, 0

	gc.writeByte(byte(tagsVersionID))
	for k, v := range ts.m {
		if k.Name() == "method" || k.Name() == "caller" {
			continue
		}

		switch k.(type) {
		case *KeyString:
			gc.writeByte(byte(keyTypeString))
			gc.writeStringWithVarintLen(k.Name())
			gc.writeBytesWithVarintLen(v)
			break
		case *KeyInt64:
			gc.writeByte(byte(keyTypeInt64))
			gc.writeStringWithVarintLen(k.Name())
			gc.writeBytes(v)
			break
		case *KeyBool:
			if v[0] == 1 {
				gc.writeTagTrue(k.Name())
			} else {
				gc.writeTagFalse(k.Name())
			}
			break
		default:
			continue
		}
	}

	return gc.buf[:gc.wIdx]
}

func (gc *GRPCCodec) DecodeTagSet(bytes []byte) (*TagSet, error) {
	gc.buf = bytes
	gc.wIdx, gc.rIdx = 0, 0
	ts := newTagSet(0)

	if len(gc.buf) == 0 {
		return ts, nil
	}

	version, err := gc.readByte()
	if err != nil {
		return nil, err
	}
	if version > tagsVersionID {
		return nil, fmt.Errorf("GRPCCodec.DecodeTagSet() doesn't support version %v. Supports only up to: %v", version, tagsVersionID)
	}

	var k Key
	var v []byte

	for !gc.readEnded() {
		typByte, err := gc.readByte()
		if err != nil {
			continue
		}
		typ := keyType(typByte)
		kName, err := gc.readStringWithVarintLen()
		if err != nil {
			continue
		}
		switch typ {
		case keyTypeString:
			v, err = gc.readBytesWithVarintLen()
			if err != nil {
				continue
			}
			k, err = DefaultKeyManager().CreateKeyString(kName)
			if err != nil {
				continue
			}
			break
		case keyTypeInt64:
			v, err = gc.readBytes(8)
			if err != nil {
				continue
			}
			k, err = DefaultKeyManager().CreateKeyInt64(kName)
			if err != nil {
				continue
			}
			break
		case keyTypeFalse:
			v = []byte{0}
			k, err = DefaultKeyManager().CreateKeyBool(kName)
			if err != nil {
				continue
			}
			break
		case keyTypeTrue:
			v = []byte{1}
			k, err = DefaultKeyManager().CreateKeyBool(kName)
			if err != nil {
				continue
			}
			break
		default:
			return nil, fmt.Errorf("toStubbyTagsFormat failed. Key type invalid %v", typ)
		}
		ts.upsertBytes(k, v)
	}
	return ts, nil
}

func (gc *GRPCCodec) growIfRequired(expected int) {
	if len(gc.buf)-gc.wIdx < expected {
		tmp := make([]byte, 2*(len(gc.buf)+1)+expected)
		copy(tmp, gc.buf)
		gc.buf = tmp
	}
}

func (gc *GRPCCodec) writeTagString(k, v string) {
	gc.writeByte(byte(keyTypeString))
	gc.writeStringWithVarintLen(k)
	gc.writeStringWithVarintLen(v)
}

func (gc *GRPCCodec) writeTagUint64(k string, i uint64) {
	gc.writeByte(byte(keyTypeInt64))
	gc.writeStringWithVarintLen(k)
	gc.writeUint64(i)
}

func (gc *GRPCCodec) writeTagTrue(k string) {
	gc.writeByte(byte(keyTypeTrue))
	gc.writeStringWithVarintLen(k)
}

func (gc *GRPCCodec) writeTagFalse(k string) {
	gc.writeByte(byte(keyTypeFalse))
	gc.writeStringWithVarintLen(k)
}

func (gc *GRPCCodec) writeTraceID(low, high uint64) {
	gc.growIfRequired(17)
	gc.writeByte(traceFieldID)
	gc.writeUint64(low)
	gc.writeUint64(high)
}

func (gc *GRPCCodec) writeTraceSpanID(i uint64) {
	gc.growIfRequired(9)
	gc.writeByte(traceSpanFieldID)
	gc.writeUint64(i)
}

func (gc *GRPCCodec) writeTraceOptions(i byte) {
	gc.growIfRequired(2)
	gc.writeByte(traceOptionsFieldID)
	gc.writeByte(i)
}

func (gc *GRPCCodec) writeTraceMask(i uint32) {
	gc.growIfRequired(5)
	gc.writeByte(traceMaskFieldID)
	gc.writeUint32(i)
}

func (gc *GRPCCodec) writeTraceInverseSamplingProbability(f float64) {
	g := float32(f)
	i := *(*uint32)(unsafe.Pointer(&g))
	gc.growIfRequired(5)
	gc.writeByte(traceInvSamplingProbabilityFieldID)
	gc.writeUint32(i)
}

func (gc *GRPCCodec) writeTraceParenSpanID(i uint64) {
	gc.growIfRequired(9)
	gc.writeByte(traceParenSpanFieldID)
	gc.writeUint64(i)
}

func (gc *GRPCCodec) writeServerStatsLatencyNs(i uint64) {
	gc.growIfRequired(9)
	gc.writeByte(srvStatsLatencyNsFieldID)
	gc.writeUint64(i)
}

func (gc *GRPCCodec) writeBytesWithVarintLen(bytes []byte) {
	length := len(bytes)

	gc.growIfRequired(binary.MaxVarintLen64 + length)
	gc.wIdx += binary.PutUvarint(gc.buf[gc.wIdx:], uint64(length))
	copy(gc.buf[gc.wIdx:], bytes)
	gc.wIdx += length
}

func (gc *GRPCCodec) writeStringWithVarintLen(s string) {
	length := len(s)

	gc.growIfRequired(binary.MaxVarintLen64 + length)
	gc.wIdx += binary.PutUvarint(gc.buf[gc.wIdx:], uint64(length))
	copy(gc.buf[gc.wIdx:], s)
	gc.wIdx += length
}

func (gc *GRPCCodec) writeByte(v byte) {
	gc.growIfRequired(1)
	gc.buf[gc.wIdx] = v
	gc.wIdx++
}

func (gc *GRPCCodec) writeBytes(bytes []byte) {
	length := len(bytes)
	gc.growIfRequired(length)
	copy(gc.buf[gc.wIdx:], bytes)
	gc.wIdx += length
}

func (gc *GRPCCodec) writeUint32(i uint32) {
	gc.growIfRequired(4)
	binary.LittleEndian.PutUint32(gc.buf[gc.wIdx:], i)
	gc.wIdx += 4
}

func (gc *GRPCCodec) writeUint64(i uint64) {
	gc.growIfRequired(8)
	binary.LittleEndian.PutUint64(gc.buf[gc.wIdx:], i)
	gc.wIdx += 8
}

func (gc *GRPCCodec) readByte() (byte, error) {
	if len(gc.buf) < gc.rIdx+1 {
		return 0, fmt.Errorf("unexpected end while readByte '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	b := gc.buf[gc.rIdx]
	gc.rIdx++
	return b, nil
}

func (gc *GRPCCodec) readBytes(count int) ([]byte, error) {
	if len(gc.buf) < gc.rIdx+count {
		return nil, fmt.Errorf("unexpected end while readBytes '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	endIdx := gc.rIdx + count
	tmp := gc.buf[gc.rIdx:endIdx]
	gc.rIdx = endIdx
	return tmp, nil
}

func (gc *GRPCCodec) readUint32() (uint32, error) {
	if len(gc.buf) < gc.rIdx+4 {
		return 0, fmt.Errorf("unexpected end while readUint32 '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	i := binary.LittleEndian.Uint32(gc.buf[gc.rIdx:])
	gc.rIdx += 4
	return i, nil
}

func (gc *GRPCCodec) readUint64() (uint64, error) {
	if len(gc.buf) < gc.rIdx+8 {
		return 0, fmt.Errorf("unexpected end while readUint64 '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	i := binary.LittleEndian.Uint64(gc.buf[gc.rIdx:])
	gc.rIdx += 8
	return i, nil
}

func (gc *GRPCCodec) readBytesWithVarintLen() ([]byte, error) {
	if gc.readEnded() {
		return nil, fmt.Errorf("unexpected end while readBytesWithVarintLen '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}
	length, valueStart := binary.Uvarint(gc.buf[gc.rIdx:])
	if valueStart <= 0 {
		return nil, fmt.Errorf("unexpected end while readBytesWithVarintLen '%x' starting at idx '%v'", gc.buf, gc.rIdx)
	}

	valueStart += gc.rIdx
	valueEnd := valueStart + int(length)
	if valueEnd > len(gc.buf) || length < 0 {
		return nil, fmt.Errorf("malformed encoding: length:%v, upper%v, maxLength:%v", length, valueEnd, len(gc.buf))
	}

	gc.rIdx = valueEnd
	return gc.buf[valueStart:valueEnd], nil
}

func (gc *GRPCCodec) readStringWithVarintLen() (string, error) {
	bytes, err := gc.readBytesWithVarintLen()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (gc *GRPCCodec) readEnded() bool {
	return gc.rIdx >= len(gc.buf)
}
