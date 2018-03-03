datadog
---------------------

`datadog` provides an `view.Exporter` for Datadog.

### Basic Usage

By default, the datadog exporter will connect to a locally running datadog
agent, `127.0.0.1:8125`.

```
exporter, _ := datadog.NewExporter()
view.RegisterExporter(exporter)

// define the view
myView, _ := view.New(
    "my.org/views/openconn",
    "open connections",
    nil,
    view.MeanAggregation{},
)

// begin exporting metrics
_ = myView.Subscribe()

// stop exporting metrics
myView.Unsubscribe()
```

