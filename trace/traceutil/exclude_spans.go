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

package traceutil

import (
	"strings"

	"go.opencensus.io/trace"
)

// ExcludeSpansExporter is an exporter that allows excluding specified spans by
// name and span kind.
type ExcludeSpansExporter struct {
	e       trace.Exporter
	exclude []SpanMatcher
}

// NewExcludeSpansExporter creates a new exporter that drops all spans that match
// any of the given SpanMatchers.
func NewExcludeSpansExporter(exporter trace.Exporter, excludeMatching []SpanMatcher) *ExcludeSpansExporter {
	return &ExcludeSpansExporter{exporter, excludeMatching}
}

var _ ExportFlusher = (*ExcludeSpansExporter)(nil)

// ExportSpan exports the given span data provided that it doesn't match any of
// the excluded matchers.
func (e *ExcludeSpansExporter) ExportSpan(spanData *trace.SpanData) {
	exclude := true
	for _, matcher := range e.exclude {
		if !matcher.matches(spanData) {
			exclude = false
			break
		}
	}
	if !exclude {
		e.e.ExportSpan(spanData)
	}
}

// Flush will flush the underlying exporter if this is supported.
func (e *ExcludeSpansExporter) Flush() {
	if fl, ok := e.e.(ExportFlusher); ok {
		fl.Flush()
	}
}

// SpanMatcher selects spans by name and kind. Both name and kind must match in
// order for the SpanMatcher to match.
type SpanMatcher struct {
	NamePrefix string // NamePrefix is a prefix of all matching span names.
	SpanKind   int    // SpanKind is the span kind of all matching spans.
}

func (sm SpanMatcher) matches(spanData *trace.SpanData) bool {
	return spanData.SpanKind == sm.SpanKind && strings.HasPrefix(spanData.Name, sm.NamePrefix)
}
