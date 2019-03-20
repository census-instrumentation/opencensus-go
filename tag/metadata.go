// Copyright 2019, OpenCensus Authors
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

package tag

const (
	// valueTTLNoPropagation prevents tag from propagating.
	valueTTLNoPropagation = 0

	// valueTTLUnlimitedPropagation allows tag to propagate without any limits on number of hops.
	valueTTLUnlimitedPropagation = -1
)

type metadatas struct {
	ttl int
}

// Metadata applies metadatas specified by the function.
type Metadata func(*metadatas)

// Not exported for the moment because we want only TTLNoPropagation and TTLUnlimitedPropagation
func withTTL(ttl int) Metadata {
	return func(m *metadatas) {
		m.ttl = ttl
	}
}

var (
	// TTLNoPropagation applies metadata with ttl value of valueTTLNoPropagation.
	// It is predefined for convenience.
	TTLNoPropagation = withTTL(valueTTLNoPropagation)

	// TTLUnlimitedPropagation applies metadata with ttl value of valueTTLUnlimitedPropagation.
	// It is predefined for convenience.
	TTLUnlimitedPropagation = withTTL(valueTTLUnlimitedPropagation)
)
