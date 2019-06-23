# Memory usage tweaks

There are some imgproxy options that can help you to optimize memory usage and decrease memory fragmentation.

**Warning:** This is an advanced part. Please make sure that you know what you're doing before changing anything.

### `IMGPROXY_DOWNLOAD_BUFFER_SIZE`

imgproxy uses memory buffers to download source images. While these buffers are empty at the start by default, they can grow to a required size when imgproxy downloads an image. Allocating new memory to grow the buffers can cause memory fragmentation. Allocating required memory at the start can eliminate much of memory fragmentation since buffers won't grow. Setting `IMGPROXY_DOWNLOAD_BUFFER_SIZE` will tell imgproxy to initialize download buffers with _at least_ the specified size. It's recommended to use the estimated 95 percentile of your image sizes as the initial download buffers size.

### `IMGPROXY_GZIP_BUFFER_SIZE`

The same as `IMGPROXY_DOWNLOAD_BUFFER_SIZE` but for GZip buffers. If you use GZip compression of the resulting images, you can reduce memory fragmentation by using the estimated maximum size of the GZipped resulting image as the initial size of GZip buffers.

### `IMGPROXY_FREE_MEMORY_INTERVAL`

Working with a large amount of data can cause allocating some memory that is not used most of the time. That's why imgproxy enforces Go's garbage collector to free as much memory as possible and return it to the OS. The default interval of this action is 10 seconds, but you can change it by setting `IMGPROXY_FREE_MEMORY_INTERVAL`. Decreasing the interval can smooth the memory usage graph but it can also slow down imgproxy a little. Increasing has the opposite effect.

### `IMGPROXY_BUFFER_POOL_CALIBRATION_THRESHOLD`

Buffer pools in imgproxy do self-calibration time by time. imgproxy collects stats about the sizes of the buffers returned to a pool and calculates the default buffer size and the maximum size of a buffer that can be returned to the pool. This allows dropping buffers that are too big for most of the images and save some memory. By default, imgproxy starts calibration after 1024 buffers were returned to a pool. You can change this number with `IMGPROXY_BUFFER_POOL_CALIBRATION_THRESHOLD` variable. Increasing the number will give you rarer but more accurate calibration.

### `MALLOC_ARENA_MAX`

`libvips` uses GLib for memory management, and it brings GLib memory fragmentation issues to heavily multi-threaded programs. imgproxy is defenitely one of them. First thing you can try if you noticed constantly growing RSS usage without Go's sys memory growth is set `MALLOC_ARENA_MAX`:

```
MALLOC_ARENA_MAX=2 imgproxy
```

This will reduce GLib memory appetites by reducing the number of malloc arenas that it can create. By default GLib creates one are per thread, and this would folow to a memory fragmentation.


### Using [jemalloc](http://jemalloc.net/)

If setting `MALLOC_ARENA_MAX` doesn't show you satisfying results, it's time to try jemalloc. As jemalloc site says:

> jemalloc is a general purpose malloc(3) implementation that emphasizes fragmentation avoidance and scalable concurrency support.

Most Linux distributives provide their jemalloc packages. Using jemalloc doesn't require rebuilding imgproxy or it's dependencies and can be enabled by `LD_PRELOAD` environment variable. See Debian example below. Not that jemalloc library path may vary in your system.

```
sudo apt-get install libjemalloc1
LD_PRELOAD='/usr/lib/x86_64-linux-gnu/libjemalloc.so.1' imgproxy
```
