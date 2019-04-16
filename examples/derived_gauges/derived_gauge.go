// Copyright 2019, OpenCensus Authors
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

// Command stats implements the stats Quick Start example from:
//   https://opencensus.io/quickstart/go/metrics/
// START entire
package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"bufio"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/metric"
	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricproducer"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// This example demonstrates the use of derived gauges. It is a simple interactive program of consumer
// and producer. User can input number of items to produce. Producer produces specified number of
// items. Consumer consumes randomly consumes 1-5 items in each attempt. It then sleeps randomly
// between 1-10 seconds before the next attempt.
//
// There are two metrics collected to monitor the queue.
// 1. queue_size: It is an instantaneous queue size represented using derived gauge int64.
// 2. queue_seconds_since_processed_last: It is the time elaspsed in seconds since the last time
//    when the queue was consumed. It is represented using derived gauge float64.
type queue struct {
	size         int
	q            []int
	lastConsumed time.Time
	mu           sync.Mutex
}

var q = &queue{}

const (
	maxItemsToConsumePerAttempt = 25
)

func init() {
	q.q = make([]int, 100)
}

// consume randomly dequeues upto 5 items from the queue
func (q *queue) consume() {
	q.mu.Lock()
	defer q.mu.Unlock()

	consumeCount := rand.Int() % maxItemsToConsumePerAttempt
	i := 0
	for i = 0; i < consumeCount; i++ {
		if q.size > 0 {
			q.q = q.q[1:]
			q.size--
		} else {
			break
		}
	}
	if i > 0 {
		q.lastConsumed = time.Now()
	}
}

// produce randomly enqueues upto 5 items from the queue
func (q *queue) produce(count int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i := 0; i < count; i++ {
		v := rand.Int() % 100
		q.q = append(q.q, v)
		q.size++
	}
	fmt.Printf("queued %d items, queue size is %d\n", count, q.size)
}

func (q *queue) runConsumer(interval int, cQuit chan bool) {
	t := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-t.C:
			q.consume()
		case <-cQuit:
			t.Stop()
			return
		}
	}
}

// Size reports instantaneous queue size.
// This is the interface supplied while creating an entry for derived gauge int64.
// START toint64
func (q *queue) Size() int64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return int64(q.size)
}

// END toint64

// Elapsed reports time elapsed since the last time an item was consumed from the queue.
// This is the interface supplied while creating an entry for derived gauge float64.
// START tofloat64
func (q *queue) Elapsed() float64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	return time.Now().Sub(q.lastConsumed).Seconds()
}

// END tofloat64

func getInput() int {
	reader := bufio.NewReader(os.Stdin)
	limit := 100
	for {
		fmt.Printf("Enter number of items to put in consumer queue? [1-%d]: ", limit)
		text, _ := reader.ReadString('\n')
		count, err := strconv.Atoi(strings.TrimSuffix(text, "\n"))
		if err == nil {
			if count < 1 || count > limit {
				fmt.Printf("invalid value %s\n", text)
				continue
			}
			return count
		}
		fmt.Printf("error %v\n", err)
	}
}

func doWork() {
	fmt.Printf("Program monitors queue using two derived gauge metrics.\n")
	fmt.Printf("   1. queue_size = the instantaneous size of the queue.\n")
	fmt.Printf("   2. queue_seconds_since_processed_last = the number of seconds elapsed since last time the queue was processed.\n")
	fmt.Printf("Go to http://localhost:9090/metrics to see the metrics.\n\n\n")

	// Take a number of items to queue as an input from the user
	// and enqueue the same number of items on to the consumer queue.
	for {
		count := getInput()
		q.produce(count)
		fmt.Printf("press CTRL+C to terminate the program\n")
	}
}

func createAndStartExporter() {
	// Create Prometheus metrics exporter to verify derived gauge metrics in this example.
	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatalf("Failed to create the prometheus metrics exporter: %v", err)
	}
	http.Handle("/metrics", exporter)
	go func() {
		log.Fatal(http.ListenAndServe(":9090", nil))

	}()
}

func main() {
	createAndStartExporter()

	// Create metric registry and register it with global producer manager.
	// START reg
	r := metric.NewRegistry()
	metricproducer.GlobalManager().AddProducer(r)
	// END reg

	// Create Int64DerviedGauge
	// START size
	queueSizeGauge, err := r.AddInt64DerivedGauge(
		"queue_size",
		metric.WithDescription("Instantaneous queue size"),
		metric.WithUnit(metricdata.UnitDimensionless))
	if err != nil {
		log.Fatalf("error creating queue size derived gauge, error%v\n", err)
	}
	// END size

	// START entrySize
	err = queueSizeGauge.UpsertEntry(q.Size)
	if err != nil {
		log.Fatalf("error getting queue size derived gauge entry, error%v\n", err)
	}
	// END entrySize

	// Create Float64DerviedGauge
	// START elapsed
	elapsedSeconds, err := r.AddFloat64DerivedGauge(
		"queue_seconds_since_processed_last",
		metric.WithDescription("time elapsed since last time the queue was processed"),
		metric.WithUnit(metricdata.UnitDimensionless))
	if err != nil {
		log.Fatalf("error creating queue_seconds_since_processed_last derived gauge, error%v\n", err)
	}
	// END elapsed

	// START entryElapsed
	err = elapsedSeconds.UpsertEntry(q.Elapsed)
	if err != nil {
		log.Fatalf("error getting queue_seconds_since_processed_last derived gauge entry, error%v\n", err)
	}
	// END entryElapsed

	cQuit := make(chan bool)
	defer func() {
		cQuit <- true
		close(cQuit)
	}()

	// Run consumer and producer
	go q.runConsumer(5, cQuit)

	for {
		doWork()
	}
}

// END entire