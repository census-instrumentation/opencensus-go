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
//

package exporter

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	"go.opencensus.io/tag"
)

// A ViewData is a set of rows about usage of the single measure associated
// with the given view. Each row is specific to a unique set of tags.
type ViewData struct {
	Name         string
	Description  string
	Unit         string // unit of this view (a function of both the Aggregation and the Measure)
	MeasureFloat bool   // does the associated measure emit floating point values (as opposed to just integer values)
	TagKeys      []tag.Key
	Aggregation  Aggregation

	Start, End time.Time
	Rows       []*Row // Rows is set of data points accumulated for this view.
}

// Row is the collected value for a specific set of tag key-value pairs.
type Row struct {
	Tags []tag.Tag
	Data AggregationData
}

func (r *Row) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("{ ")
	buffer.WriteString("{ ")
	for _, t := range r.Tags {
		buffer.WriteString(fmt.Sprintf("{%v %v}", t.Key.Name(), t.Value))
	}
	buffer.WriteString(" }")
	buffer.WriteString(fmt.Sprintf("%v", r.Data))
	buffer.WriteString(" }")
	return buffer.String()
}

func (r *Row) Equal(other *Row) bool {
	return reflect.DeepEqual(r.Tags, other.Tags) && r.Data.Equal(other.Data)
}
