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
	"context"
	"sync"
	"time"

	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

// Logger defines the standard opencensus logger.
//
// To simplify testing, it can be instantiated directly.  In production cases,
// there should be no need to directly access Logger.
//
// Logger can be safely instantiated via logger := &Logger{}
type Logger struct {
	mutex     sync.Mutex
	config    Config
	exporters map[Exporter]struct{}
}

// ApplyConfig applies configuration to logger
func (l *Logger) ApplyConfig(cfg Config) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.config = cfg
}

// RegisterExporter adds to the list of Exporters that will receive log data.
//
// Binaries can register exporters, libraries shouldn't register exporters.
func (l *Logger) RegisterExporter(e Exporter) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// treat l.exporters as immutable
	exporters := map[Exporter]struct{}{}
	for exporter := range l.exporters {
		exporters[exporter] = struct{}{}
	}

	exporters[e] = struct{}{}
	l.exporters = exporters
}

// UnregisterExporter removes from the list of Exporters the Exporter that was
// registered with the given name.
func (l *Logger) UnregisterExporter(e Exporter) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// treat l.exporters as immutable
	exporters := map[Exporter]struct{}{}
	for exporter := range l.exporters {
		exporters[exporter] = struct{}{}
	}

	delete(exporters, e)
	l.exporters = exporters
}

// now returns the current time
func (l *Logger) now(fn func() time.Time) time.Time {
	if fn != nil {
		return fn()
	}

	return time.Now()
}

// tags returns a map of tags found in the current context
func (l *Logger) tags(ctx context.Context, tagKeys []tag.Key) map[string]string {
	var tags map[string]string

	tagMap := tag.FromContext(ctx)
	for _, key := range tagKeys {
		if v, ok := tagMap.Value(key); ok {
			if tags == nil {
				tags = map[string]string{}
			}
			tags[key.Name()] = v
		}
	}

	return tags
}

// log defines an internal helper method to log at any level
func (l *Logger) log(ctx context.Context, level Level, message string, fields ...Field) {
	l.mutex.Lock()
	config := l.config
	exporters := l.exporters
	l.mutex.Unlock()

	if level < config.LogLevel {
		return
	}

	var (
		now          = l.now(config.TimeFunc)
		tags         = l.tags(ctx, config.Tags)
		mergedFields = mergeFields(fields, config.Fields)
	)

	data := Data{
		LogLevel:  level,
		Timestamp: now,
		Message:   message,
		Tags:      tags,
		Fields:    mergedFields,
	}

	if span := trace.FromContext(ctx); span != nil {
		data.TraceID = span.SpanContext().TraceID.String()
		data.SpanID = span.SpanContext().SpanID.String()
	}

	for exporter := range exporters {
		exporter.ExportLog(data)
	}
}

func (l *Logger) Debug(ctx context.Context, message string, fields ...Field) {
	l.log(ctx, DebugLevel, message, fields...)
}

func (l *Logger) Info(ctx context.Context, message string, fields ...Field) {
	l.log(ctx, InfoLevel, message, fields...)
}
