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

// Package propagation contains implementations of common HTTP propagation
// formats. Most formats are found in named sub-packages.
package propagation // import "go.opencensus.io/plugin/ochttp/propagation"

import (
	"net/http"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
)

// None returns a propagation format that ignores any incoming headers and
// doesn't write any outbound headers. This is appropriate to use if you do
// not want to continue traces from requests your server receives, or if you do
// not want to propagate traces on outbound requests.
func None() propagation.HTTPFormat {
	return (*noPropagation)(nil)
}

type noPropagation struct{}

func (f *noPropagation) SpanContextFromRequest(_ *http.Request) (trace.SpanContext, bool) {
	return trace.SpanContext{}, false
}

func (f *noPropagation) SpanContextToRequest(_ trace.SpanContext, _ *http.Request) {
}
