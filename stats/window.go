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

import "time"

// Window represents the interval/samples count over which the aggregation
// occurs.
type Window interface {
	isWindow() bool
}

// WindowCumulative indicates that the aggregation occurs over the lifetime of
// the view.
type WindowCumulative struct{}

func (w WindowCumulative) isWindow() bool { return true }

// WindowSlidingTime indicates that the aggregation occurs over a sliding
// window of time: i.e. last n seconds, minutes, hours...
type WindowSlidingTime struct {
	duration     time.Duration
	subIntervals int
}

func (w WindowSlidingTime) isWindow() bool { return true }

// WindowSlidingCount indicates that the aggregation occurs over a sliding
// number of samples.
type WindowSlidingCount struct {
	n       int
	subSets int
}

func (w WindowSlidingCount) isWindow() bool { return true }
