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

import "go.opencensus.io/trace"

// FilteringExporter is a trace exporter that filters out some spans.
type FilteringExporter struct {
	Exporter trace.Exporter
	Matchers []*Matcher
}

func (f *FilteringExporter) ExportSpan(*trace.SpanData) {}

// Matcher allows to set rules what spans
// needs to be exported.
type Matcher struct {
	matcher func(*trace.SpanData) (ok bool)
}

func BySpanNamePrefixMatcher(prefix string) *Matcher {
	panic("not implemented")
}

func BySpanKindMatcher(kind int) *Matcher {
	panic("not implemented")
}
