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

// Measure is the interface for all measure types. A measure is required when
// defining a view.
type Measure interface {
	Name() string
	addView(v View)
	removeView(v View)
	viewsCount() int
}

// Measurement is the interface for all measurement types. Measurements are
// required when recording stats.
type Measurement interface {
	isMeasurement() bool
}
