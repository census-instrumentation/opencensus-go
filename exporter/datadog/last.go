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

package datadog // import "go.opencensus.io/exporter/datadog"

import (
	"encoding/binary"
	"hash"
	"sync"
	"time"

	"github.com/cespare/xxhash"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

type lastValue struct {
	createdAt int64
	data      view.AggregationData
}

type lastValues struct {
	hasher  *hasher
	mutex   sync.Mutex
	content map[string]map[uint64]*lastValue
}

func newLastValues() *lastValues {
	var (
		content = map[string]map[uint64]*lastValue{}
	)

	return &lastValues{
		content: content,
	}
}

func (l *lastValues) store(viewName string, tagHash uint64, data view.AggregationData) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	contentByTags, ok := l.content[viewName]
	if !ok {
		contentByTags = map[uint64]*lastValue{}
		l.content[viewName] = contentByTags
	}

	v, ok := contentByTags[tagHash]
	if !ok {
		v = &lastValue{}
		contentByTags[tagHash] = v
	}

	v.createdAt = time.Now().Unix()
	v.data = data
}

func (l *lastValues) lookup(viewName string, tagHash uint64) (view.AggregationData, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	contentByTags, ok := l.content[viewName]
	if !ok {
		return nil, false
	}

	v, ok := contentByTags[tagHash]
	if !ok {
		return nil, false
	}

	return v.data, true
}

type hasher struct {
	hash64 hash.Hash64
	buffer []byte
}

const (
	sep = ":"
)

func (h *hasher) Hash(tags []tag.Tag) uint64 {
	//if len(tags) > 1 {
	//	sort.Sort(byName(tags))
	//}
	//
	h.hash64.Reset()
	for _, t := range tags {
		h.buffer = h.buffer[:0]
		h.buffer = append(h.buffer, t.Key.Name()...)
		h.buffer = append(h.buffer, sep...)
		h.buffer = append(h.buffer, t.Value...)
		h.buffer = append(h.buffer, sep...)
		h.hash64.Write(h.buffer)
	}

	h.buffer = h.buffer[:0]
	h.buffer = h.hash64.Sum(h.buffer)
	return binary.BigEndian.Uint64(h.buffer)
}

var (
	hasherPool = &sync.Pool{
		New: func() interface{} {
			return newHasher()
		},
	}
)

func newHasher() *hasher {
	var (
		hash64 = xxhash.New()
		buffer = make([]byte, 0, 128)
	)

	return &hasher{
		hash64: hash64,
		buffer: buffer,
	}
}

func borrow() *hasher {
	return hasherPool.Get().(*hasher)
}

func release(h *hasher) {
	hasherPool.Put(h)
}
