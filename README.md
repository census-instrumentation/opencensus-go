# Go stats and tracing collection libraries

[![Build
Status](https://travis-ci.org/census-instrumentation/instrumentation-go.svg?branch=master)](https://travis-ci.org/census-instrumentation/instrumentation-go)

This is still at a very early stage of development and a lot of the API calls
are in the process of being changed and will certainly break your code.

# Go stats core library

## Keys and tagging

### To create/retrieve a key
A key is defined by its name. To use a key a user needs to know its name and type (currently only keys of type string are supported. Later support for keys of type int64 and bool will be supported).
To create/retrieve a key the user calls a keys manager appropriate function. Keys are safe to be reused and called from multiple go routines.

Create/retrieve key:

    key1 := tagging.DefaultKeyManager().CreateKeyString("keyNameID1")
    key2 := tagging.DefaultKeyManager().CreateKeyInt64("keyNameID2")
    ...

### Create a set of tags associated with keys
To create a new tag set from scratch using changes:

    change1 := key1.CreateChange("tagValue", TagOpUpsert)   // key1 is a KeyString so its value is a string
    change2 := key2.CreateChange(10, TagOpInsert)           // key2 is a KeyInt64 so its value is an int64
    builder := stats.NewTagsBuilder()
    builder.StartFromEmpty()
    builder.Apply(c1,c2)
    tags := builder.Build()

To create a new tag set from scratch using direct values (fastest):

    builder := stats.NewTagsBuilder()
    builder.StartFromEmpty()
    builder.UpsertString(key1, "tagValue")  // key1 is a KeyString so its value is a string
    builder.InsertInt64(key2, 10)           // key2 is a KeyInt64 so its value is an int64
    tags := builder.Build()

### Add tags to a context 
Add tags to a context for propagation to downstream methods and downstream rpcs:
To create a new context with the tags. This will create a new context where all the existing tags in the current context are deleted and replaced with the tags passed as argument.    
    
    ctx2 := stats.ContextWithNewTags(ctx, tags)

Create a new context derived from an existing context using a set of changes. This will create a new context where the changed passed as argument as applied to the existing tags in the current context.

    var changes []tags.Change
    changes = append(changes, tags.Change(key1, "someValue", tagOpInsert))
    changes = append(changes, ...)
    ctx2 := stats.ContextWithChanges(ctx, changes)

Extract tagsSet from a context:
    
    tags := stats.FromContext()

## Registering views and retrieving their collected data.

### To register a measure a.k.a resource
Create/load measures units:

    bytes := &stats.MeasurementUnit{
		Power10:    1,
	    Numerators: []stats.BasicUnit{stats.BytesUnit},
	}
	count := &stats.MeasurementUnit{
	    Power10:    1,
	    Numerators: []stats.BasicUnit{stats.ScalarUnit},
	}
    ...

Create/load measures definitions:

    RPCclientErrorCount   := stats.NewMeasureDescFloat64("/rpc/client/error_count", "RPC Errors", count)
    RPCclientRequestBytes := stats.NewMeasureDescFloat64("/rpc/client/request_bytes", "Request bytes", bytes)
    ...

Register measures:

	stats.RegisterMeasureDesc(RPCclientErrorCount)
    stats.RegisterMeasureDesc(RPCclientRequestBytes)
    ...

### To create/retrieve a key

    key1 := tagging.DefaultKeyManager().CreateKeyString("keyNameID1")
    key2 := tagging.DefaultKeyManager().CreateKeyInt64("keyNameID2")
    ...

### To register/unregister a view
Create views definition:

    RPCclientErrorCountDist = stats.NewDistributionViewDesc("/rpc/client/error_count/distribution_cumulative", "RPC Errors", "/rpc/client/error_count", []Keys{key1, key2}, ...)
    RPCclientRequestBytesDist = stats.NewDistributionViewDesc("/rpc/client/request_bytes/distribution_cumulative", "Request bytes", "/rpc/client/request_bytes", []TagKeys{key1, key2}, ...)
    ...

Register view definition:

    stats.RegisterViewDesc(RPCclientErrorCountDist)
    stats.RegisterViewDesc(RPCclientRequestBytesDist)  
    ... 

UnRegister view definition:

    stats.UnRegisterViewDesc(RPCclientErrorCountDist)    // Fails if subscribers exist
    stats.UnRegisterViewDesc(RPCclientRequestBytesDist)  // Fails if subscribers exist
    ... 

### To subscribe/unsubscribe to a view's collected data
Once a subscriber subscribes to a view, its collected date is reported at a regular interval. This interval is configured system wide.

Create subscription:

    subscription1 := &stats.SingleSubscription{
        ViewName: "/some/view/name"
        respChannel: make(chan []*stats.View)
    }
    ...   

Subscribe to a view:

    stats.Subscribe(subscription1)
    ...    

Unubscribe from a view:

    stats.Unsubscribe(subscription1)                    // Fails if not subscribers exist
    ...

### To retrieve the snapshot of a view's collected data on-demand

    viewResults := stats.Retrieve("/some/view/name")    // Fails if view is not registered

## Recording usage/measurements against the registered views using a set of tags / context:

Record measurement against the measure RPCclientErrorCount with the tags:   
    
    stats.RecordMeasurement(tags, RPCclientErrorCount, 1)

Record measurement against the measure RPCclientErrorCount with the tags embeded in the context. This is just a "sugar" helper function that extracts the tags from the context and then calls stats.RecordMeasurement(tags,...)
    
    stats.RecordMeasurement(ctx, RPCclientErrorCount, 1)

# Go tracing core library

TODO