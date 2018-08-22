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

// Package resource defines the resource type and provides helpers to derive them as well
// as the generic population through environment variables.
package resource

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	envVarType = "OC_RESOURCE_TYPE"
	envVarTags = "OC_RESOURCE_TAGS"
)

// Resource describes an entity about which data is exposed.
type Resource struct {
	Type string
	Tags map[string]string
}

// EncodeTags encodes a tags to a string as provided via the OC_RESOURCE_TAGS environment variable.
func EncodeTags(tags map[string]string) (s string) {
	i := 0
	for k, v := range tags {
		if i > 0 {
			s += ","
		}
		s += k + "=" + strconv.Quote(v)
		i++
	}
	return s
}

// We accept domain names and paths as tag keys. Values may be quoted or unquoted in general.
// If a value contains whitespaces, =, or " characters, it must always be quoted.
var tagRegex = regexp.MustCompile(`\s*([a-zA-Z0-9-_./]+)=(?:(".*?")|([^\s="]+))\s*,`)

func DecodeTags(s string) (map[string]string, error) {
	m := map[string]string{}
	// Ensure a trailing comma, which allows us to keep the regex simpler
	s = strings.TrimRight(strings.TrimSpace(s), ",") + ","

	for len(s) > 0 {
		match := tagRegex.FindStringSubmatch(s)
		if len(match) == 0 {
			return nil, fmt.Errorf("invalid tag formatting, remainder: %s", s)
		}
		v := match[2]
		if v == "" {
			v = match[3]
		} else {
			var err error
			if v, err = strconv.Unquote(v); err != nil {
				return nil, fmt.Errorf("invalid tag formatting, remainder: %s, err: %s", s, err)
			}
		}
		m[match[1]] = v

		s = s[len(match[0]):]
	}
	return m, nil
}

// FromEnvVars loads resource information from the OC_TYPE and OC_RESOURCE_TAGS environment variables.
func FromEnvVars(context.Context) (*Resource, error) {
	res := &Resource{
		Type: strings.TrimSpace(os.Getenv(envVarType)),
	}
	tags := strings.TrimSpace(os.Getenv(envVarTags))
	if tags == "" {
		return res, nil
	}
	var err error
	if res.Tags, err = DecodeTags(tags); err != nil {
		return nil, err
	}
	return res, nil
}

// Merge resource information from b into a. In case of a collision, a takes precedence.
func Merge(a, b *Resource) *Resource {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	res := &Resource{
		Type: a.Type,
		Tags: map[string]string{},
	}
	for k, v := range a.Tags {
		res.Tags[k] = v
	}
	if res.Type == "" {
		res.Type = b.Type
	}
	for k, v := range b.Tags {
		if _, ok := res.Tags[k]; !ok {
			res.Tags[k] = v
		}
	}
	return res
}

// Detector attempts to detect resource information.
// If the detector cannot find specific information, the respective Resource fields should
// be left empty but no error should be returned.
// An error should only be returned if unexpected errors occur during lookup.
type Detector func(context.Context) (*Resource, error)

// NewDetectorFromResource returns a detector that will always return resource r.
func NewDetectorFromResource(r *Resource) Detector {
	return func(context.Context) (*Resource, error) {
		return r, nil
	}
}

// ChainedDetector returns a Detector that calls all input detectors sequentially an
// merges each result with the previous one.
// It returns on the first error that a sub-detector encounters.
func ChainedDetector(detectors ...Detector) Detector {
	return func(ctx context.Context) (*Resource, error) {
		return DetectAll(ctx, detectors...)
	}
}

// Detectall calls all input detectors sequentially an merges each result with the previous one.
// It returns on the first error that a sub-detector encounters.
func DetectAll(ctx context.Context, detectors ...Detector) (*Resource, error) {
	var res *Resource
	for _, d := range detectors {
		r, err := d(ctx)
		if err != nil {
			return nil, err
		}
		res = Merge(res, r)
	}
	return res, nil
}
