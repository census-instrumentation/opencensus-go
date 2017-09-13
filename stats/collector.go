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
//

// Package stats defines the stats collection API and its native Go
// implementation.
package stats

import (
	"time"

	"github.com/census-instrumentation/opencensus-go/tags"
)

type collector struct {
	// signatures holds the aggregations values for each unique tag signature
	// (values for all keys) to its Window.
	signatures map[string]aggregator
	// Aggregation is the description of the aggregation to perform for this
	// view.
	a Aggregation

	// window is the window under which the aggregation is performed.
	w Window
}

func (c *collector) addSample(s string, v interface{}, now time.Time) {
	aggregator, ok := c.signatures[s]
	if !ok {
		aggregator = c.w.newAggregator(now, c.a.aggregationValueConstructor())
		c.signatures[s] = aggregator
	}
	aggregator.addSample(v, now)
}

func (c *collector) collectedRows(keys []tags.Key, now time.Time) []*Row {
	var rows []*Row
	for sig, aggregator := range c.signatures {
		ts := tags.ToOrderedTagsSlice(sig, keys)
		rows = append(rows, &Row{
			ts,
			aggregator.retrieveCollected(now),
		})
	}
	return rows
}

func (c *collector) clearRows() {
	c.signatures = make(map[string]aggregator)
}
