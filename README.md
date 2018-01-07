# OpenCensus Libraries for Go

[![Build Status][travis-image]][travis-url]
[![Windows Build Status][appveyor-image]][appveyor-url]
[![GoDoc][godoc-image]][godoc-url]
[![Gitter chat][gitter-image]][gitter-url]

OpenCensus Go is a Go implementation of OpenCensus, a toolkit for
collecting application performance and behavior monitoring data.
Currently it consists of three major components: tags, stats, and tracing.

This project is still at a very early stage of development. The API is changing
rapidly, vendoring is recommended.


## Installation

```
$ go get -u go.opencensus.io/...
```

## Prerequisites

OpenCensus Go libraries require Go 1.8 or later.

## Exporters

OpenCensus can export instrumentation data to various backends. 
Currently, OpenCensus supports:

* [Prometheus][exporter-prom] for stats
* [OpenZipkin][exporter-zipkin] for traces
* Stackdriver [Monitoring][exporter-sdstats] and [Trace][exporter-sdtrace]

## Tags

Tags represent propagated key-value pairs. They can be propagated using context.Context
in the same process or can be encoded to be transmitted on the wire and decoded back
to a tag.Map at the destination.

### Getting a key by a name

A key is defined by its name. To use a key, a user needs to know its name and type.
Currently, only keys of type string are supported.
Other types will be supported in the future.

[embedmd]:# (tags.go stringKey)
```stringKey
// Copyright 2017, OpenCensus Authors
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

package readme

import (
	"context"
	"log"

	"go.opencensus.io/tag"
)

func tagsExamples() {
	ctx := context.Background()

	// START stringKey
	// Get a key to represent user OS.
	key, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	// END stringKey
	_ = key

	// START tagMap
	osKey, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	userIDKey, err := tag.NewKey("my.org/keys/user-id")
	if err != nil {
		log.Fatal(err)
	}

	tagMap, err := tag.NewMap(ctx,
		tag.Insert(osKey, "macOS-10.12.5"),
		tag.Upsert(userIDKey, "cde36753ed"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// END tagMap

	// START newContext
	ctx = tag.NewContext(ctx, tagMap)
	// END newContext

	// START replaceTagMap
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	// END replaceTagMap

	// START profiler
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	tag.Do(ctx, func(ctx context.Context) {
		// Do work.
		// When profiling is on, samples will be
		// recorded with the key/values from the tag map.
	})
	// END profiler
}
```

### Creating a map of tags associated with keys

tag.Map is a map of tags. Package tags provide a builder to create tag maps.

[embedmd]:# (tags.go tagMap)
```tagMap
// Copyright 2017, OpenCensus Authors
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

package readme

import (
	"context"
	"log"

	"go.opencensus.io/tag"
)

func tagsExamples() {
	ctx := context.Background()

	// START stringKey
	// Get a key to represent user OS.
	key, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	// END stringKey
	_ = key

	// START tagMap
	osKey, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	userIDKey, err := tag.NewKey("my.org/keys/user-id")
	if err != nil {
		log.Fatal(err)
	}

	tagMap, err := tag.NewMap(ctx,
		tag.Insert(osKey, "macOS-10.12.5"),
		tag.Upsert(userIDKey, "cde36753ed"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// END tagMap

	// START newContext
	ctx = tag.NewContext(ctx, tagMap)
	// END newContext

	// START replaceTagMap
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	// END replaceTagMap

	// START profiler
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	tag.Do(ctx, func(ctx context.Context) {
		// Do work.
		// When profiling is on, samples will be
		// recorded with the key/values from the tag map.
	})
	// END profiler
}
```

### Propagating a tag map in a context

To propagate a tag map to downstream methods and RPCs, add a tag map
to the current context. NewContext will return a copy of the current context,
and put the tag map into the returned one.
If there is already a tag map in the current context, it will be replaced.

[embedmd]:# (tags.go newContext)
```newContext
// Copyright 2017, OpenCensus Authors
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

package readme

import (
	"context"
	"log"

	"go.opencensus.io/tag"
)

func tagsExamples() {
	ctx := context.Background()

	// START stringKey
	// Get a key to represent user OS.
	key, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	// END stringKey
	_ = key

	// START tagMap
	osKey, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	userIDKey, err := tag.NewKey("my.org/keys/user-id")
	if err != nil {
		log.Fatal(err)
	}

	tagMap, err := tag.NewMap(ctx,
		tag.Insert(osKey, "macOS-10.12.5"),
		tag.Upsert(userIDKey, "cde36753ed"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// END tagMap

	// START newContext
	ctx = tag.NewContext(ctx, tagMap)
	// END newContext

	// START replaceTagMap
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	// END replaceTagMap

	// START profiler
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	tag.Do(ctx, func(ctx context.Context) {
		// Do work.
		// When profiling is on, samples will be
		// recorded with the key/values from the tag map.
	})
	// END profiler
}
```

