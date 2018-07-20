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

package trace

import "go.opencensus.io/trace/internal"

// Default limits for the number of attributes, message events and links on each span
// in order to prevent unbounded memory increase for long-running spans.
// These defaults can be overriden with trace.ApplyConfig.
// These defaults can also be overriden per-span by using trace.StartOptions
// when creating a new span.
// TODO: Add an annnoation limit when priorities are implemented.
const (
	DefaultMaxAttributes    = 32
	DefaultMaxMessageEvents = 128
	DefaultMaxLinks         = 32
)

// Config represents the global tracing configuration.
type Config struct {
	// DefaultSampler is the default sampler used when creating new spans.
	DefaultSampler Sampler

	// IDGenerator is for internal use only.
	IDGenerator internal.IDGenerator

	// The below config options must be set with a GlobalOption.
	// maxAttributes sets a global limit on the number of attributes.
	maxAttributes int
	// maxMessageEvents sets a global limit on the number of message events.
	maxMessageEvents int
	// maxLinks sets a global limit on the number of links.
	maxLinks int
}

// GlobalOption apply changes to global tracing configuration.
type GlobalOption func(*Config)

// WithDefaultMaxAttributes sets the default limit on the number of attributes.
func WithDefaultMaxAttributes(limit int) GlobalOption {
	return func(c *Config) {
		c.maxAttributes = limit
	}
}

// WithDefaultMaxMessageEvents sets the default limit on the number of message events.
func WithDefaultMaxMessageEvents(limit int) GlobalOption {
	return func(c *Config) {
		c.maxMessageEvents = limit
	}
}

// WithDefaultMaxLinks sets the default limit on the number of links.
func WithDefaultMaxLinks(limit int) GlobalOption {
	return func(c *Config) {
		c.maxLinks = limit
	}
}

// ApplyConfig applies changes to the global tracing configuration.
//
// Fields not provided in the given config are going to be preserved.
func ApplyConfig(cfg Config, o ...GlobalOption) {
	c := *config.Load().(*Config)
	if cfg.DefaultSampler != nil {
		c.DefaultSampler = cfg.DefaultSampler
	}
	if cfg.IDGenerator != nil {
		c.IDGenerator = cfg.IDGenerator
	}
	for _, op := range o {
		op(&c)
	}
	config.Store(&c)
}
