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

type measureDescFloat64 struct {
	name         string
	description  string
	aggViewDescs map[ViewDesc]struct{}
}

func (md *measureDescFloat64) isMeasureDesc() bool { return true }

func (md *measureDescFloat64) CreateMeasurement(v float64) Measurement {
	return &measurementFloat64{
		m: md,
		v: v,
	}
}

type measureDescInt64 struct {
	name         string
	description  string
	aggViewDescs map[ViewDesc]struct{}
}

func (md *measureDescInt64) isMeasureDesc() bool { return true }

func (md *measureDescInt64) CreateMeasurement(v int64) Measurement {
	return &measurementInt64{
		m: md,
		v: v,
	}
}

func registerMeasureDescFloat64(name string, description string) MeasureDescFloat64 {
	return &measureDescFloat64{
		name:         name,
		description:  description,
		aggViewDescs: make(map[ViewDesc]struct{}),
	}
}
