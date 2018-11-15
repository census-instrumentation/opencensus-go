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

package gauge

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"go.opencensus.io/metric"
	"go.opencensus.io/resource"
	"go.opencensus.io/tag"
)

// Gauge represents a quantity that can go up an down, for example queue depth
// or number of outstanding requests.
//
// Gauge maintains a value for each combination of of tag values and resource
// in the context passed to the Set or Add methods.
//
// Gauge should not be used directly, use Double or Int64.
type Gauge struct {
	vals     sync.Map
	desc     *metric.Descriptor
	start    time.Time
	keys     []tag.Key
	newEntry func() gaugeEntry
}

type gaugeEntry interface {
	read(t time.Time) metric.Point
}

type resourceKey struct {
	type_, labels string
}

type gaugeKey struct {
	tags string
	res  resourceKey
}

var _ metric.Producer = (*Gauge)(nil)

// Read returns the current values of the gauge for each combination of tags
// and resource that has been seen.
func (g *Gauge) Read() []*metric.Metric {
	now := time.Now()
	metrics := make(map[resourceKey]*metric.Metric)
	g.vals.Range(func(k, v interface{}) bool {
		entry := v.(gaugeEntry)
		key := k.(gaugeKey)
		labelVals := g.labelValues(key.tags)
		m, ok := metrics[key.res]
		if !ok {
			resLabels, err := resource.DecodeLabels(key.res.labels)
			if err != nil {
				// should never happen since we encoded them in the first place
				panic(err)
			}
			m = &metric.Metric{
				Descriptor: g.desc,
				Resource: &resource.Resource{
					Type:   key.res.type_,
					Labels: resLabels,
				},
			}
			metrics[key.res] = m
		}
		m.TimeSeries = append(m.TimeSeries, &metric.TimeSeries{
			StartTime:   g.start,
			LabelValues: labelVals,
			Points: []metric.Point{
				entry.read(now),
			},
		})
		return true
	})
	result := []*metric.Metric{nil}[:0] // optimize for case len(metrics) == 1
	for _, m := range metrics {
		result = append(result, m)
	}
	return result
}

func (g *Gauge) getEntry(ctx context.Context) gaugeEntry {
	mapKey := g.mapKey(ctx)
	if entry, ok := g.vals.Load(mapKey); ok {
		return entry.(gaugeEntry)
	} else {
		entry, _ := g.vals.LoadOrStore(mapKey, g.newEntry())
		return entry.(gaugeEntry)
	}
}

func (g *Gauge) mapKey(ctx context.Context) gaugeKey {
	tm := tag.FromContext(ctx)
	res, ok := resource.FromContext(ctx)
	if !ok {
		res = &resource.Resource{}
	}
	return gaugeKey{
		tags: tagsToString(tm),
		res:  resourceKey{type_: res.Type, labels: resource.EncodeLabels(res.Labels)},
	}
}

func tagsToString(tm *tag.Map) string {
	return string(tag.Encode(tm))
}

func (g *Gauge) labelValues(s string) []metric.LabelValue {
	tm, err := tag.Decode([]byte(s))
	if err != nil {
		panic(err) // should never happen, since we called Encode
	}
	var vals []metric.LabelValue
	for _, k := range g.keys {
		val, ok := tm.Value(k)
		if ok {
			vals = append(vals, metric.NewLabelValue(val))
		} else {
			vals = append(vals, nil)
		}
	}
	return vals
}

// Double represents a float64 value that can go up and down.
//
// Double maintains a value for each combination of of tag values and resource
// in the context passed to the Set or Add methods.
type Double struct {
	Gauge
}

type doubleEntry struct {
	val uint64
}

func (e *doubleEntry) read(t time.Time) metric.Point {
	return metric.NewDoublePoint(t, math.Float64frombits(atomic.LoadUint64(&e.val)))
}

// NewDoubleWithRegistry creates a new gauge with a float64 value.
func NewDouble(name, description string, unit metric.Unit, keys ...tag.Key) *Double {
	labelKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		labelKeys = append(labelKeys, k.Name())
	}
	g := &Double{
		Gauge{
			newEntry: func() gaugeEntry {
				return new(doubleEntry)
			},
			keys:  keys,
			start: time.Now(),
			desc: &metric.Descriptor{
				Name:        name,
				Description: description,
				Unit:        unit,
				LabelKeys:   labelKeys,
			},
		},
	}
	return g
}

// Set sets the current gauge value.
func (g *Double) Set(ctx context.Context, val float64) {
	ge := g.getEntry(ctx).(*doubleEntry)
	atomic.StoreUint64(&ge.val, math.Float64bits(val))
}

// Add increments the current gauge value by val.
func (g *Double) Add(ctx context.Context, val float64) {
	ge := g.getEntry(ctx).(*doubleEntry)
	var swapped bool
	for !swapped {
		oldVal := atomic.LoadUint64(&ge.val)
		newVal := math.Float64bits(math.Float64frombits(oldVal) + val)
		swapped = atomic.CompareAndSwapUint64(&ge.val, oldVal, newVal)
	}
}

// Int64 represents a int64 gauge value that can go up and down.
//
// Int64 maintains a value for each combination of of tag values
// in the context passed to the Set or Add methods.
type Int64 struct {
	Gauge
}

type int64GaugeValue struct {
	val int64
}

func (v *int64GaugeValue) read(t time.Time) metric.Point {
	return metric.NewInt64Point(t, atomic.LoadInt64(&v.val))
}

// NewInt64WithRegistry creates a new int64-valued gauge and adds it to the
// given registry.
func NewInt64(name, description string, unit metric.Unit, keys ...tag.Key) *Int64 {
	labelKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		labelKeys = append(labelKeys, k.Name())
	}
	g := &Int64{
		Gauge{
			newEntry: func() gaugeEntry {
				return new(int64GaugeValue)
			},
			keys:  keys,
			start: time.Now(),
			desc: &metric.Descriptor{
				Name:        name,
				Description: description,
				Unit:        unit,
				LabelKeys:   labelKeys,
			},
		},
	}
	return g
}

// Set sets the current gauge value.
func (g *Int64) Set(ctx context.Context, val int64) {
	ge := g.getEntry(ctx).(*int64GaugeValue)
	atomic.StoreInt64(&ge.val, val)
}

// Add increments the current gauge value by val.
func (g *Int64) Add(ctx context.Context, val int64) {
	ge := g.getEntry(ctx).(*int64GaugeValue)
	atomic.AddInt64(&ge.val, val)
}
