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

package metricexporter

import (
	"context"
	"encoding/json"
	"go.opencensus.io/metric"
	"log"
	"os"
)

// Log is a metrics exporter that periodically logs all metric in JSON
// format.
type Log struct {
	Push
	// Logger is where metrics will be written. By default, a logger
	// that logs to standard error will be configured.
	Logger interface {
		Println(...interface{})
	}
}

// NewLogging calls NewLoggingWithRegistry with the default registry.
func NewLogging() *Log {
	le := &Log{}
	le.Push.Init(le.log)
	le.Logger = log.New(os.Stderr, "", 0)
	return le
}

func (le *Log) log(_ context.Context, ms []*metric.Metric) error {
	for _, m := range ms {
		bb, err := json.Marshal(m)
		if err != nil {
			return err
		}
		le.Logger.Println(string(bb))
	}
	return nil
}
