# Presets

imgproxy preset is a named set of processing options. Presets can be used in [advanced URL format](generating_the_url_advanced.md#preset) to get shorter and somewhat readable URLs.

## Presets definition

The preset definition looks like this:

```
%preset_name=%processing_options
```

Processing options should be defined in the same way as you define them in the [advanced URL format](generating_the_url_advanced.md#preset). For example, here is a preset named `awesome` that sets the resizing type to `fill` and resulting format to `jpg`:

```
awesome=resizing_type:fill/format:jpg
```

Read how to specify your presets with imgproxy in the [Configuration](configuration.md) guide.

## Default preset

A preset named `default` will be applied to each image. Useful in case you want your default processing options to be different from the imgproxy default ones.

## Only presets

Setting `IMGPROXY_ONLY_PRESETS` as `true` switches imgproxy into "presets-only mode". In this mode imgproxy accepts presets list as processing options just like you'd specify them for the `preset` option:

```
http://imgproxy.example.com/unsafe/thumbnail:blurry:watermarked/plain/http://example.com/images/curiosity.jpg@png
```

All other URL formats are disabled in this mode.

## Exclude presets from URL signature validation

Setting `IMGPROXY_EXCLUDE_PRESETS_FROM_SIGNATURE` as `true` excludes the presets segment of the URL path from signature validation. Note that this only works when `IMGPROXY_ONLY_PRESETS` is set to `true`. Thus, you can use a single digitally signed URL for fetching multiple pre-defined formats of the same image URL.

