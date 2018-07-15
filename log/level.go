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
)

// Level defines the priority level of the log
type Level int

const (
	DebugLevel Level = iota - 1 // DebugLevel provides detailed logging; should be disabled for production
	InfoLevel                   // InfoLevel is the default priority level
)

// Debug message (requires DebugLevel or better)
func Debug(ctx context.Context, message string, fields ...Field) {
	defaultLogger.Debug(ctx, message, fields...)
}

// Info message.  General purpose messages to be logged.
func Info(ctx context.Context, message string, fields ...Field) {
	defaultLogger.Info(ctx, message, fields...)
}
