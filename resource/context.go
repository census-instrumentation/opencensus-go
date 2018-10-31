// Copyright 2018, OpenCensus Authors
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

package resource

import "context"

type resourceKey struct{}

// NewContext returns a new context with the given resource added.
// For now, this is only supported for use in Gauges but will eventually
// also be supported for Views.
func NewContext(ctx context.Context, resource *Resource) context.Context {
	return context.WithValue(ctx, resourceKey{}, resource)
}

// FromContext extracts the resource from the context.
func FromContext(ctx context.Context) (resource *Resource, ok bool) {
	if val := ctx.Value(resourceKey{}); val != nil {
		return val.(*Resource), true
	}
	return nil, false
}
