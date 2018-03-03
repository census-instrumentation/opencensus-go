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

// Package datadog contains an OpenCensus stats exporter for Datadog.
package datadog // import "go.opencensus.io/exporter/datadog"

import (
	"io"
	"log"
	"os"
	"reflect"
	"regexp"

	"github.com/DataDog/datadog-go/statsd"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

const (
	defaultStatsAddr = "127.0.0.1:8125"
)

// Client defines the subset of the dogstatsd api used by the datadog exporter
type Client interface {
	// Counts occurrences of an instance
	Count(name string, value int64, tags []string, rate float64) error

	// Gauge measures the value of a metric at a point in time
	Gauge(name string, value float64, tags []string, rate float64) error

	// Histogram measures the statistical distribution of a value
	Histogram(name string, value float64, tags []string, rate float64) error
}

type config struct {
	statsAddr string // endpoint for dogstatd agent; defaults to 127.0.0.1:8125
	client    Client
	logger    *log.Logger // output error messages
}

// Option provides functional options for the Exporter
type Option interface {
	apply(c *config)
}

type optionFunc func(c *config)

func (fn optionFunc) apply(c *config) {
	fn(c)
}

// WithOutput - optional writer for error messages
func WithOutput(w io.Writer) Option {
	return optionFunc(func(c *config) {
		c.logger = log.New(w, "", log.LstdFlags)
	})
}

// WithStatsAddr configure stats endpoint
func WithStatsAddr(addr string) Option {
	return optionFunc(func(c *config) {
		c.statsAddr = addr
	})
}

// WithClient allows user provided dogstatsd client to be provided
func WithClient(client Client) Option {
	return optionFunc(func(c *config) {
		c.client = client
	})
}

// Exporter is an implementation of view.Exporter that publishes metrics to Datadog
type Exporter struct {
	logger     *log.Logger
	lastValues *lastValues    // lastValues holds the last values we've seen for a given set of view-tags
	client     Client         // client is dogstatd client
	bak        *statsd.Client // client is dogstatd client
}

// buildConfig returns a configuration with all the options and default values applied
func buildConfig(opts ...Option) (config, error) {
	var c config
	for _, opt := range opts {
		opt.apply(&c)
	}

	if c.statsAddr == "" {
		c.statsAddr = defaultStatsAddr
	}
	if c.logger == nil {
		c.logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	if c.client == nil {
		client, err := statsd.New(c.statsAddr)
		if err != nil {
			return config{}, err
		}
		c.client = client
	}

	return c, nil
}

// NewExporter returns a new Datadog stats Exporter
func NewExporter(opts ...Option) (*Exporter, error) {
	c, err := buildConfig(opts...)
	if err != nil {
		return nil, err
	}

	var (
		lastValues = newLastValues()
		exporter   = Exporter{
			lastValues: lastValues,
			client:     c.client,
			logger:     c.logger,
		}
	)

	return &exporter, nil
}

// ExportView implements view.Exporter
func (e *Exporter) ExportView(data *view.Data) {
	go e.publish(data)
}

// publishCount publishes CountAggregation metrics via dogstatsd
func (e *Exporter) publishCount(name string, data, previous view.AggregationData, tags []string, rate float64) {
	var (
		ptr   = data.(*view.CountData)
		value = *ptr
	)

	if lastValue, ok := previous.(*view.CountData); ok {
		value = value - *lastValue
	}

	if err := e.client.Count(name, int64(value), tags, rate); err != nil {
		e.logger.Printf("unable to publish Count metric, %v\n", err)
	}
}

// publishSum publishes SumAggregation metrics via dogstatsd
func (e *Exporter) publishSum(name string, data, previous view.AggregationData, tags []string, rate float64) {
	var (
		ptr   = data.(*view.SumData)
		value = *ptr
	)

	if lastValue, ok := previous.(*view.SumData); ok {
		value = value - *lastValue
	}

	if err := e.client.Count(name, int64(value), tags, rate); err != nil {
		e.logger.Printf("unable to publish Count metric, %v\n", err)
	}
}

// publishMean publishes MeanAggregation metrics via dogstatsd
func (e *Exporter) publishMean(name string, data view.AggregationData, tags []string, rate float64) {
	var (
		value = data.(*view.MeanData)
	)

	if err := e.client.Gauge(name, value.Mean, tags, rate); err != nil {
		e.logger.Printf("unable to publish Gauge metric, %v\n", err)
	}
}

// publishDistribution publishes DistributionAggregation metrics via dogstatsd
func (e *Exporter) publishDistribution(name string, buckets view.DistributionAggregation, data, previous view.AggregationData, tags []string, rate float64) {
	var (
		value             = data.(*view.DistributionData)
		length            = len(value.CountPerBucket)
		lastValue, lastOk = previous.(*view.DistributionData)
		lastLength        int
	)

	if lastOk {
		lastLength = len(lastValue.CountPerBucket)
	}

	for index, bucket := range buckets {
		var count int64
		if index < length {
			count = value.CountPerBucket[index]
		}
		if index < lastLength {
			count = count - lastValue.CountPerBucket[index]
		}

		for i := int64(0); i < count; i++ {
			e.client.Histogram(name, bucket, tags, rate)
		}
	}
}

// publishRow publishes a single row of data
func (e *Exporter) publishRow(myView *view.View, row *view.Row, timestamp float64) {
	var (
		name, nameOk = fixMetricName(myView.Name)
		hasher       = borrow()
		tagHash      = hasher.Hash(row.Tags)
		tags         = makeAllTags(row.Tags)
	)
	defer release(hasher)

	if !nameOk {
		return
	}

	last, _ := e.lastValues.lookup(myView.Name, tagHash)
	defer func() {
		e.lastValues.store(myView.Name, tagHash, row.Data)
	}()

	switch aggregation := myView.Aggregation.(type) {
	case view.CountAggregation:
		e.publishCount(name, row.Data, last, tags, 1)

	case view.SumAggregation:
		e.publishSum(name, row.Data, last, tags, 1)

	case view.MeanAggregation:
		e.publishMean(name, row.Data, tags, 1)

	case view.DistributionAggregation:
		e.publishDistribution(name, aggregation, row.Data, last, tags, 1)

	default:
		e.logger.Printf("don't know how to handle aggregation, %v.  metric dropped.", reflect.TypeOf(myView.Aggregation).String())
	}
}

// publish metrics; should be called from a separate goroutine
func (e *Exporter) publish(data *view.Data) {
	var (
		timestamp = float64(data.Start.Unix())
	)

	for _, row := range data.Rows {
		e.publishRow(data.View, row, timestamp)
	}
}

// makeTag converts an opencensus tag into datadog compatible tags
func makeTag(tag tag.Tag) string {
	return tag.Key.Name() + ":" + tag.Value
}

// makeAllTags converts a slice of opencensus tag to datadog tags
func makeAllTags(tags []tag.Tag) []string {
	var tagStrings []string
	for _, t := range tags {
		tagStrings = append(tagStrings, makeTag(t))
	}

	return tagStrings
}

var (
	reInvalidPrefix         = regexp.MustCompile(`^[^a-zA-Z]+`)
	reSlashes               = regexp.MustCompile(`/`)
	reInvalidNameCharacters = regexp.MustCompile(`[^A-Za-z0-9_.]`)
)

const (
	// maxMetricNameLength contains the max length of a datadog metric
	maxMetricNameLength = 200
)

// fixMetricName fixes metric name as per datadog rules
// See https://docs.datadoghq.com/developers/metrics/
//  * Metric names must start with a letter
//  * Can only contain ASCII alphanumerics, underscore and periods (other characters gets converted to underscores)
//  * Should not exceed 200 characters (though less than 100 is generally preferred from a UI perspective)
//  * Unicode is not supported
//  * We recommend avoiding spaces
func fixMetricName(name string) (string, bool) {
	if reSlashes.MatchString(name) {
		name = reSlashes.ReplaceAllString(name, ".")
	}

	if reInvalidNameCharacters.MatchString(name) {
		name = reInvalidNameCharacters.ReplaceAllString(name, "_")
	}

	if reInvalidPrefix.MatchString(name) {
		name = reInvalidPrefix.ReplaceAllString(name, "")
	}

	if length := len(name); length > maxMetricNameLength {
		name = name[0:maxMetricNameLength]

	} else if length == 0 {
		return "", false
	}

	return name, true
}
