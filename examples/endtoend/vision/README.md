# vision-api

## About
This example shows how to use both the trace and stats/exporters to 
instrument your backend. The API sends information both to Prometheus
and to StackDriver monitoring sending spans of methods invoked.

The API server presented is one a company might use for
their app and its purpose is to give them information about their
backend's usage e.g:
* images processed
* methods used i.e. upload with body vs by URI
* bytes processsed altogether
* the performance and time spent within each service
* error rates

The information collected will help their engineering teams
figure out what services to optimize for, what metrics
their business development team will present, the popularity
of their service, general performance, error rates.

## Routes
Route|Purpose
---|---
/upload|Multipart uploads for content say from your filesystem
/metrics|Prometheus exporter route
/url|JSON based API invocation


## Sample response using JSON based route "/url"

```shell
$ curl -X POST http://localhost:8899/url --data '{"url": "https://avatars3.githubusercontent.com/u/4898263?s=460&v=4"}'
```
which gives
```JSON
{"faces":[{"bounding_poly":{"vertices":[{"x":111,"y":70},{"x":291,"y":70},{"x":291,"y":279},{"x":111,"y":279}]},"fd_bounding_poly":{"vertices":[{"x":125,"y":122},{"x":254,"y":122},{"x":254,"y":251},{"x":125,"y":251}]},"landmarks":[{"type":1,"position":{"x":170.74551,"y":163.39682,"z":-0.00009324717}},{"type":2,"position":{"x":226.9261,"y":162.27528,"z":-3.3436992}},{"type":3,"position":{"x":153.67024,"y":150.73282,"z":6.6217027}},{"type":4,"position":{"x":184.95543,"y":150.01039,"z":-11.03244}},{"type":5,"position":{"x":211.29453,"y":149.94893,"z":-12.643309}},{"type":6,"position":{"x":244.78865,"y":151.11148,"z":1.3010101}},{"type":7,"position":{"x":197.63452,"y":162.08948,"z":-13.04061}},{"type":8,"position":{"x":197.02707,"y":190.10608,"z":-29.71414}},{"type":9,"position":{"x":196.98128,"y":211.07048,"z":-18.037405}},{"type":10,"position":{"x":197.19337,"y":230.66444,"z":-14.682273}},{"type":11,"position":{"x":172.81384,"y":220.45654,"z":-1.3366407}},{"type":12,"position":{"x":225.52855,"y":219.70993,"z":-3.4198322}},{"type":13,"position":{"x":197.22342,"y":220.48212,"z":-14.274697}},{"type":14,"position":{"x":214.58203,"y":196.89085,"z":-9.772623}},{"type":15,"position":{"x":180.6429,"y":196.00703,"z":-7.762398}},{"type":16,"position":{"x":196.71512,"y":200.28355,"z":-17.826805}},{"type":17,"position":{"x":171.36452,"y":159.27315,"z":-3.6517794}},{"type":18,"position":{"x":181.57576,"y":164.14264,"z":-0.5285912}},{"type":19,"position":{"x":170.20926,"y":166.77805,"z":-0.6425587}},{"type":20,"position":{"x":160.3396,"y":163.68492,"z":5.7165737}},{"type":29,"position":{"x":170.58768,"y":163.04672,"z":-1.5383104}},{"type":21,"position":{"x":225.21132,"y":159.61885,"z":-6.943501}},{"type":22,"position":{"x":236.61172,"y":163.82162,"z":0.9396738}},{"type":23,"position":{"x":226.92223,"y":166.23566,"z":-3.9666982}},{"type":24,"position":{"x":214.53113,"y":163.99612,"z":-2.5623853}},{"type":30,"position":{"x":225.85002,"y":163.37703,"z":-5.023317}},{"type":25,"position":{"x":169.09799,"y":143.48747,"z":-5.7417264}},{"type":26,"position":{"x":227.47415,"y":143.87534,"z":-9.315583}},{"type":27,"position":{"x":140.72449,"y":190.75964,"z":68.05329}},{"type":28,"position":{"x":263.5416,"y":192.86166,"z":60.262157}},{"type":31,"position":{"x":197.83046,"y":149.26822,"z":-13.928979}},{"type":32,"position":{"x":197.46735,"y":254.6251,"z":-8.149228}},{"type":33,"position":{"x":145.47163,"y":225.18817,"z":44.309467}},{"type":34,"position":{"x":255.8797,"y":225.8844,"z":37.550243}}],"roll_angle":0.5147558,"pan_angle":-3.5062222,"tilt_angle":2.0259361,"detection_confidence":0.8725568,"landmarking_confidence":0.5131235,"joy_likelihood":4,"sorrow_likelihood":1,"anger_likelihood":1,"surprise_likelihood":1,"under_exposed_likelihood":1,"blurred_likelihood":1,"headwear_likelihood":1}],"labels":[{"mid":"/m/04yx4","description":"man","score":0.9350149},{"mid":"/m/01g317","description":"person","score":0.93318623},{"mid":"/m/05zppz","description":"male","score":0.8540857},{"mid":"/m/08t9c_","description":"grass","score":0.73753375},{"mid":"/m/07j7r","description":"tree","score":0.7349218},{"mid":"/m/06bm2","description":"recreation","score":0.643857},{"mid":"/m/02vzx9","description":"player","score":0.63084155},{"mid":"/m/01qkbx","description":"professional","score":0.5828397},{"mid":"/m/0ds99lh","description":"fun","score":0.55307543},{"mid":"/m/05s2s","description":"plant","score":0.5302758}]}
```

## Sample Prometheus output
```HTTP
# HELP opencensus_bytes_bucket_cum number of bytes ingested over time
# TYPE opencensus_bytes_bucket_cum histogram
opencensus_bytes_bucket_cum_bucket{le="0"} 0
opencensus_bytes_bucket_cum_bucket{le="1024"} 0
opencensus_bytes_bucket_cum_bucket{le="102400"} 1
opencensus_bytes_bucket_cum_bucket{le="1.048576e+06"} 1
opencensus_bytes_bucket_cum_bucket{le="1.048576e+07"} 0
opencensus_bytes_bucket_cum_bucket{le="1.048576e+08"} 0
opencensus_bytes_bucket_cum_bucket{le="1.073741824e+10"} 0
opencensus_bytes_bucket_cum_bucket{le="+Inf"} 2
opencensus_bytes_bucket_cum_sum 151606
opencensus_bytes_bucket_cum_count 2
# HELP opencensus_images_cum number of images uploaded over time
# TYPE opencensus_images_cum counter
opencensus_images_cum 2
```
