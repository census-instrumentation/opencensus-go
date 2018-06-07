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

// Package zpages implements a collection of HTML pages that display RPC stats
// and trace data, and also functions to write that same data in plain text to
// an io.Writer.
//
// Users can also embed the HTML for stats and traces in custom status pages.
//
// zpages are currrently work-in-process and cannot display minutely and
// hourly stats correctly.
//
// Performance
//
// Installing the zpages has a performance overhead because additional traces
// and stats will be collected in-process. In most cases, we expect this
// overhead will not be significant but it depends on many factors, including
// how many spans your process creates and how richly annotated they are.
package zpages // import "go.opencensus.io/zpages"

import (
	"fmt"
	"net/http"
	"sync"

	"go.opencensus.io/internal"
)

// Handler is deprecated: Use AppendHandlers.
var Handler http.Handler

var enableOnce sync.Once

func init() {
	mux := http.NewServeMux()
	AppendHandlers("", mux)
	Handler = mux
}

// AppendHandlers adds the z-pages to the given ServeMux rooted at pathPrefix.
func AppendHandlers(pathPrefix string, mux *http.ServeMux) *http.ServeMux {
	enableOnce.Do(func() {
		internal.LocalSpanStoreEnabled = true
		registerRPCViews()
	})
	mux.HandleFunc(fmt.Sprintf("%s/rpcz", pathPrefix), rpczHandler)
	mux.HandleFunc(fmt.Sprintf("%s/tracez", pathPrefix), tracezHandler)
	mux.Handle(fmt.Sprintf("%s/public/", pathPrefix), http.FileServer(fs))
	return mux
}
