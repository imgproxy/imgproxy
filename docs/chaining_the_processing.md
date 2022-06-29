# Chaining the processing![pro](/assets/pro.svg)

Though imgproxy's [processing pipeline](about_processing_pipeline.md) is suitable for most cases, sometimes it's handy to run multiple chained pipelines with different options.

imgproxy Pro allows you to start a new pipeline by inserting a section with a minus sign (`-`) to the URL path:

```
.../width:500/crop:1000/-/trim:10/...
                        ^ the new pipeline starts here
```

### Example 1: Multiple watermarks

If you need to place multiple watermarks on the same image, you can use chained pipelines for that:

```
.../rs:fit:500:500/wm:0.5:nowe/wmu:aW1hZ2UxCg/-/wm:0.7:soea/wmu:aW1hZ2UyCg/...
```

In this example, the first pipeline resizes the image and places the first watermark, and the second pipeline places the second watermark.

### Example 2: Fast trim

The `trim` operation is pretty heavy as it involves loading the whole image to the memory at the very start of processing. However, if you're going to scale down your image and the trim accuracy is not very important to you, it's better to move trimming to a separate pipeline.

```
.../rs:fit:500:500/-/trim:10/...
```

In this example, the first pipeline resizes the image, and the second pipeline trims the result. Since the result of the first pipeline is already resized and loaded to the memory, trimming will be done much faster.

## Using with presets

You can use presets in your chained pipelines, and you can use chained pipelines in your presets. However, the behaior may be not obvious. The rules are the following:

* Prest is applied to the pipeline where is was used.
* Preset may contain chained pipelined, and ones will be chained to the pipeline where the preset was used.
* Chained pipelines from the preset and from the URL are merged.

### Example

If we have the following preset

```
test=width:300/height:300/-/width:200/height:200/-/width:100/height:200
```

and the following URL

```
.../width:400/-/preset:test/width:500/-/width:600/...
```

The result will look like this:

```
.../width:400/-/width:500/height:300/-/width:600/height:200/-/width:100/height:200/...
```
