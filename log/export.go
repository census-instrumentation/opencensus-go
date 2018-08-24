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

package log

import (
	"time"
)

// defaultLogger defines the global logger
var defaultLogger = &Logger{}

// Exporter is a type for functions that receive log data.
//
// The ExportLog method should be safe for concurrent use and should return
// quickly; if an Exporter takes a significant amount of time to process a
// Data, that work should be done on another goroutine.
type Exporter interface {
	ExportLog(d Data)
}

// Data contains a single log record
type Data struct {
	TraceID   string            // TraceID associated with current trace.Span (if present)
	SpanID    string            // SpanID associated with current trace.Span (if present)
	LogLevel  Level             // LogLevel is the level of the message; either InfoLevel or DebugLevel
	Timestamp time.Time         // Timestamp when the log record was received
	Message   string            // Message recorded
	Tags      map[string]string // Tags contains the optional list of tags found in context
	Fields    []Field           // Fields contains the log fields merged with the global fields (from ApplyConfig)
}

// RegisterExporter adds to the list of Exporters that will receive log data.
//
// Binaries can register exporters, libraries shouldn't register exporters.
func RegisterExporter(e Exporter) {
	defaultLogger.RegisterExporter(e)
}

// UnregisterExporter removes from the list of Exporters the Exporter that was
// registered with the given name.
func UnregisterExporter(e Exporter) {
	defaultLogger.UnregisterExporter(e)
}
