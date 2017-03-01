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

// Key is the interface for all key types.
type Key interface {
	Name() string
	Type() keyType
	setFromBytes([]byte, int32) (int32, error)

	writeKeyToBuffer(dst *bytes.Buffer)
}

// KeyType defines the types of keys allowed.
type keyType byte

const (
	keyTypeStringUTF8 keyType = iota
	keyTypeInt64
	keyTypeBool
	keyTypeBytes
)
