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

// ExportFlusher extends Exporter to support synchronous flushing of any
// internal buffered spans. As with most aspects of exporting, it is
// best-effort only.
type ExportFlusher interface {
	trace.Exporter
	Flush() // Flush synchronously sends any internally buffered spans.
}
