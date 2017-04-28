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

// Package tracing defines the two handlers (ClientHandler and ServerHandler)
// for processing GRPC lifecycle events and process tracing data. Both are
// different implementations of the "google.golang.org/grpc/stats.Handler"
// interface.
package tracing

// traceKey is the metadata key used to identify the tracing info in the
// GRPC context metadata.
const traceKey = "grpc-tracing-bin"
