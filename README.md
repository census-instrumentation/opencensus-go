# Go stats collection libraries

[![Build Status][travis-image]][travis-url] [![GoDoc][godoc-image]][godoc-url]

This is still at a very early stage of development and a lot of the API calls
are in the process of being changed and might break your code in hte future.

[travis-image]: https://travis-ci.org/census-instrumentation/opencensus-go.svg?branch=master
[travis-url]: https://travis-ci.org/census-instrumentation/opencensus-go
[godoc-image]: https://godoc.org/github.com/census-instrumentation/opencensus-go?status.svg
[godoc-url]: https://godoc.org/github.com/census-instrumentation/opencensus-go

# Go stats core library
The Go implementation of opencensus.
TODO: add a link to the language independent opencensus doc when it is available.

## Installation
To install this package, you need to install Go and setup your Go workspace on your computer. The simplest way to install the library is to run:

$ go get -u github.com/census-instrumentation/opencensus-go

## Prerequisites
This requires Go 1.8 or later as it uses the convenience function sort.Slice(...) introduced in Go 1.8.
It also uses "context".Context introduced in Go 1.7 heavily. Therefore, in order to integrate with any library still using "golang.org/x/net/context".Context, Go 1.9 or higher is required to take advantage of type aliasing. 

## Tags

### To create/retrieve a key
A key is defined by its name. To use a key a user needs to know its name and type (currently only keys of type string are supported. Later support for keys of type int64 and bool will be supported). Calling CreateKeyString(...) multiple times with the same name returns the same key.
To create/retrieve a key the user calls:

Create/retrieve key:

    if err, key1 := tags.CreateKeyString("keyNameID1"); err != nil {
        // handle error
    }
    ...

    if err, key2 := tags.CreateKeyString("keyNameID2"); err != nil {
        // handle error
    }
    ...

### Create a set of tags associated with keys
To create a new tag set from scratch using changes:

    tsb := NewTagSetBuilder(nil)
    tsb.InsertString(key1, "foo value")
    tsb.UpdateString(key1, "foo value2")
    tsb.UpsertString(key2, "bar value")
    tagsSet := tsb.Build()

    // A shorter way of achieving the same is:
    tagsSet := NewTagSetBuilder().InsertString(key1, "foo value").
                                  UpdateString(key1, "foo value2").
                                  UpsertString(key2, "bar value").
                                  Build()



To create a new tagsSet from an existing tag set oldTagSet:
    oldTagSet := ...
    tsb := NewTagSetBuilder(oldTagSet)
    tsb.InsertString(key1, "foo value")
    tsb.UpdateString(key1, "foo value2")
    tsb.UpsertString(key2, "bar value")
    newTagSet := tsb.Build()

    // A shorter way of achieving the same is:
    oldTagSet := ...
	newTagSet := NewTagSetBuilder(oldTagSet).InsertString(key1, "foo value").
		UpdateString(key1, "foo value2").
		UpsertString(key2, "bar value").
		Build()    

### Add new tagSet to a context / Modify tagSet in a context 
Add tags to a context for propagation to downstream methods and downstream rpcs:
To create a new context with the tags. This will create a new context where all the existing tags in the current context are deleted and replaced with the tags passed as argument.
    
    newTagSet  := ...
    ctx2 := tags.NewContext(ctx, newTagSet)

Create a new context keeping the old tag set and adding new tags to it, removing specific tags from it, or modifying the values fo some tags. This is just a matter of getting the oldTagSet from the context, apply changes to it, then create a new  context with the newTagSet.
    
    oldTagSet := tags.FromContext(ctx)
    newTagSet := NewTagSetBuilder(oldTagSet).InsertString(key1, "foo value").
        UpdateString(key1, "foo value2").
        UpsertString(key2, "bar value").
        Build()

    ctx2 := tags.NewContext(ctx, newTagSet)

## Stats

### To create/retrieve/delete a measure a.k.a resource
Create/load measures units:

    // returns a *MeasureFloat64
    mf, err := stats.NewMeasureFloat64("/my/float64/measureName", "some measure")
    if err != nil {
        // handle error
    }
    mi, err := stats.NewMeasureInt64("/my/otherName", "some other measure")
    if err != nil {
        // handle error
    }    
    ...

Retrieve measure by name:

    mf, err := stats.GetMeasureByName("/my/float64/measureName")
	if err != nil {
        // handle error
    }
    mi, err := stats.GetMeasureByName("/my/otherName")
    if err != nil {
        // handle error
    }
    ...

Delete measure (this can be useful when replacing a measure by another measure with the same name):

	if err := stats.DeleteMeasure(mf); err != nil {
        // handle error
    }
    if err := stats.DeleteMeasure(mi); err != nil {
        // handle error
    }
    ...

