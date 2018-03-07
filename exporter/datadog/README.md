datadog
---------------------

`datadog` provides an `view.Exporter` for Datadog.

### Basic Usage

By default, the datadog exporter will connect to a locally running datadog
agent, `127.0.0.1:8125`.

```
exporter, _ := datadog.NewExporter()
view.RegisterExporter(exporter)

// define the reporting measure
m, _ := stats.Int64("my.org/measure/openconns", "open connections", "")

// define the view
myView, _ := &view.View(
    Name:        "my.org/views/openconn",
    Description: "open connections",
    TagKeys:     nil,
    Measure:     m,
    Aggregation: view.MeanAggregation{},
)

// begin exporting metrics
_ = view.Subscribe(myView)
```

