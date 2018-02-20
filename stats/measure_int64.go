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

package stats

// Int64 creates a new measure of type Int64Measure. It returns an
// error if a measure with the same name already exists.
func Int64(m *Measure) (func(int64) Measurement, error) {
	if err := checkName(m.Name); err != nil {
		return discardInt64, err
	}
	_, err := register(m)
	if err != nil {
		return discardInt64, err
	} else {
		record := func(v int64) Measurement {
			return Measurement{Measure: m, Value: float64(v)}
		}
		return record, err
	}
}

func discardInt64(_ int64) Measurement {
	return Measurement{}
}
