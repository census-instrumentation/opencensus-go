// Copyright 2017, OpenCensus Authors
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

// +build !go1.8

package stackdriver

import (
	"runtime"

	"go.opencensus.io/trace"
	tracepb "google.golang.org/genproto/googleapis/devtools/cloudtrace/v2"
)

func pbStackTrace(s *trace.SpanData) *tracepb.StackTrace {
	pcs := s.StackTrace
	if pcs == nil {
		return nil
	}
	sf := &tracepb.StackTrace_StackFrames{}
	for _, pc := range pcs {
		// The idea is to expand and find the function-name and line numbers.
		// However, Go1.6 and below didn't have runtime.CallersFrame.
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		entryPC := fn.Entry()
		fileName, lineNumber := fn.FileLine(entryPC)
		sf.Frame = append(sf.Frame, &tracepb.StackTrace_StackFrame{
			FunctionName: trunc(fn.Name(), 1024),
			FileName:     trunc(fileName, 256),
			LineNumber:   int64(lineNumber),
		})
	}
	return &tracepb.StackTrace{StackFrames: sf}
}
