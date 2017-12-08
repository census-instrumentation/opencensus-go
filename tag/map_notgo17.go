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

// +build !go1.6,!go1.7

package tag

import "sort"

func SortKeysByName(keys []Key) {
	sort.Slice(keys, func(i, j int) bool { return keys[i].Name() < keys[j].Name() })
}

func SortTagsByKeyName(tags []Tag) {
	sort.Slice(tags, func(i, j int) bool { return tags[i].Key.Name() < tags[j].Key.Name() })
}