In order to update an existing tag map, get the tag map from the current context,
use NewMap and put the new tag map back to the context.

[embedmd]:# (tags.go replaceTagMap)
```replaceTagMap
// Copyright 2017, OpenCensus Authors
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

package readme

import (
	"context"
	"log"

	"go.opencensus.io/tag"
)

func tagsExamples() {
	ctx := context.Background()

	// START stringKey
	// Get a key to represent user OS.
	key, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	// END stringKey
	_ = key

	// START tagMap
	osKey, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	userIDKey, err := tag.NewKey("my.org/keys/user-id")
	if err != nil {
		log.Fatal(err)
	}

	tagMap, err := tag.NewMap(ctx,
		tag.Insert(osKey, "macOS-10.12.5"),
		tag.Upsert(userIDKey, "cde36753ed"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// END tagMap

	// START newContext
	ctx = tag.NewContext(ctx, tagMap)
	// END newContext

	// START replaceTagMap
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	// END replaceTagMap

	// START profiler
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	tag.Do(ctx, func(ctx context.Context) {
		// Do work.
		// When profiling is on, samples will be
		// recorded with the key/values from the tag map.
	})
	// END profiler
}
```


## Stats

### Creating, retrieving and deleting a measure

Create and load measures with units:

[embedmd]:# (stats.go measure)
```measure
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

Retrieve measure by name:

[embedmd]:# (stats.go findMeasure)
```findMeasure
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

Delete measure (this can be useful when replacing a measure by
another measure with the same name):

[embedmd]:# (stats.go deleteMeasure)
```deleteMeasure
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```
However, it is an error to delete a Measure that's used by at least one View. The
View using the Measure has to be unregistered first.

### Creating an aggregation

Currently 4 types of aggregations are supported. The CountAggregation is used to count
the number of times a sample was recorded. The DistributionAggregation is used to
provide a histogram of the values of the samples. The SumAggregation is used to
sum up all sample values. The MeanAggregation is used to calculate the mean of
sample values.

[embedmd]:# (stats.go aggs)
```aggs
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

### Create an aggregation window

Currently only two types of aggregation windows are supported. The Cumulative
is used to continuously aggregate the data received.
The Interval window is used to aggregate the data received over the last specified time interval.
Currently all aggregation types are compatible with all aggregation windows.
Later we might provide aggregation types that are incompatible with some windows.

[embedmd]:# (stats.go windows)
```windows
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

### Creating, registering and unregistering a view

Create and register a view:

[embedmd]:# (stats.go view)
```view
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

Find view by name:

[embedmd]:# (stats.go findView)
```findView
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

Unregister view:

[embedmd]:# (stats.go unregisterView)
```unregisterView
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

Configure the default interval between reports of collected data.
This is a system wide interval and impacts all views. The default
interval duration is 10 seconds. Trying to set an interval with
a duration less than a certain minimum (maybe 1s) should have no effect.

[embedmd]:# (stats.go reportingPeriod)
```reportingPeriod
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

### Recording measurements

Recording usage can only be performed against already registered measures
and their registered views. Measurements are implicitly tagged with the
tags in the context:

[embedmd]:# (stats.go record)
```record
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

### Retrieving collected data for a view

Users need to subscribe to a view in order to retrieve collected data.

[embedmd]:# (stats.go subscribe)
```subscribe
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

Subscribed views' data will be exported via the registered exporters.

[embedmd]:# (stats.go registerExporter)
```registerExporter
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

An example logger exporter is below:

[embedmd]:# (stats.go exporter)
```exporter
// Copyright 2017, OpenCensus Authors
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

// Package readme generates the README.
package readme

import (
	"context"
	"log"
	"time"

	"go.opencensus.io/stats"
)

// README.md is generated with the examples here by using embedmd.
// For more details, see https://github.com/rakyll/embedmd.

