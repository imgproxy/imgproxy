# Memory usage tweaks

There are some imgproxy options that can help you to optimize memory usage and decrease memory fragmentation.

**Warning:** This is an advanced part. Please make sure that you know what're you doing before changing anything.

### `IMGPROXY_DOWNLOAD_BUFFER_SIZE`

imgproxy uses memory buffers to download source images. While these buffers are empty at the start by default, they can grow to a required size when imgproxy downloads an image. Allocating new memory to grow the buffers can cause memory fragmentation. Allocating required memory at the start can eliminate much of memory fragmentation since buffers won't grow. Setting `IMGPROXY_DOWNLOAD_BUFFER_SIZE` will tell imgproxy to initialize download buffers with _at least_ the specified size. It's recommended to use the estimates size of your biggest image as the initial download buffers size.

### `IMGPROXY_GZIP_BUFFER_SIZE`

The same as `IMGPROXY_DOWNLOAD_BUFFER_SIZE` but for GZip buffers. If you use GZip compression of the resulting images, you can reduce memory fragmentation by using the estimated maximum size of the GZipped resulting image as the initial size of GZip buffers.

### `IMGPROXY_FREE_MEMORY_INTERVAL`

Working with a large amount of data can cause allocating some memory that is not used most of the time. That's why imgproxy enforces Go's garbage collector to free as much memory as possible and return it to the OS. The default interval of this action is 10 seconds, but you can change it by setting `IMGPROXY_FREE_MEMORY_INTERVAL`. Decreasing the interval can smooth the memory usage graph but it can also slow down imgproxy a little. Increasing has the opposite effect.
