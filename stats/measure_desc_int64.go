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

package stats

// measureDescInt64 represents a measure description of type int64.
type measureDescInt64 struct {
	*measureDesc
}

func NewMeasureDescInt64(name string, description string, unit *MeasurementUnit) *measureDescInt64 {
	return &measureDescInt64{
		&measureDesc{
			name:         name,
			description:  description,
			unit:         unit,
			aggViewDescs: make(map[ViewDesc]struct{}),
		},
	}
}

func (md *measureDescInt64) Meta() *measureDesc {
	return md.measureDesc
}

func (md *measureDescInt64) CreateMeasurement(v int64) Measurement {
	return &measurementInt64{
		md: md,
		v:  v,
	}
}

