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

package internal

import (
	"sort"
	"time"

	"go.opencensus.io/internal/tagencoding"
	"go.opencensus.io/stats/aggregation"
	"go.opencensus.io/tag"
)

// Row is the collected value for a specific set of key value pairs a.k.a tags.
type Row struct {
	Tags []tag.Tag
	Data aggregation.Data
}

type ViewWorker struct {
	Name                string
	TagKeys             []tag.Key
	MeasureDesc         *MeasureDesc
	NewData             func() aggregation.Data
	NewWindowAggregator func(start time.Time, fn func() aggregation.Data) aggregation.WindowAggregator
	IsCumulative        bool

	startTime  time.Time
	signatures map[string]aggregation.WindowAggregator
}

func (v *ViewWorker) clearRows() {
	v.signatures = make(map[string]aggregation.WindowAggregator)
}

func (v *ViewWorker) collectedRows(now time.Time) []*Row {
	var rows []*Row
	for sig, aggregator := range v.signatures {
		tags := decodeTags([]byte(sig), v.TagKeys)
		row := &Row{tags, aggregator.Collect(now)}
		rows = append(rows, row)
	}
	return rows
}

func (v *ViewWorker) addSample(m *tag.Map, val interface{}, now time.Time) {
	sig := string(encodeWithKeys(m, v.TagKeys))
	// TODO(jbd): Buffer not process.
	aggregator, ok := v.signatures[sig]
	if !ok {
		aggregator = v.NewWindowAggregator(now, v.NewData)
		v.signatures[sig] = aggregator
	}
	aggregator.AddSample(v, now)
}

// encodeWithKeys encodes the map by using values
// only associated with the keys provided.
func encodeWithKeys(m *tag.Map, keys []tag.Key) []byte {
	vb := &tagencoding.Values{
		Buffer: make([]byte, len(keys)),
	}
	for _, k := range keys {
		v, _ := m.Value(k)
		vb.WriteValue([]byte(v))
	}
	return vb.Bytes()
}

// decodeTags decodes tags from the buffer and
// orders them by the keys.
func decodeTags(buf []byte, keys []tag.Key) []tag.Tag {
	vb := &tagencoding.Values{Buffer: buf}
	var tags []tag.Tag
	for _, k := range keys {
		v := vb.ReadValue()
		if v != nil {
			tags = append(tags, tag.Tag{Key: k, Value: string(v)})
		}
	}
	vb.ReadIndex = 0
	sort.Slice(tags, func(i, j int) bool { return tags[i].Key.Name() < tags[j].Key.Name() })
	return tags
}
