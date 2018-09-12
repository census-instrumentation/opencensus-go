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
package tracestate

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	ts "go.opencensus.io/trace/tracestate"
)

const (
	maxTracestateLen = 512
	tsHeader         = "tracestate"
	trimOWSRegexFmt  = `^[\x09\x20]*(.*[^\x20\x09])[\x09\x20]*$`
)

var trimOWSRegExp = regexp.MustCompile(trimOWSRegexFmt)
var errInvalidTs = errors.New("invalid tracestate header")

// FromRequest extracts Tracestate from http request as per the spec at
// https://github.com/w3c/distributed-tracing
// If tracestate header is not present then nil, nil is returned.
// If extraction fails then nil, error is returned.
func FromRequest(req *http.Request) (*ts.Tracestate, error) {
	h := req.Header.Get(tsHeader)
	if h == "" {
		return nil, nil
	}

	// TODO(rghetia): Revisit to return appropriate error when following issues are resolved.
	// https://github.com/w3c/distributed-tracing/issues/172
	// https://github.com/w3c/distributed-tracing/issues/175
	var entries []ts.Entry
	pairs := strings.Split(h, ",")
	hdrLenWithoutOWS := len(pairs) - 1 // Number of commas
	for _, pair := range pairs {
		matches := trimOWSRegExp.FindStringSubmatch(pair)
		if matches == nil {
			return nil, errInvalidTs
		}
		pair := matches[1]
		hdrLenWithoutOWS += len(pair)
		if hdrLenWithoutOWS > maxTracestateLen {
			return nil, errInvalidTs
		}
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errInvalidTs
		}
		entries = append(entries, ts.Entry{Key: kv[0], Value: kv[1]})
	}
	return ts.New(nil, entries...)
}

// ToRequest injects tracestate header as per the spec at
// https://github.com/w3c/distributed-tracing
// if header len exceeds maxTracestateLen then the header is not injected.
func ToRequest(ts *ts.Tracestate, req *http.Request) {
	var pairs = make([]string, 0, len(ts.Entries()))
	if ts != nil {
		for _, entry := range ts.Entries() {
			pairs = append(pairs, strings.Join([]string{entry.Key, entry.Value}, "="))
		}
		h := strings.Join(pairs, ",")

		if h != "" && len(h) <= maxTracestateLen {
			req.Header.Set(tsHeader, h)
		}
	}
}
