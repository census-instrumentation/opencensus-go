## Benchmarks

Benchmarks instrumenting an OpenCensus enabled gRPC backend vs one that
isn't. Once tests are run it generates some charts of:

* allocs/op vs QPS
* throughput/op vs QPS


### Running it
```shell
make all
```

and this will generate 2 HTML files that can then be inspected.
