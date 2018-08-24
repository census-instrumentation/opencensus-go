// Copyright 2018, OpenCensus Authors
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

package logjson

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.opencensus.io/log"
)

var (
	hex       = "0123456789abcdef"
	jsonFalse = []byte(`false`)
	jsonNull  = []byte(`null`)
	jsonTrue  = []byte(`true`)
)

func Encode(buf *bytes.Buffer, fields []log.Field) error {
	buf.WriteByte('{')
	for index, field := range fields {
		if field.Type == log.NoOpType {
			continue
		}

		if index > 0 {
			buf.WriteByte(',')
		}

		encodeString(buf, field.Key)
		buf.WriteByte(':')

		if err := encodeField(buf, field); err != nil {
			return err
		}
	}
	buf.WriteByte('}')
	return nil
}

func encodeField(buf *bytes.Buffer, field log.Field) error {
	switch field.Type {
	case log.BoolType:
		return encodeBool(buf, field.Int == 1)
	case log.DurationType:
		encodeInt64(buf, field.Int/int64(time.Millisecond))
		return nil
	case log.ErrorType:
		encodeString(buf, field.Interface.(error).Error())
		return nil
	case log.Float32Type:
		return encodeFloat64(buf, field.Float, 32)
	case log.Float64Type:
		return encodeFloat64(buf, field.Float, 64)
	case log.IntType:
		return encodeInt64(buf, field.Int)
	case log.Int8Type:
		return encodeInt64(buf, field.Int)
	case log.Int16Type:
		return encodeInt64(buf, field.Int)
	case log.Int32Type:
		return encodeInt64(buf, field.Int)
	case log.Int64Type:
		return encodeInt64(buf, field.Int)
	case log.StringType:
		encodeString(buf, field.String)
		return nil
	case log.StringsType:
		if v, ok := field.Interface.([]string); ok && len(v) == 0 {
			buf.Write(jsonNull)
			return nil
		} else {
			encodeString(buf, strings.Join(field.Interface.([]string), ","))
			return nil
		}
	case log.StringerType:
		if field.Interface == nil {
			buf.Write(jsonNull)
			return nil
		} else {
			encodeString(buf, field.Interface.(fmt.Stringer).String())
			return nil
		}
	case log.TimeType:
		if t := field.Interface.(time.Time); t.IsZero() {
			buf.Write(jsonNull)
			return nil
		} else {
			encodeString(buf, t.In(time.UTC).Format(time.RFC3339))
			return nil
		}
	case log.UintType:
		return encodeInt64(buf, field.Int)
	case log.Uint8Type:
		return encodeInt64(buf, field.Int)
	case log.Uint16Type:
		return encodeInt64(buf, field.Int)
	case log.Uint32Type:
		return encodeInt64(buf, field.Int)
	case log.Uint64Type:
		return encodeInt64(buf, field.Int)
	default:
		panic(fmt.Errorf("unable to encode unexpected field type, %#v", field))
	}
}

func encodeBool(buf *bytes.Buffer, v bool) error {
	if v {
		buf.Write(jsonTrue)
	} else {
		buf.Write(jsonFalse)
	}
	return nil
}

func encodeFloat64(buf *bytes.Buffer, f float64, bits int) error {
	b := request()
	defer release(b)

	if math.IsInf(f, 0) || math.IsNaN(f) {
		return fmt.Errorf("unsupported value, %v", f)
	}

	// Convert as if by ES6 number to string conversion.
	// This matches most other JSON generators.
	// See golang.org/issue/6384 and golang.org/issue/14135.
	// Like fmt %g, but the exponent cutoffs are different
	// and exponents themselves are not padded to two digits.
	abs := math.Abs(f)
	format := byte('f')
	// Note: Must use float32 comparisons for underlying float32 value to get precise cutoffs right.
	if abs != 0 {
		if bits == 64 && (abs < 1e-6 || abs >= 1e21) || bits == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			format = 'e'
		}
	}
	b = strconv.AppendFloat(b, f, format, -1, int(bits))
	if format == 'e' {
		// clean up e-09 to e-9
		n := len(b)
		if n >= 4 && b[n-4] == 'e' && b[n-3] == '-' && b[n-2] == '0' {
			b[n-2] = b[n-1]
			b = b[:n-1]
		}
	}

	buf.Write(b)
	return nil
}

func encodeInt64(buf *bytes.Buffer, v int64) error {
	b := request()
	defer release(b)
	b = strconv.AppendInt(b, v, 10)

	_, err := buf.Write(b)
	return err
}

// encodeString was lifted from the go standard library encoding/json/encode.go
func encodeString(buf *bytes.Buffer, v string) {
	buf.WriteByte('"')
	start := 0
	for i := 0; i < len(v); {
		if b := v[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}
			if start < i {
				buf.WriteString(v[start:i])
			}
			switch b {
			case '\\', '"':
				buf.WriteByte('\\')
				buf.WriteByte(b)
			case '\n':
				buf.WriteByte('\\')
				buf.WriteByte('n')
			case '\r':
				buf.WriteByte('\\')
				buf.WriteByte('r')
			case '\t':
				buf.WriteByte('\\')
				buf.WriteByte('t')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				buf.WriteString(`\u00`)
				buf.WriteByte(hex[b>>4])
				buf.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(v[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				buf.WriteString(v[start:i])
			}
			buf.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				buf.WriteString(v[start:i])
			}
			buf.WriteString(`\u202`)
			buf.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(v) {
		buf.WriteString(v[start:])
	}
	buf.WriteByte('"')
}
