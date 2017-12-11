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

// +build go1.6

package trace

func randUint64() uint64 {
	// Copied from later Go version after 1.6 since 1.6 doesn't have it
	// https://github.com/golang/go/blob/70f441bc49afa4e9d10c27d7ed5733c4df7bddd3/src/math/rand/rand.go#L87-L93
	return uint64(traceIDRand.Int63())>>31 | uint64(traceIDRand.Int63())<<32
}
