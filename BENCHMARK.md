# imgproxy vs alternatives benchmark

## Setup

- c5.xlarge AWS instance: 4 CPUs, 8 GB RAM
- Ubuntu 18.04
- Go 1.12
- Python 2.7
- Vips 8.7.4

All the tools were launched with their default settings except the following (where applicable):

- Concurrency was set to 8;
- Filesystem image source was configured. Where not applicable, local nginx was used as the image source.

## Benchmarking

We used [Apache HTTP server benchmarking tool](https://httpd.apache.org/docs/2.4/programs/ab.html) (a.k.a `ab`):

```bash
ab -n 1000 -c 4 $url
```

The source image is a [photo of Wat Arun](https://upload.wikimedia.org/wikipedia/commons/b/b7/The_sculptures_of_two_mythical_giant_demons%2C_Thotsakan_and_Sahatsadecha%2C_guarding_the_eastern_gate_of_the_main_chapel_of_Wat_Arun%2C_Bangkok.jpg) (JPEG, 7360x4912, 29MB).

The tools were requested to resize it to fit 500x500.

## Tested tools

- [imgproxy](https://github.com/imgproxy/imgproxy) itself, v2.2.13

   URL: `/unsigned/rs:fit:500:0/plain/local:///wat-arun.jpg`

- [thumbor](https://github.com/thumbor/thumbor), v6.7.0

   URL: `/unsafe/500x0/wat-arun.jpg`

- [imaginary](https://github.com/h2non/imaginary), master branch ([519c5ff](https://github.com/h2non/imaginary/tree/519c5ffc2c0b1bbae1b100a24acb2241474f11bd))

   URL: `/fit?width=500&height=500&file=wat-arun.jpg`

- [Pilbox](https://github.com/agschwender/pilbox), v1.3.4

   URL: `/?url=http%3A%2F%2Fimages.dev.com%2Fwat-arun.jpg&w=500&h=500&mode=clip`

- [picfit](https://github.com/thoas/picfit), master branch ([fff7d2e](https://github.com/thoas/picfit/tree/fff7d2e83d23b1c716fed484bb5d0775a49c9a71))

   URL: `/display/resize/500x0/wat-arun.jpg`

- [imageproxy](https://github.com/willnorris/imageproxy), master branch ([d4246a0](https://github.com/willnorris/imageproxy/tree/d4246a08fdec341ddf01b09e74e56ee03f9929d0))

   URL: `/500x/http://images.dev.com/wat-arun.jpg`

## Results

| Tool          | Time taken for tests<br>(sec) | Requests per second | Time per request<br>(sec, mean) | Memory peak usage<br>(MB) | Result file size<br>(KB) |
| :------------ | -------------------------: | ------------------: | ---------------------------: | ---------------------: | --------------------: |
| imgproxy      | 103.405                    | 9.67                | 413.618                      | 194                    | 43.51                 |
| thumbor       | 160.505                    | 6.23                | 642.021                      | 461                    | 45.10                 |
| imaginary     | 104.873                    | 9.54                | 419.494                      | 562                    | 92.93                 |
| Pilbox        | 179.482                    | 5.57                | 717.927                      | 1060                   | 95.64                 |
| picfit        | 1220.412                   | 0.82                | 4881.646                     | 1934                   | 98.67                 |
| imageproxy    | 1209.361                   | 0.83                | 4837.443                     | 2392                   | 98.74                 |

## Detailed results

### imgproxy

```
Concurrency Level:      4
Time taken for tests:   103.405 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      44855000 bytes
HTML transferred:       44559000 bytes
Requests per second:    9.67 [#/sec] (mean)
Time per request:       413.618 [ms] (mean)
Time per request:       103.405 [ms] (mean, across all concurrent requests)
Transfer rate:          423.61 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:   323  413   5.6    413     446
Waiting:      323  413   5.6    412     446
Total:        323  413   5.6    413     446

Percentage of the requests served within a certain time (ms)
  50%    413
  66%    414
  75%    416
  80%    417
  90%    419
  95%    422
  98%    426
  99%    429
 100%    446 (longest request)

Memory peak usage: 194 [MB]
Result file size:  43.51 [KB]
```

### thumbor

```
Concurrency Level:      4
Time taken for tests:   160.505 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      46435000 bytes
HTML transferred:       46180000 bytes
Requests per second:    6.23 [#/sec] (mean)
Time per request:       642.021 [ms] (mean)
Time per request:       160.505 [ms] (mean, across all concurrent requests)
Transfer rate:          282.52 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:   586  642  12.3    641     886
Waiting:      586  641  12.3    641     886
Total:        586  642  12.3    641     886

Percentage of the requests served within a certain time (ms)
  50%    641
  66%    644
  75%    646
  80%    647
  90%    650
  95%    652
  98%    657
  99%    658
 100%    886 (longest request)

Memory peak usage: 461 [MB]
Result file size:  45.10 [KB]
```

### imaginary

```
Concurrency Level:      4
Time taken for tests:   104.873 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      95389568 bytes
HTML transferred:       95166000 bytes
Requests per second:    9.54 [#/sec] (mean)
Time per request:       419.494 [ms] (mean)
Time per request:       104.873 [ms] (mean, across all concurrent requests)
Transfer rate:          888.25 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:   310  419   6.4    418     459
Waiting:      310  418   6.4    418     459
Total:        310  419   6.4    418     460

Percentage of the requests served within a certain time (ms)
  50%    418
  66%    420
  75%    421
  80%    422
  90%    425
  95%    428
  98%    432
  99%    437
 100%    460 (longest request)

Memory peak usage: 562 [MB]
Result file size:  92.93 [KB]
```

### Pilbox

```
Concurrency Level:      4
Time taken for tests:   179.482 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      98164000 bytes
HTML transferred:       97934000 bytes
Requests per second:    5.57 [#/sec] (mean)
Time per request:       717.927 [ms] (mean)
Time per request:       179.482 [ms] (mean, across all concurrent requests)
Transfer rate:          534.11 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       1
Processing:   456  717 232.5    658    1851
Waiting:      456  717 232.5    657    1851
Total:        457  717 232.5    658    1851

Percentage of the requests served within a certain time (ms)
  50%    658
  66%    663
  75%    668
  80%    674
  90%   1176
  95%   1278
  98%   1340
  99%   1618
 100%   1851 (longest request)

Memory peak usage: 1060 [MB]
Result file size:  95.64 [KB]
```

### picfit

```
Concurrency Level:      4
Time taken for tests:   1220.412 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      101206000 bytes
HTML transferred:       101038000 bytes
Requests per second:    0.82 [#/sec] (mean)
Time per request:       4881.646 [ms] (mean)
Time per request:       1220.412 [ms] (mean, across all concurrent requests)
Transfer rate:          80.98 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:  2322 4875 118.6   4880    5265
Waiting:     2322 4874 118.6   4879    5265
Total:       2322 4875 118.6   4880    5265

Percentage of the requests served within a certain time (ms)
  50%   4880
  66%   4906
  75%   4924
  80%   4936
  90%   4978
  95%   5009
  98%   5078
  99%   5124
 100%   5265 (longest request)

Memory peak usage: 1934 [MB]
Result file size:  98.67 [KB]
```

### imageproxy

```
Concurrency Level:      4
Time taken for tests:   1209.361 seconds
Complete requests:      1000
Failed requests:        0
Total transferred:      101318000 bytes
HTML transferred:       101108000 bytes
Requests per second:    0.83 [#/sec] (mean)
Time per request:       4837.443 [ms] (mean)
Time per request:       1209.361 [ms] (mean, across all concurrent requests)
Transfer rate:          81.81 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
Connect:        0    0   0.0      0       0
Processing:  4520 4837  74.9   4838    5090
Waiting:     4520 4836  74.7   4837    5086
Total:       4520 4837  74.9   4838    5090

Percentage of the requests served within a certain time (ms)
  50%   4838
  66%   4865
  75%   4884
  80%   4895
  90%   4931
  95%   4960
  98%   4993
  99%   5026
 100%   5090 (longest request)

Memory peak usage: 2392 [MB]
Result file size:  98.74 [KB]
```
