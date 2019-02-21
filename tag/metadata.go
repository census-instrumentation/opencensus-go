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

package tag

const (
	// TagTTLNoPropagation prevents tag from propagating.
	TagTTLNoPropagation = 0

	// TagTTLUnlimitedPropagation allows tag to propagate without any limits on number of hops.
	TagTTLUnlimitedPropagation = -1
)

// Metadata represents a tag Metadata.
type Metadata struct {
	ttl int
}

var (
	// NonPropagatingMetadata has non propagating property. It is predefined for convenience.
	NonPropagatingMetadata = &Metadata{}

	// PropagatingMetadata has unlimited propagating property. It is predefined for convenience.
	PropagatingMetadata = &Metadata{TagTTLUnlimitedPropagation}
)
