# OpenCensus Libraries for Go

[![Build Status][travis-image]][travis-url]
[![Windows Build Status][appveyor-image]][appveyor-url]
[![GoDoc][godoc-image]][godoc-url]
[![Gitter chat][gitter-image]][gitter-url]

OpenCensus Go is a Go implementation of OpenCensus, a toolkit for
collecting application performance and behavior monitoring data.
Currently it consists of three major APIs: tags, stats, and tracing.

This project is still at a very early stage of development and
a lot of the API calls are in the process of being changed and
might break your code in the future.


TODO: Add a link to the language independent OpenCensus doc when it is available.

## Installation
To install this package, you need to install Go and setup your Go workspace on your computer. The simplest way to install the library is to run:

```
$ go get -u github.com/census-instrumentation/opencensus-go/...
```

## Prerequisites
OpenCensus libraries require Go 1.8 or later as it uses the convenience function sort.Slice(...) introduced in Go 1.8.

## Tags API

Tags represent propagated key values. They can propagated using context.Context
in the same process or can be encoded to be transmitted on wire and decoded back
to a tag.Map at the destination.

### Getting a key by a name

A key is defined by its name. To use a key a user needs to know its name and type.
Currently, only keys of type string are supported.
Other types will be supported in the future.

See the [NewStringKey][newstringkey-ex] example.

### Creating a map of tags associated with keys

tag.Map is a map of tags. Package tags provide a builder to create tag maps.

See the [NewMap][newtags-ex] example.

### Propagating a tag map in a context
To propagate a tag map to downstream methods and downstream RPCs, add a tag map
to the current context. NewContext will return a copy of the current context,
and put the tag map into the returned one.
If there is already a tag map in the current context, it will be replaced.

```go
tagMap  := ...
ctx = tag.NewContext(ctx, tagMap)
```

In order to update an existing tag map, get the tag map from the current context,
use NewMap and put the new tag map back to the context.

See the [NewMap (Replace)][newtags-replace-ex] example.


## Stats API

### Creating, retrieving and deleting a measure

Create and load measures with units:

```go
// returns a *MeasureFloat64
mf, err := stats.NewMeasureFloat64("/my/float64/measureName", "some measure", "MBy")
if err != nil {
    // handle error
}
mi, err := stats.NewMeasureInt64("/my/otherName", "some other measure", "1")
if err != nil {
    // handle error
}
```

Retrieve measure by name:

```go
mf, err := stats.FindMeasure("/my/float64/measureName")
if err != nil {
    // handle error
}
mi, err := stats.FindMeasure("/my/otherName")
if err != nil {
    // handle error
}
```

Delete measure (this can be useful when replacing a measure by another measure with the same name):

```go
if err := stats.DeleteMeasure(mf); err != nil {
    // handle error
}
if err := stats.DeleteMeasure(mi); err != nil {
    // handle error
}
```

### Creating an aggregation
Currently only 2 types of aggregations are supported. The CountAggregation is used to count
the number of times a sample was recorded. The DistributionAggregation is used to
provide a histogram of the values of the samples.

```go
agg1 := stats.DistributionAggregation([]float64 {-10, 0, 10, 20})
agg2 := stats.CountAggregation{}
```

### Create an aggregation window

Currently only 3 types of aggregation windows are supported. The CumulativeWindow
is used to continuously aggregate the data received.
The SlidingTimeWindow to aggregate the data received over the last specified time interval.
The SlidingCountWindow to aggregate the data received over the last specified sample count.
Currently all aggregation types are compatible with all aggregation windows.
Later we might provide aggregation types that are incompatible with some windows.

```go
wnd1 := stats.SlidingTimeWindow{
    Duration:  10 * time.Second,
    Intervals: 5,
}

wnd2 := stats.SlidingCountWindow{
    N:       100,
    Subsets: 10,
}

wn3 := stats.CumulativeWindow{}
```

### Creating, registering and unregistering a view

TODO: define "view" (link to the spec).

