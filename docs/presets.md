# Presets

imgproxy preset is a named set of processing options. Presets can be used in [advanced URL format](./generating_the_url_advanced.md#preset) to get shorter and somewhat readable URLs.

### Presets definition

The preset definition looks like this:

```
%preset_name=%processing_options
```

Processing options should be defined in the same way as you define them in the [advanced URL format](./generating_the_url_advanced.md#preset). For example, here is a preset named `awesome` that sets the resizing type to `fill` and resulting format to `jpg`:

```
awesome=resizing_type:fill/format:jpg
```

Read how to specify your presets with imgproxy in the [Configuration](./configuration.md) guide.

### Default preset

A preset named `default` will be applied to each image. Useful in case you want your default processing options to be different from the imgproxy default ones.

### Only presets

If you set `IMGPROXY_ONLY_PRESETS` as `true`, a preset is obligatory, and all other URL formats are disabled.

In this case, you always need to inform a preset in your URLs without the `preset` or `pr` statement. Example: `http://imgproxy.example.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/thumbnail/aHR0cDovL2V4YW1w/bGUuY29tL2ltYWdl/cy9jdXJpb3NpdHku/anBn.png`

It's possible to use more than one preset separing them with `:` like `thumbnail:gray`.