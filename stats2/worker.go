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

package stats2

import (
	"context"
	"time"
)

// RecordFloat64 records a float64 value against a measure and the tags passed
// as part of the context.
func RecordFloat64(ctx context.Context, mf MeasureFloat64, v float64) {}

// RecordInt64 records an int64 value against a measure and the tags passed as
// part of the context.
func RecordInt64(ctx context.Context, mf MeasureInt64, v int64) {}

// Record records one or multiple measurements with the same tags at once.
func Record(ctx context.Context, ms []Measurement) {}

// SetCallbackPeriod sets the minimum and maximum periods for aggregation
// reporting for all registered views in the program. The maximum period is
// only advisory; reports may be generated less frequently than this. The
// default period is determined by internal memory usage.  Calling
// SetCallbackPeriod with either argument equal to zero re-enables the default
// behavior.
func SetCallbackPeriod(min, max time.Duration) {}
