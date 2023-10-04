# Presets

An imgproxy preset is a named set of processing or info options. Presets can be used in [processing URLs](generating_the_url.md#preset) or [info URLs](getting_the_image_info.md#preset) to make them shorter and more human-readable.

## Presets definition

A preset definition looks like this:

```
%preset_name=%options
```

Options should be defined in the same way they are defined in [processing URLs](generating_the_url.md#processing-options) and [info URLs](getting_the_image_info.md#processing-options). For example, here's a preset named `awesome` that sets the resizing type to `fill` and the resulting format to `jpg`:

```
awesome=resizing_type:fill/format:jpg
```

Read how to specify your presets with imgproxy in the [Configuration](configuration.md#presets) guide.

## Default preset

A preset named `default` will be applied to each image. This is useful when you want your default processing options to be different from the default imgproxy options.

## Only presets

Setting `IMGPROXY_ONLY_PRESETS` to `true` switches imgproxy into presets-only mode. In this mode, imgproxy accepts a presets list as processing options just like you'd specify them for the `preset` option:

```
http://imgproxy.example.com/unsafe/thumbnail:blurry:watermarked/plain/http://example.com/images/curiosity.jpg@png
```

You can enable or disable the presets-only mode for the [info](getting_the_image_info.md) endpoint using the `IMGPROXY_INFO_ONLY_PRESETS` config. If `IMGPROXY_INFO_ONLY_PRESETS` is not set, the info endpoint respects the `IMGPROXY_ONLY_PRESETS` value.
