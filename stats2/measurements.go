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

import "context"

type measurementFloat64 struct {
	m *measureFloat64
	v float64
}

func (md *measurementFloat64) record(ts tagsSet) {
	for v := range md.m.viewAggs {
		// TODO(acetechnologist): record
		// v.Record(ts, md.v)
	}
}

type measurementInt64 struct {
	m *measureInt64
	v int64
}

func (md *measurementInt64) record(ctx context.Context) {
	for v := range md.m.viewAggs {
		// TODO(acetechnologist): record
		// v.Record(ts, md.v)
	}
}
