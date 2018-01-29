// Package zpages implements a collection of HTML pages that display RPC stats
// and trace data, and also functions to write that same data in plain text to
// an io.Writer.
//
// Users can also embed the HTML for stats and traces in custom status pages.
//
// To add the handlers to the default HTTP request multiplexer with the patterns
// /rpcz and /tracez, call:
// 	zpages.AddDefaultHTTPHandlers()
// If your program does not already start an HTTP server, you can use:
// 	go func() { log.Fatal(http.ListenAndServe(":8080", nil)) }()
package zpages

import (
	"net/http"
	"sync"
)

var once sync.Once

// AddDefaultHTTPHandlers adds handlers for /rpcz and /tracez to the default HTTP request multiplexer.
func AddDefaultHTTPHandlers() {
	once.Do(func() {
		http.HandleFunc("/rpcz", RpczHandler)
		http.HandleFunc("/tracez", TracezHandler)
	})
}
