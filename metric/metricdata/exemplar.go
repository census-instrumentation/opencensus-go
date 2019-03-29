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

package metricdata

import (
	"time"

	"go.opencensus.io/trace"
)

// Exemplars keys.
const (
	AttachmentKeySpanContext = "SpanContext"
)

// Exemplar is an example data point associated with each bucket of a
// distribution type aggregation.
//
// Their purpose is to provide an example of the kind of thing
// (request, RPC, trace span, etc.) that resulted in that measurement.
type Exemplar struct {
	Value       float64                // the value that was recorded
	Timestamp   time.Time              // the time the value was recorded
	Attachments map[string]interface{} // attachments (if any)
}

// Attachment is a key-value pair associated with a recorded example data point.
type Attachment struct {
	Key   string
	Value interface{}
}

// SpanContextAttachment returns a span context valued attachment.
func SpanContextAttachment(key string, value trace.SpanContext) Attachment {
	return Attachment{Key: key, Value: value}
}

// StringAttachment returns a string attachment.
func StringAttachment(key string, value string) Attachment {
	return Attachment{Key: key, Value: value}
}
