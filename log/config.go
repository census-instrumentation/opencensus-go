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

	"go.opencensus.io/tag"
)

// Config represents the global log configuration.
type Config struct {
	LogLevel Level            // LogLevel; defaults to InfoLevel
	TimeFunc func() time.Time // TimeFunc provides optional time generator
	Fields   []Field          // Fields that will be included in all log messages
	Tags     []tag.Key        // Tags that will be added to the log
}

// ApplyConfig applies changes to the global tracing configuration.
//
// Fields not provided in the given config are going to be preserved.
func ApplyConfig(cfg Config) {
	defaultLogger.ApplyConfig(cfg)
}
