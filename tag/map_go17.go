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

// +build go1.6,go1.7

package tag

import "sort"

type keySlice []Key

func (ks keySlice) Len() int           { return len(ks) }
func (ks keySlice) Less(i, j int) bool { return ks[i].Name() < ks[j].Name() }
func (ks keySlice) Swap(i, j int)      { ks[i], ks[j] = ks[j], ks[i] }

func SortKeysByName(keys []Key) {
	sort.Sort(keySlice(keys))
}

type tagsByName []Tag
func (tn tagsByName) Len() int { return len(tn) }
func (tn tagsByName) Less(i, j int) bool { return tn[i].Key.Name() < tn[j].Key.Name() }
func (tn tagsByName) Swap(i, j int) { tn[i], tn[j] = tn[j], tn[i] }

func SortTagsByKeyName(tags []Tag) {
	sort.Sort(tagsByName(tags))
}
