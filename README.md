# OpenCensus Libraries for Go

[![Build Status][travis-image]][travis-url]
[![Windows Build Status][appveyor-image]][appveyor-url]
[![GoDoc][godoc-image]][godoc-url]
[![Gitter chat][gitter-image]][gitter-url]

OpenCensus Go is a Go implementation of OpenCensus, a toolkit for
collecting application performance and behavior monitoring data.
Currently it consists of three major components: tags, stats, and tracing.

This project is still at a very early stage of development and
a lot of the API calls are in the process of being changed and
might break your code in the future.

TODO: Add a link to the language independent OpenCensus
doc when it is available.

## Installation

```
$ go get -u go.opencensus.io/...
```

## Prerequisites

OpenCensus Go libraries require Go 1.8 or later.

## Tags

Tags represent propagated key values. They can propagated using context.Context
in the same process or can be encoded to be transmitted on wire and decoded back
to a tag.Map at the destination.

### Getting a key by a name

A key is defined by its name. To use a key a user needs to know its name and type.
Currently, only keys of type string are supported.
Other types will be supported in the future.

[embedmd]:# (tags.go stringKey)
```go
// Get a key to represent user OS.
key, err := tag.NewKey("my.org/keys/user-os")
if err != nil {
	log.Fatal(err)
}
```

### Creating a map of tags associated with keys

tag.Map is a map of tags. Package tags provide a builder to create tag maps.

[embedmd]:# (tags.go tagMap)
```go
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
```

### Propagating a tag map in a context

To propagate a tag map to downstream methods and downstream RPCs, add a tag map
to the current context. NewContext will return a copy of the current context,
and put the tag map into the returned one.
If there is already a tag map in the current context, it will be replaced.

[embedmd]:# (tags.go newContext)
```go
ctx = tag.NewContext(ctx, tagMap)
```

In order to update an existing tag map, get the tag map from the current context,
use NewMap and put the new tag map back to the context.

[embedmd]:# (tags.go replaceTagMap)
```go
tagMap, err = tag.NewMap(ctx,
	tag.Insert(key, "macOS-10.12.5"),
	tag.Upsert(key, "macOS-10.12.7"),
	tag.Upsert(userIDKey, "fff0989878"),
)
if err != nil {
	log.Fatal(err)
}
ctx = tag.NewContext(ctx, tagMap)
```


## Stats

### Creating, retrieving and deleting a measure

Create and load measures with units:

[embedmd]:# (stats.go measure)
```go
videoSize, err := stats.NewMeasureInt64("my.org/video_size", "processed video size", "MB")
if err != nil {
	log.Fatal(err)
}
```

Retrieve measure by name:

[embedmd]:# (stats.go findMeasure)
```go
m := stats.FindMeasure("my.org/video_size")
if m == nil {
	log.Fatalln("measure not found")
}
```

Delete measure (this can be useful when replacing a measure by
another measure with the same name):

[embedmd]:# (stats.go deleteMeasure)
```go
if err := stats.DeleteMeasure(m); err != nil {
	log.Fatal(err)
}
```

### Creating an aggregation

Currently 4 types of aggregations are supported. The CountAggregation is used to count
the number of times a sample was recorded. The DistributionAggregation is used to
provide a histogram of the values of the samples. The SumAggregation is used to
sum up all sample values. The MeanAggregation is used to calculate the mean of
sample values.

[embedmd]:# (stats.go aggs)
```go
distAgg := stats.DistributionAggregation([]float64{0, 1 << 32, 2 << 32, 3 << 32})
countAgg := stats.CountAggregation{}
sumAgg := stats.SumAggregation{}
meanAgg := stats.MeanAggregation{}
```

### Create an aggregation window

Currently only two types of aggregation windows are supported. The Cumulative
is used to continuously aggregate the data received.
The Interval window is used to aggregate the data received over the last specified time interval.
Currently all aggregation types are compatible with all aggregation windows.
Later we might provide aggregation types that are incompatible with some windows.

[embedmd]:# (stats.go windows)
```go
interval := stats.Interval{
	Duration:  10 * time.Second,
	Intervals: 5,
}

cum := stats.Cumulative{}
```

### Creating, registering and unregistering a view

Create and register a view:

[embedmd]:# (stats.go view)
```go
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
```

Find view by name:

[embedmd]:# (stats.go findView)
```go
v := stats.FindView("my.org/video_size_distribution")
if v == nil {
	log.Fatalln("view not found")
}
```

Unregister view:

[embedmd]:# (stats.go unregisterView)
```go
if stats.UnregisterView(v); err != nil {
	log.Fatal(err)
}
```

Configure the default interval between reports of collected data.
This is a system wide interval and impacts all views. The default
interval duration is 10 seconds. Trying to set an interval with
a duration less than a certain minimum (maybe 1s) should have no effect.

[embedmd]:# (stats.go reportingPeriod)
```go
stats.SetReportingPeriod(5 * time.Second)
```

### Recording measurements

Recording usage can only be performed against already registered measure
and their registered views. Measurements are implicitly tagged with the
tags in the context:

[embedmd]:# (stats.go record)
```go
stats.Record(ctx, videoSize.M(102478))
```

### Retrieving collected data for a view

Users need to subscribe to a view in order to retrieve collected data.

[embedmd]:# (stats.go subscribe)
```go
if view.Subscribe(); err != nil {
	log.Fatal(err)
}
```

Subscribed views' data will be exported via the registered exporters.

[embedmd]:# (stats.go registerExporter)
```go
// Register an exporter to be able to retrieve
// the data from the subscribed views.
stats.RegisterExporter(&exporter{})
```

An example logger exporter is below:

[embedmd]:# (stats.go exporter)
```go

type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
	log.Println(vd)
}

```

## Tracing

### Starting and ending a span

[embedmd]:# (trace.go startend)
```go
ctx = trace.StartSpan(ctx, "your choice of name")
defer trace.EndSpan(ctx)
```


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
