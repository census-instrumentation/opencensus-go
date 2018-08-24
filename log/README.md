log
--------

`log` package supports structured logging that is both tag and trace aware.


#### QuickStart

The simplest way to get started is to register a logger and get started.

```go
var exporter log.Exporter = ...
log.RegisterExporter(exporter)
log.Info(ctx, "hello world", log.String("userID", "abc"), log.Int("count", 123))
```

#### Customizing

The behavior of the `log` package can be customized via `ApplyConfig`.

```go
var exporter log.Exporter = ...
userID, _ := tag.NewKey("userID")
log.RegisterExporter(exporter)

// customize behavior of log package here
log.ApplyConfig(log.Config{
	LogLevel: log.DebugLevel,                          // change log level to debug
	TimeFunc: func() time.Time { return time.Now() },  // customize how now is generated
	Fields: []log.Field{log.String("global", "field")} // global fields to be added to all logs
	Tags: []tag.Key{userID},                           // extract tags from context when present
})

log.Info(ctx, "hello world")

```
