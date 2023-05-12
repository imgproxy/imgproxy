# Presets

An imgproxy preset is a named set of processing options. Presets can be used in [URLs](generating_the_url.md#preset) to make shorter and more human-readable.

## Presets definition

A preset definition looks like this:

```
%preset_name=%processing_options
```

Processing options should be defined in the same way they are defined in [URLs](generating_the_url.md#processing-options). For example, here's a preset named `awesome` that sets the resizing type to `fill` and the resulting format to `jpg`:

```
awesome=resizing_type:fill/format:jpg
```

Read how to specify your presets with imgproxy in the [Configuration](configuration.md) guide.

## Default preset

A preset named `default` will be applied to each image. This is useful when you want your default processing options to be different from the default imgproxy options.

## Only presets

Setting `IMGPROXY_ONLY_PRESETS` to `true` switches imgproxy into "presets-only mode". In this mode, imgproxy accepts a presets list as processing options just like you'd specify them for the `preset` option:

```
http://imgproxy.example.com/unsafe/thumbnail:blurry:watermarked/plain/http://example.com/images/curiosity.jpg@png
```

All other URL formats are disabled in this mode.