func statsExamples() {
	ctx := context.Background()

	// START measure
	videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
	if err != nil {
		log.Fatal(err)
	}
	// END measure
	_ = videoSize

	// START findMeasure
	m := stats.FindMeasure("my.org/video_size")
	if m == nil {
		log.Fatalln("measure not found")
	}
	// END findMeasure

	_ = m

	// START deleteMeasure
	if err := stats.DeleteMeasure(m); err != nil {
		log.Fatal(err)
	}
	// END deleteMeasure

	// START aggs
	distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
	countAgg := stats.CountAggregation{}
	sumAgg := stats.SumAggregation{}
	meanAgg := stats.MeanAggregation{}
	// END aggs

	_, _, _, _ = distAgg, countAgg, sumAgg, meanAgg

	// START windows
	interval := stats.Interval{
		Duration:  10 * time.Second,
		Intervals: 5,
	}

	cum := stats.Cumulative{}
	// END windows

	_, _ = interval, cum

	// START view
	view, err := stats.NewView(
		"my.org/video_size_distribution",
		"distribution of processed video size over time",
		nil,
		videoSize,
		distAgg,
		cum,
	)
	if err != nil {
		log.Fatalf("cannot create view: %v", err)
	}
	if err := stats.RegisterView(view); err != nil {
		log.Fatal(err)
	}
	// END view

	// START findView
	v := stats.FindView("my.org/video_size_distribution")
	if v == nil {
		log.Fatalln("view not found")
	}
	// END findView

	_ = v

	// START unregisterView
	if err = stats.UnregisterView(v); err != nil {
		log.Fatal(err)
	}
	// END unregisterView

	// START reportingPeriod
	stats.SetReportingPeriod(5 * time.Second)
	// END reportingPeriod

	// START record
	stats.Record(ctx, videoSize.M(102478))
	// END record

	// START subscribe
	if err := view.Subscribe(); err != nil {
		log.Fatal(err)
	}
	// END subscribe

	// START registerExporter
	// Register an exporter to be able to retrieve
	// the data from the subscribed views.
	stats.RegisterExporter(&exporter{})
	// END registerExporter
}

// START exporter

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

// END exporter
```

## Traces

### Starting and ending a span

[embedmd]:# (trace.go startend)
```startend
// Copyright 2017, OpenCensus Authors
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

package readme

import (
	"context"

	"go.opencensus.io/trace"
)

func traceExamples() {
	ctx := context.Background()

	// START startend
	ctx = trace.StartSpan(ctx, "your choice of name")
	defer trace.EndSpan(ctx)
	// END startend
}
```

More tracing examples are coming soon...

## Profiles

OpenCensus tags can be applied as profiler labels
for users who are on Go 1.9 and above.

[embedmd]:# (tags.go profiler)
```profiler
// Copyright 2017, OpenCensus Authors
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

package readme

import (
	"context"
	"log"

	"go.opencensus.io/tag"
)

func tagsExamples() {
	ctx := context.Background()

	// START stringKey
	// Get a key to represent user OS.
	key, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	// END stringKey
	_ = key

	// START tagMap
	osKey, err := tag.NewKey("my.org/keys/user-os")
	if err != nil {
		log.Fatal(err)
	}
	userIDKey, err := tag.NewKey("my.org/keys/user-id")
	if err != nil {
		log.Fatal(err)
	}

	tagMap, err := tag.NewMap(ctx,
		tag.Insert(osKey, "macOS-10.12.5"),
		tag.Upsert(userIDKey, "cde36753ed"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// END tagMap

	// START newContext
	ctx = tag.NewContext(ctx, tagMap)
	// END newContext

	// START replaceTagMap
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	// END replaceTagMap

	// START profiler
	tagMap, err = tag.NewMap(ctx,
		tag.Insert(key, "macOS-10.12.5"),
		tag.Upsert(key, "macOS-10.12.7"),
		tag.Upsert(userIDKey, "fff0989878"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx = tag.NewContext(ctx, tagMap)
	tag.Do(ctx, func(ctx context.Context) {
		// Do work.
		// When profiling is on, samples will be
		// recorded with the key/values from the tag map.
	})
	// END profiler
}
```

A screenshot of the CPU profile from the program above:

![CPU profile](https://i.imgur.com/jBKjlkw.png)


[travis-image]: https://travis-ci.org/census-instrumentation/opencensus-go.svg?branch=master
[travis-url]: https://travis-ci.org/census-instrumentation/opencensus-go
[appveyor-image]: https://ci.appveyor.com/api/projects/status/vgtt29ps1783ig38?svg=true
[appveyor-url]: https://ci.appveyor.com/project/opencensusgoteam/opencensus-go/branch/master
[godoc-image]: https://godoc.org/go.opencensus.io?status.svg
[godoc-url]: https://godoc.org/go.opencensus.io
[gitter-image]: https://badges.gitter.im/census-instrumentation/lobby.svg
[gitter-url]: https://gitter.im/census-instrumentation/lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge


[newtags-ex]: https://godoc.org/go.opencensus.io/tag#example-NewMap
[newtags-replace-ex]: https://godoc.org/go.opencensus.io/tag#example-NewMap--Replace

[exporter-prom]: https://godoc.org/go.opencensus.io/exporter/stats/prometheus
[exporter-sdstats]: https://godoc.org/go.opencensus.io/exporter/stats/stackdriver
[exporter-zipkin]: https://godoc.org/go.opencensus.io/exporter/trace/zipkin
[exporter-sdtrace]: https://godoc.org/go.opencensus.io/exporter/trace/stackdriver
