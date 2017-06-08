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

type measureFloat64 struct {
	name        string
	description string
	viewAggs    map[*View]struct{}
}

func (md *measureFloat64) isMeasure() bool { return true }

func (md *measureFloat64) Is(v float64) Measurement {
	return &measurementFloat64{
		m: md,
		v: v,
	}
}

func createMeasureFloat64(name string, description string) MeasureFloat64 {
	return &measureFloat64{
		name:        name,
		description: description,
		viewAggs:    make(map[*View]struct{}),
	}
}

type measureInt64 struct {
	name        string
	description string
	viewAggs    map[*View]struct{}
}

func (md *measureInt64) isMeasure() bool { return true }

func (md *measureInt64) Is(v int64) Measurement {
	return &measurementInt64{
		m: md,
		v: v,
	}
}

func createMeasureInt64(name string, description string) MeasureInt64 {
	return &measureInt64{
		name:        name,
		description: description,
		viewAggs:    make(map[*View]struct{}),
	}
}