### To create an aggregation type
Currently only 2 types of aggregations are supported. The AggregationCount is used to count the number of times a sample was recorded. The AggregationDistribution is used to provide a histogram of the values of the samples.

    histogramBounds := []float64 { -10, 0, 10, 20}
    agg1 := stats.NewAggregationDistribution(histogramBounds)
    agg2 := stats.NewAggregationCount()


### To create an aggregation window
Currently only 3 types of aggregation windows are supported. The WindowCumulative is used to continuously aggregate the data received. The WindowSlidingTime to aggregate the data received over the last specified time interval. The NewWindowSlidingCount to aggregate the data received over the last specified sample count.
Currently all aggregation types are compatible with all aggregation windows. Later we might provide aggregation types that are incompatible with some windows.

    duration := 10 * time.Second
    precisionIntervals := 5
    wnd1 := stats.NewWindowSlidingTime(duration, precisionIntervals)

    lastNSamples := 100
    precisionSubsets := 10
    wnd2 := stats.NewWindowSlidingCount(lastNSamples, precisionSubsets)
    
    wn3 := stats.NewWindowCumulative()


### To creater/register a view
Create a view:

    myView1 = stats.NewView("/my/int64/viewName", "some description", []Keys{key1, key2}, mf, agg1, wnd1)
    myView2 := stats.NewView("/my/float64/viewName", "some other description", []Keys{key1}, mi, agg2, wnd3)

Register view:

    if err := stats.RegisterView(myView1); err != nil {
      // handle error
    }
    if err := stats.RegisterView(myView2); err != nil {
      // handle error
    }

Retrieve view by name:
    
    myView1, err := stats.GetViewByName("/my/int64/viewName")
	if err != nil {
        // handle error
    }
    myView2, err := stats.GetViewByName("/my/float64/viewName")
    if err != nil {
        // handle error
    }
    ...


Unregister view:

    if err := stats.UnregisterView(myView1); err != nil {
        // handle error
    }
    if err := stats.UnregisterView(myView2); err != nil {
        // handle error
    }
    ... 

### To subscribe/unsubscribe to a view's collected data
Once a subscriber subscribes to a view, its collected date is reported at a regular interval. This interval is configured system wide.

Subscribe to a view:

    c1 := make(c chan *ViewData)
    if err := stats.SubscribeToView(myView1, c1); err != nil {
        // handle error
    }

    c2 := make(c chan *ViewData)
    if err := stats.SubscribeToView(myView2, c2); err != nil {
        // handle error
    }
    ... 

Unsubscribe from a view:

    if err := stats.UnsubscribeFromView(myView1, c1); err != nil {
        // handle error
    }
    if err := stats.UnsubscribeFromView(myView2, c2); err != nil {
        // handle error
    }    
    ...

Configure/modify the default interval between reports of collected data. This is a system wide interval and impacts all views. The default interval duration is 10 seconds. Trying to set an interval with a duration less than a certain minimum (maybe 1s) should have no effect.

    d := 20 * time.Second
    stats.SetReportingPeriod(d)

### To collect data for on-demand retrieveal
Even if a view is registered, if it has no subscriber no data for it is collected. In order to retrieve data on-demand for view, either the view needs to have at least 1 subscriber or the libray needs to be instructed explicitly to collect collect data for the desired view.

    // explicitly instruct the library to collect the view data for an on-demand retrieval.
    if err := stats.ForceCollection(myView1); err != nil {
        // handle error
    }
    ...

    // Retrieve data. The returned rows are a []stats.Row. Each row is a struct:
    //  {
	//      Tags             []tags.Tag
	//      AggregationValue AggregationValue
    //  }
    rows, err := stats.RetrieveData(v View)
    if err != nil {
        // handle error
    }
    ...

    // explicitly instruct the library to stop collecting the view data for the
    // on-demand retrieval. Has no impact on subscriptions, and if the view
    // still has subscribers, the data for the view will still keep being
    // collected.
    if err := stats.StopForcedCollection(myView1); err != nil {
        // handle error
    }
    ...

## Recording usage/measurements
Recording usage can only be performed against already registered measure and and their registered views. Measurements are implicitly tagged with the tags in the context:
    
    // mi is a *MeasureInt64 and v is an int64 .
    stats.RecordInt64(ctx, mi, v)

    // mf is a *RecordFloat64 and v is an float64 .
    stats.RecordFloat64(ctx, mf, v)

    // multiple measurements can be performed at once.
    stats.Record(ctx, mi.Is(4), mf.Is(10.5))
