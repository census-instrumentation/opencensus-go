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

// Package opentracer contains an OpenTracing implementation for OpenCensus.
package opentracer // import "go.opencensus.io/contrib/opentracer"

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/opentracing/opentracing-go/log"
)

type Logger interface {
	LogFields(fields ...log.Field)
	LogFieldsTime(t time.Time, fields ...log.Field)
}

var (
	Stdout Logger = writeLogger{w: os.Stdout}
)

type writeLogger struct {
	w io.Writer
}

func (s writeLogger) LogFields(fields ...log.Field) {
	s.LogFieldsTime(time.Now(), fields...)
}

func (s writeLogger) LogFieldsTime(t time.Time, fields ...log.Field) {
	var buffer = bytes.NewBuffer(nil)

	buffer.WriteString(t.Format(time.RFC822Z))
	buffer.WriteString(" ")

	for index, field := range fields {
		if index > 0 {
			buffer.WriteString(" ")
		}
		fmt.Fprintf(buffer, "%v=%v", field.Key(), field.Value())
	}
	fmt.Fprintln(s.w, buffer.String())
}
