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

// Package tracestate contains HTTP propagator for Tracestate as specified
// in Tracecontext standard.
// See https://github.com/w3c/distributed-tracing for more information.
package tracestate // import "go.opencensus.io/trace/propagation/tracestate"

import (
	"net/http"
	"strings"

	ts "go.opencensus.io/trace/tracestate"
	"regexp"
)

const (
	MaxTracestateLen  = 512
	tracestateHeader  = "tracestate"
	trimOWSRegexFmt   = `^[\x09\x20]*(.*[^\x20\x09])[\x09\x20]*$`
)

var trimOWSRegExp = regexp.MustCompile(trimOWSRegexFmt)

// TODO(rghetia): return an empty Tracestate when parsing tracestate header encounters an error.
// Revisit to return additional boolean value to indicate parsing error when following issues
// are resolved.
// https://github.com/w3c/distributed-tracing/issues/172
// https://github.com/w3c/distributed-tracing/issues/175
func FromRequest(req *http.Request) *ts.Tracestate {
	h := req.Header.Get(tracestateHeader)
	if h == "" {
		return nil
	}

	var entries []ts.Entry
	pairs := strings.Split(h, ",")
	headerLenWithoutTrailingSpaces := len(pairs) - 1 // Number of commas
	for _, pair := range pairs {
		matches := trimOWSRegExp.FindStringSubmatch(pair)
		if matches == nil {
			return nil
		}
		pair = matches[1]
		headerLenWithoutTrailingSpaces += len(pair)
		if headerLenWithoutTrailingSpaces > MaxTracestateLen {
			return nil
		}
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil
		}
		entries = append(entries, ts.Entry{Key: kv[0], Value: kv[1]})
	}
	ts, err := ts.New(nil, entries...)
	if err != nil {
		return nil
	}

	return ts
}

func ToRequest(ts *ts.Tracestate, req *http.Request) {
	var pairs = make([]string, 0, len(ts.Entries()))
	if ts != nil {
		for _, entry := range ts.Entries() {
			pairs = append(pairs, strings.Join([]string{entry.Key, entry.Value}, "="))
		}
		h := strings.Join(pairs, ",")

		if h != "" && len(h) <= MaxTracestateLen {
			req.Header.Set(tracestateHeader, h)
		}
	}
}