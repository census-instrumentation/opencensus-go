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

package zipkin

import "github.com/openzipkin/zipkin-go/model"

type Option interface {
	mutateExporter(*Exporter)
}

type remoteEndpoint struct {
	endpoint *model.Endpoint
}

var _ Option = (*remoteEndpoint)(nil)

// WithRemoteEndpoint sets the remote endpoint of the exporter.
func WithRemoteEndpoint(endpoint *model.Endpoint) Option {
	return &remoteEndpoint{endpoint: endpoint}
}

func (re *remoteEndpoint) mutateExporter(exp *Exporter) {
	exp.remoteEndpoint = re.endpoint
}
