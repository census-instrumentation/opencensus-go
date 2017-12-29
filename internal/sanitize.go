// Copyright 2017, OpenCensus Authors
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

package internal

import (
	"fmt"
	"strings"
	"sync"
	"unicode"
)

const labelKeySizeLimit = 100

// Sanitize returns a string that is trunacated to 100 characters if it's too
// long, and replaces non-alphanumeric characters to underscores.
func Sanitize(s string) string {
	if len(s) == 0 {
		return s
	}
	if len(s) > labelKeySizeLimit {
		s = s[:labelKeySizeLimit]
	}
	s = strings.Map(sanitizeRune, s)
	if unicode.IsDigit(rune(s[0])) {
		s = "key_" + s
	}
	if s[0] == '_' {
		s = "key" + s
	}
	return s
}

var (
	smu       sync.Mutex
	sanitized = make(map[string]string)
)

// SanitizeNoClash returns an error if in's sanitization clashes
// with that of a previously sanitized string yet the two strings are
// lexicographically different.
//
// For example:
//   Sanitize("_012_foo")     == "key_012_foo"
//   Sanitize("key_012_foo")  == "key_012_foo"
//   Sanitize("012?foo")      == "key_012_foo"
// all produce the same sanitization "key_012_foo"
// yet none of them are the same.
func SanitizeNoClash(in string) (string, error) {
	sn := Sanitize(in)
	smu.Lock()
	defer smu.Unlock()
	if prev, ok := sanitized[sn]; ok && prev != in {
		return "", fmt.Errorf("sanitization %q of %q clashes with already sanitized %q", sn, in, prev)
	}
	sanitized[sn] = in
	return sn, nil
}

// converts anything that is not a letter or digit to an underscore
func sanitizeRune(r rune) rune {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return r
	}
	// Everything else turns into an underscore
	return '_'
}