Create a view:

```go
myView1 := stats.NewView("/my/int64/viewName", "some description", []tag.Key{key1, key2}, mf, agg1, wnd1)
myView2 := stats.NewView("/my/float64/viewName", "some other description", []tag.Key{key1}, mi, agg2, wnd3)
```

Register view:

```go
if err := stats.RegisterView(myView1); err != nil {
  // handle error
}
if err := stats.RegisterView(myView2); err != nil {
  // handle error
}
```

Retrieve view by name:

```go
myView1, err := stats.FindView("/my/int64/viewName")
if err != nil {
    // handle error
}
myView2, err := stats.FindView("/my/float64/viewName")
if err != nil {
    // handle error
}
```

Unregister view:

```go
if err := myView1.Unregister(); err != nil {
    // handle error
}
if err := myView2.Unregister(); err != nil {
    // handle error
}
```

Configure/modify the default interval between reports of collected data. This is a system wide interval and impacts all views. The default interval duration is 10 seconds. Trying to set an interval with a duration less than a certain minimum (maybe 1s) should have no effect.

```go
stats.SetReportingPeriod(5 * time.Second)
```


### Recording measurements
Recording usage can only be performed against already registered measure
and their registered views. Measurements are implicitly tagged with the
tags in the context:

```go
stats.Record(ctx, mi.M(4), mf.M(10.5))
```

### Retrieving collected data for a view

Users need to subscribe to a view in order to retrieve collected data.

```go
if err := view.Subscribe(); err != nil {
    // handle error
}
```

Subscribed views' data will be exported via the registered exporters.

```go
// Register an exporter to be able to retrieve
// the data from the subscribed views.
stats.RegisterExporter(&exporter{})
```

An example logger exporter is below:

``` go
type exporter struct{}

func (e *exporter) Export(vd *stats.ViewData) {
    log.Println(vd)
}
```

### Force collecting data on demand
Even if a view is registered, if it has no subscriber no data for it is collected. In order to retrieve data on-demand for view, either the view needs to have at least one subscriber or the library needs to be instructed explicitly to collect collect data for the desired view.

```go
// To explicitly instruct the library to collect the view data for an on-demand
// retrieval, force collect. When done, stop force collection.
if err := view.ForceCollect(); err != nil {
    // handle error
}

// Use RetrieveData to pull collected data synchronously from the library. This
// assumes that a subscription to the view exists or force collection is enabled.
rows, err := view.RetrieveData()
if err != nil {
    // handle error
}
for _, r := range rows {
    // process a single row of type *stats.Row
}

// To explicitly instruct the library to stop collecting the view data for the
// on-demand retrieval, StopForceCollection should be used. This call has no
// impact on the view's subscription status.
if err := view.StopForceCollection(); err != nil {
    // handle error
}
```

## Tracing API
TODO: update the doc once tracing API is ready.


[travis-image]: https://travis-ci.org/census-instrumentation/opencensus-go.svg?branch=master
[travis-url]: https://travis-ci.org/census-instrumentation/opencensus-go
[appveyor-image]: https://ci.appveyor.com/api/projects/status/vgtt29ps1783ig38?svg=true
[appveyor-url]: https://ci.appveyor.com/project/opencensusgoteam/opencensus-go/branch/master
[godoc-image]: https://godoc.org/github.com/census-instrumentation/opencensus-go?status.svg
[godoc-url]: https://godoc.org/github.com/census-instrumentation/opencensus-go
[gitter-image]: https://badges.gitter.im/census-instrumentation/lobby.svg
[gitter-url]: https://gitter.im/census-instrumentation/lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge


[newstringkey-ex]: https://godoc.org/github.com/census-instrumentation/opencensus-go/tag#example-NewStringKey
[newtags-ex]: https://godoc.org/github.com/census-instrumentation/opencensus-go/tag#example-NewMap
[newtags-replace-ex]: https://godoc.org/github.com/census-instrumentation/opencensus-go/tag#example-NewMap--Replace
