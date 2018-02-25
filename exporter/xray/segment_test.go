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

package xray

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"testing"
	"time"

	"go.opencensus.io/trace"
)

func BenchmarkSerializeSegment(t *testing.B) {
	var (
		w       = bytes.NewBuffer(make([]byte, 0, 2048))
		encoder = json.NewEncoder(w)
		s       = segment{
			Name:      "example.com",
			ID:        "70de5b6f19ff9a0a",
			TraceID:   "1-581cf771-a006649127e371903a2de979",
			StartTime: 1.478293361271E9,
			EndTime:   1.478293361449E9,
		}
	)

	for i := 0; i < t.N; i++ {
		w.Reset()
		if err := encoder.Encode(s); err != nil {
			t.FailNow()
		}
	}
}

func TestMakeID(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var (
			spanID   = trace.SpanID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8}
			expected = "0102030405060708"
			id       = MakeAmazonSpanID(spanID)
		)

		if id != expected {
			t.Errorf("got %v; want %v", id, expected)
		}
	})

	t.Run("zero", func(t *testing.T) {
		var (
			spanID   = trace.SpanID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
			expected = ""
			id       = MakeAmazonSpanID(spanID)
		)

		if id != expected {
			t.Errorf("got %v; want %v", id, expected)
		}
	})
}

func TestMakeTraceID(t *testing.T) {
	t.Run("epoch out of range", func(t *testing.T) {
		var (
			traceID       = trace.TraceID{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10}
			amazonTraceID = MakeAmazonTraceID(traceID)
		)

		parsedID, err := ParseAmazonTraceID(amazonTraceID)
		if err != nil {
			t.Fatalf("got %v; want nil", err)
		}

		if !bytes.Equal(traceID[4:16], parsedID[4:16]) {
			t.Error("expected identifier to be copied successfully")
		}
		if bytes.Equal(traceID[0:4], parsedID[0:4]) {
			t.Error("expected epoch to have been replaced, but was unchanged")
		}

		var (
			epoch = int64(binary.BigEndian.Uint32(parsedID[0:4]))
			now   = time.Now().Unix()
		)
		if delta := float64(now - epoch); math.Abs(delta) > float64(time.Second) {
			t.Error("expected epoch to be current time")
		}
	})
}

func TestParseAmazonTraceID(t *testing.T) {
	var (
		input    = "1-5759e988-05060708090a0b0c0d0e0f10"
		expected = trace.TraceID{0x57, 0x59, 0xe9, 0x88, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf, 0x10}
	)

	traceID, err := ParseAmazonTraceID(input)
	if err != nil {
		t.Fatalf("expected nil; got %v", err)
	}

	if expected != traceID {
		t.Fatalf("extracted traceID does not match expected")
	}

	const want = 1465510280
	if v := binary.BigEndian.Uint32(traceID[0:4]); v != want {
		t.Fatalf("got %v; want %v", v, want)
	}
}

func TestParseAmazonSpanID(t *testing.T) {
	var (
		input    = "53995c3f42cd8ad8"
		expected = trace.SpanID{0x53, 0x99, 0x5c, 0x3f, 0x42, 0xcd, 0x8a, 0xd8}
	)

	spanID, err := ParseAmazonSpanID(input)
	if err != nil {
		t.Fatalf("expected true; got false")
	}

	if expected != spanID {
		t.Fatalf("extracted traceID does not match expected")
	}
}
