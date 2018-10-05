# Presets

Preset is named set of processing options. Presets can be used in [advanced URL format](./generating_the_url_advanced.md#preset) to get shorter and more readable URLs.

### Presets definition

Preset definition looks like this:

```
%preset_name=%processing_options
```

Processing options should be defined the same way as you define them in the [advanced URL format](./generating_the_url_advanced.md#preset). For example, preset named `awesome` that sets the resizing type to `fill` and resulting format to `jpg` will look like this:

```
awesome=resizing_type:fill/format:jpg
```

Read how to specify your presets to imgproxy in [Configuration](./configuration.md) guide.

### Default preset

Preset named `default` will be applied to each image. This is useful when you want your default processing options to be different from imgproxy default ones.
