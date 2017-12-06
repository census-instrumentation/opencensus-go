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

// +build !go1.6

package stackdriver

import (
	"runtime"
)

func pbStackTrace(s *trace.SpanData) *tracepb.StackTrace {
	pcs := s.StackTrace
	if pcs == nil {
		return nil
	}
	sf := &tracepb.StackTrace_StackFrames{}
	sp.StackTrace = &tracepb.StackTrace{StackFrames: sf}
	frames := runtime.CallersFrames(pcs)
	dropped := 0
	for {
		frame, more := frames.Next()
		if len(sf.Frame) >= 128 {
			// TODO: drop from the middle
			dropped++
		} else {
			sf.Frame = append(sf.Frame, &tracepb.StackTrace_StackFrame{
				FunctionName: trunc(frame.Function, 1024),
				FileName:     trunc(frame.File, 256),
				LineNumber:   int64(frame.Line),
			})
		}
		if !more {
			break
		}
	}
	sf.DroppedFramesCount = clip32(dropped)
	return &tracepb.StackTrace{StackFrames: sf}
}
