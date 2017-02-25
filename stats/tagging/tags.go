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

// Tag is the tuple (key, value) interface for all tag types.
type Tag interface {
	WriteValueToBuffer(dst *bytes.Buffer)
	WriteKeyValueToBuffer(dst *bytes.Buffer)
	Key() Key
}

// tagString is the tuple (key, value) implementation for tags of value type
// string.
type tagString struct {
	*keyString
	v string
}

func (ts *tagString) WriteValueToBuffer(dst *bytes.Buffer) {
	if len(ts.v) == 0 {
		// string length is zero. Will not be encoded.
		dst.Write(int32ToBytes(0))
	}
	dst.Write(int32ToBytes(len(ts.v)))
	dst.Write([]byte(ts.v))
}

func (ts *tagString) WriteKeyValueToBuffer(dst *bytes.Buffer) {
	// TODO(acetechnologist): implement
}

func (ts *tagString) Key() Key {
	return ts.keyString
}

func (ts *tagString) Value() string {
	return ts.v
}

func (ts *tagString) String() string {
	return fmt.Sprintf("{%s, %s}", ts.name, ts.v)
}

// tagBool is the tuple (key, value) implementation for tags of value type
// bool.
type tagBool struct {
	*keyBool
	v bool
}

func (tb *tagBool) WriteValueToBuffer(dst *bytes.Buffer) {
	dst.Write(int32ToBytes(1))
	dst.WriteByte(boolToByte(tb.v))
}

func (tb *tagBool) WriteKeyValueToBuffer(dst *bytes.Buffer) {
	// TODO(acetechnologist): implement
}

func (tb *tagBool) Key() Key {
	return tb.keyBool
}

func (tb *tagBool) Value() bool {
	return tb.v
}

func (tb *tagBool) String() string {
	return fmt.Sprintf("{%s, %v}", tb.name, tb.v)
}

// tagInt64 is the tuple (key, value) implementation for tags of value type
// int64.
type tagInt64 struct {
	*keyInt64
	v int64
}

func (ti *tagInt64) WriteValueToBuffer(dst *bytes.Buffer) {
	dst.Write(int32ToBytes(8))
	dst.Write(int64ToBytes(ti.v))
}

func (ti *tagInt64) WriteKeyValueToBuffer(dst *bytes.Buffer) {
	// TODO(acetechnologist): implement
}

func (ti *tagInt64) Key() Key {
	return ti.keyInt64
}

func (ti *tagInt64) Value() int64 {
	return ti.v
}

func (ti *tagInt64) String() string {
	return fmt.Sprintf("{%s, %v}", ti.name, ti.v)
}

/// ON THE WIRE:
// // StatsContext describes the encoding of stats context information (tags)
// // for passing across RPC's.
// message StatsContext {
//   // Tags are encoded as a single byte sequence. The format is:
//   // [tag_type key_len key_bytes value_len value_bytes]*
//   //
//   // Where:
//   //  * tag_type is one byte, and is used to describe the format of value_bytes.
//   //    In particular, the low 2 bits of this byte are used as follows:
//   //    00 (value 0): string (UTF-8) encoding
//   //    01 (value 1): integer (varint int64 encoding). See
//   //      https://developers.google.com/protocol-buffers/docs/encoding#varints
//   //      for documentation on the varint format.
//   //    10 (value 2): boolean format. In this case value_len should equal 1, and
//   //       the value_bytes will be a single byte containing either 0 (false) or
//   //       1 (true).
//   //    11 (value 3): byte sequence. Arbitrary uninterpreted bytes.
//   //  * The key_len and value_len fields are represented using a varint, with a
//   //    maximum value of 16383 bytes (this value is guaranteed to fit in at most
//   //    2 bytes). Zero length keys or values are not allowed.
//   //  * The value in key_bytes is a US-ASCII format string.
//   bytes tags = 1;
// }
