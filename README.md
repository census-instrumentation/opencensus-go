# Go stats and tracing collection libraries
This is still at a very early stage of development and a lot of the API calls
are in the process of being changed and will certainly break your code.

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

### To create/retrive a key
To use a key. It needs to be created/retrieved. If a key is already created the KeysManager will just return it. Calling CreateKey(...) multiple times with the same name for the same type will return the same key.

Register/retrieve key:

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

## Recording usage against the registered views using the registered/created keys

### Create tags
To create a new tag set from scratch.

    builder := stats.NewTagsBuilder()
    builder.StartFromEmpty()
    builder.AddTag(key1.CreateTag("tagValue"))  // key1 is a KeyString so its value is a string
    builder.UpsertTag(key2.CreateTag(10))       // key2 is a KeyInt64 so its value is an int64
    tags := builder.Build()

To create a set of tags from an existing tag set:

    builder.StartFromTags(existingTags)
    builder.AddTag(key1.CreateTag("tagValue"))   // key1 is a KeyString so its value is a string
    builder.UpdateTag(key2.CreateTag(10))        // key2 is a KeyInt64 so its value is an int64
    tags := builder.Build()

### Add tags 
Add tags to a context for propagation to downstream methods and downstream rpcs:
Create a new context with the tags.
    
    ctx := stats.ContextWithNewTags(ctx, tags)       // this will create a new context where the existing tags in the current context are replaced with the tags passed as argument.

Create a new context derived from an existing context using a changes
    
    var changes []tags.Change
    changes = append(changes, tags.Change(key1, "someValue", tagOpInsert))
    changes = append(changes, ...)
    ctx := stats.ContextWithChanges(ctx, changes)    // this will create a new context where the changed passed as argument as applied to the existing tags in the current context.

Extract tags from a context:
    
    tags := stats.FromContext()

## Record measurements against Measures with a set of tags:

Record measurement against the measure RPCclientErrorCount with the tags:   
    
    stats.RecordMeasurement(tags, RPCclientErrorCount, 1)

Record measurement against the measure RPCclientErrorCount with the tags embeded in the context. This is just a "sugar" helper function that extracts the tags from the context and then calls stats.RecordMeasurement(tags,...)
    
    stats.RecordMeasurement(ctx, RPCclientErrorCount, 1)
