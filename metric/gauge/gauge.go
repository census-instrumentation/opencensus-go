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
	"math"
	"sync"
	"sync/atomic"
	"time"

	"go.opencensus.io/internal/tagencoding"

	"go.opencensus.io/metric/metricexport"

	"go.opencensus.io/metric"
)

// Gauge represents a quantity that can go up an down, for example queue depth
// or number of outstanding requests.
//
// Gauge maintains a value for each combination of of label values passed to
// the Set or Add methods.
//
// Gauge should not be used directly, use Float64 or Int64.
type Gauge struct {
	vals     sync.Map
	desc     metricexport.Descriptor
	start    time.Time
	keys     []string
	newEntry func() gaugeEntry
}

type gaugeEntry interface {
	read(t time.Time) metricexport.Point
}

var _ metricexport.Producer = (*Gauge)(nil)

// Read returns the current values of the gauge as a metric for export.
func (g *Gauge) Read() []*metricexport.Metric {
	now := time.Now()
	m := &metricexport.Metric{
		Descriptor: g.desc,
	}
	g.vals.Range(func(k, v interface{}) bool {
		entry := v.(gaugeEntry)
		key := k.(string)
		labelVals := g.labelValues(key)
		m.TimeSeries = append(m.TimeSeries, &metricexport.TimeSeries{
			StartTime:   g.start,
			LabelValues: labelVals,
			Points: []metricexport.Point{
				entry.read(now),
			},
		})
		return true
	})
	return []*metricexport.Metric{m}
}

func (g *Gauge) getEntry(labelVals []metric.LabelValue) gaugeEntry {
	if len(labelVals) != len(g.keys) {
		panic("must supply the same number of label values as keys used to construct this Gauge")
	}
	mapKey := g.mapKey(labelVals)
	if entry, ok := g.vals.Load(mapKey); ok {
		return entry.(gaugeEntry)
	} else {
		entry, _ := g.vals.LoadOrStore(mapKey, g.newEntry())
		return entry.(gaugeEntry)
	}
}

func (g *Gauge) mapKey(labelVals []metric.LabelValue) string {
	vb := &tagencoding.Values{}
	for _, v := range labelVals {
		b := make([]byte, 1, len(v.Value)+1)
		if v.Present {
			b[0] = 1
			b = append(b, []byte(v.Value)...)
		}
		vb.WriteValue(b)
	}
	return string(vb.Bytes())
}

func (g *Gauge) labelValues(s string) []metric.LabelValue {
	vals := make([]metric.LabelValue, 0, len(g.keys))
	vb := &tagencoding.Values{Buffer: []byte(s)}
	for range g.keys {
		v := vb.ReadValue()
		if v[0] == 0 {
			vals = append(vals, metric.LabelValue{})
		} else {
			vals = append(vals, metric.NewLabelValue(string(v[1:])))
		}
	}
	return vals
}

// Float64 represents a float64 value that can go up and down.
//
// Float64 maintains a float64 value for each combination of of label values
// passed to the Set or Add methods.
type Float64 struct {
	Gauge
}

type float64Entry struct {
	val uint64
}

func (e *float64Entry) read(t time.Time) metricexport.Point {
	return metricexport.NewFloat64Point(t, math.Float64frombits(atomic.LoadUint64(&e.val)))
}

// NewFloat64WithRegistry creates a new gauge with a float64 value.
func NewFloat64(name, description string, unit metric.Unit, keys ...string) *Float64 {
	g := &Float64{
		Gauge{
			newEntry: func() gaugeEntry {
				return new(float64Entry)
			},
			keys:  keys,
			start: time.Now(),
			desc: metricexport.Descriptor{
				Name:        name,
				Description: description,
				Unit:        unit,
				LabelKeys:   keys,
			},
		},
	}
	return g
}

// Set sets the current gauge value.
func (g *Float64) Set(val float64, labelVals ...metric.LabelValue) {
	ge := g.getEntry(labelVals).(*float64Entry)
	atomic.StoreUint64(&ge.val, math.Float64bits(val))
}

// Add increments the current gauge value by val.
func (g *Float64) Add(val float64, labelVals ...metric.LabelValue) {
	ge := g.getEntry(labelVals).(*float64Entry)
	var swapped bool
	for !swapped {
		oldVal := atomic.LoadUint64(&ge.val)
		newVal := math.Float64bits(math.Float64frombits(oldVal) + val)
		swapped = atomic.CompareAndSwapUint64(&ge.val, oldVal, newVal)
	}
}

// Int64 represents a int64 gauge value that can go up and down.
//
// Int64 maintains an int64 value for each combination of label values passed to the
// Set or Add methods.
type Int64 struct {
	Gauge
}

type int64GaugeValue struct {
	val int64
}

func (v *int64GaugeValue) read(t time.Time) metricexport.Point {
	return metricexport.NewInt64Point(t, atomic.LoadInt64(&v.val))
}

// NewInt64WithRegistry creates a new int64-valued gauge and adds it to the
// given registry.
func NewInt64(name, description string, unit metric.Unit, keys ...string) *Int64 {
	g := &Int64{
		Gauge{
			newEntry: func() gaugeEntry {
				return new(int64GaugeValue)
			},
			keys:  keys,
			start: time.Now(),
			desc: metricexport.Descriptor{
				Name:        name,
				Description: description,
				Unit:        unit,
				LabelKeys:   keys,
			},
		},
	}
	return g
}

// Set sets the current gauge value.
func (g *Int64) Set(val int64, labelVals ...metric.LabelValue) {
	ge := g.getEntry(labelVals).(*int64GaugeValue)
	atomic.StoreInt64(&ge.val, val)
}

// Add increments the current gauge value by val.
func (g *Int64) Add(val int64, labelVals ...metric.LabelValue) {
	ge := g.getEntry(labelVals).(*int64GaugeValue)
	atomic.AddInt64(&ge.val, val)
}
