// Copyright 2017 Google Inc.
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

import "github.com/google/working-instrumentation-go/tags"

// View is the generic interface defining the various type of views.
type View interface {
	Name() string        // Name returns the name of a View.
	Description() string // Description returns the description of a View.
	addSubscription(c chan *ViewData)
	deleteSubscription(c chan *ViewData)
	subscriptionExists(c chan *ViewData) bool
	subscriptionsCount() int
	subscriptions() map[chan *ViewData]subscription

	startCollectingForAdhoc()
	stopCollectingForAdhoc()

	isCollecting() bool

	aggregation() Aggregation
	window() Window
	measure() Measure
	collectedRows() []*Row
}

// A ViewData is a set of rows about usage of the single measure associated
// with the given view during a particular window. Each row is specific to a
// unique set of tags.
type ViewData struct {
	v    View
	rows []*Row
}

// Row is the collected value for a specific set of key value pairs a.k.a tags.
type Row struct {
	Tags           *tags.TagSet
	AggregateValue AggregateValue
}
