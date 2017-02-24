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

import "bytes"

// Tag is the tuple (key, value) interface for all tag types.
type Tag interface {
	WriteValueToBuffer(dst bytes.Buffer)
	WriteKeyValueToBuffer(dst bytes.Buffer)
	Key() Key
}

// tagString is the tuple (key, value) implementation for tags of value type
// string.
type tagString struct {
	*keyString
	v string
}

func (ts *tagString) WriteValueToBuffer(dst bytes.Buffer) {
	dst.Write(int32ToBytes(len(ts.v)))
	dst.Write([]byte(ts.v))
}

func (ts *tagString) WriteKeyValueToBuffer(dst bytes.Buffer) {
	// TODO(acetechnologist): implement
}

func (ts *tagString) Key() Key {
	return ts.keyString
}

func (ts *tagString) Value() string {
	return ts.v
}

// tagBool is the tuple (key, value) implementation for tags of value type
// bool.
type tagBool struct {
	*keyBool
	v bool
}

func (tb *tagBool) WriteValueToBuffer(dst bytes.Buffer) {
	dst.Write(int32ToBytes(1))
	dst.WriteByte(boolToByte(tb.v))
}

func (tb *tagBool) WriteKeyValueToBuffer(dst bytes.Buffer) {
	// TODO(acetechnologist): implement
}

func (tb *tagBool) Key() Key {
	return tb.keyBool
}

func (tb *tagBool) Value() bool {
	return tb.v
}

// tagInt64 is the tuple (key, value) implementation for tags of value type
// int64.
type tagInt64 struct {
	*keyInt64
	v int64
}

func (ti *tagInt64) WriteValueToBuffer(dst bytes.Buffer) {
	dst.Write(int32ToBytes(8))
	dst.Write(int64ToBytes(ti.v))
}

func (ti *tagInt64) WriteKeyValueToBuffer(dst bytes.Buffer) {
	// TODO(acetechnologist): implement
}

func (ti *tagInt64) Key() Key {
	return ti.keyInt64
}

func (ti *tagInt64) Value() int64 {
	return ti.v
}
