# Autoquality![pro](./assets/pro.svg)

imgproxy can calculate quality for your resultant images so they best fit the selected metric. The supported methods are [none](#none), [size](#autoquality-by-file-size), [dssim](#autoquality-by-dssim), and [ml](#autoquality-with-ml).

**⚠️Warning:** Autoquality requires an image to be saved several times. Use it only when you prefer the resultant quality and size over speed.

You can enable autoquality using [config options](configuration.md#autoquality) (for all images) or with [processing options](generating_the_url.md#autoquality) (for each image individually).

## None

Disable the autoquality.

**Method name:** `none`

#### Config example

```bash
IMGPROXY_AUTOQUALITY_METHOD="none"
```

#### Processing options example

```
.../autoquality:none/...
```

## Autoquality by file size

With this method, imgproxy will try to degrade the quality so your image fits the desired file size.

**Method name:** `size`

**Target:** the desired file size

#### Config example

```bash
IMGPROXY_AUTOQUALITY_METHOD="size"
# Change value to the desired size in bytes
IMGPROXY_AUTOQUALITY_TARGET=10240
IMGPROXY_AUTOQUALITY_MIN=10
IMGPROXY_AUTOQUALITY_MAX=80
# Quality 50 for AVIF is pretty the same as 80 for JPEG
IMGPROXY_AUTOQUALITY_FORMAT_MAX="avif=50"
```

#### Processing options example

```
.../autoquality:size:10240:10:80/...
```

## Autoquality by DSSIM

With this method imgproxy will try to select a level of quality so that the saved image will have the desired [DSSIM](https://en.wikipedia.org/wiki/Structural_similarity#Structural_Dissimilarity) value.

**Method name:** `dssim`

**Target:** the desired DSSIM value

#### Config example

```bash
IMGPROXY_AUTOQUALITY_METHOD="dssim"
# Change value to the desired DSSIM
IMGPROXY_AUTOQUALITY_TARGET=0.02
# We're happy enough if the resulting DSSIM will differ from the desired by 0.001
IMGPROXY_AUTOQUALITY_ALLOWED_ERROR=0.001
IMGPROXY_AUTOQUALITY_MIN=70
IMGPROXY_AUTOQUALITY_MAX=80
# Quality 50 for AVIF is pretty the same as 80 for JPEG
IMGPROXY_AUTOQUALITY_FORMAT_MIN="avif=40"
IMGPROXY_AUTOQUALITY_FORMAT_MAX="avif=50"
```

#### Processing options example

```
.../autoquality:dssim:0.02:70:80:0.001/...
```

## Autoquality with ML

This method is almost the same as autoquality with [DSSIM](#autoquality-by-dssim) but imgproxy will try to predict the initial quality using neural networks. This requires neural networks to be configured (see the config examlpe or the config documentation). If a neural network for the resulting format is not provided, the [DSSIM](#autoquality-by-dssim) method will be used instead.

**📝Note:** When this method is used, imgproxy will save JPEG images with the most optimal [advanced JPEG compression](configuration.md#advanced-jpeg-compression) settings, ignoring config and processing options.

**Method name:** `ml`

**Target:** the desired DSSIM value

#### Config example

```bash
IMGPROXY_AUTOQUALITY_METHOD="ml"
# Change value to the desired DSSIM
IMGPROXY_AUTOQUALITY_TARGET=0.02
# We're happy enough if the resulting DSSIM will differ from the desired by 0.001
IMGPROXY_AUTOQUALITY_ALLOWED_ERROR=0.001
IMGPROXY_AUTOQUALITY_MIN=70
IMGPROXY_AUTOQUALITY_MAX=80
# Quality 50 for AVIF is pretty the same as 80 for JPEG
IMGPROXY_AUTOQUALITY_FORMAT_MIN="avif=40"
IMGPROXY_AUTOQUALITY_FORMAT_MAX="avif=50"
# Neural networks paths for JPEG, WebP, and AVIF
IMGPROXY_AUTOQUALITY_JPEG_NET="/networks/autoquality-jpeg.pb"
IMGPROXY_AUTOQUALITY_WEBP_NET="/networks/autoquality-webp.pb"
IMGPROXY_AUTOQUALITY_AVIF_NET="/networks/autoquality-avif.pb"
```

**📝Note:** If you trust your neural network's autoquality, you may want to set `IMGPROXY_AUTOQUALITY_ALLOWED_ERROR` to 1 (the maximum possible DSSIM value). In this case, imgproxy will always use the quality predicted by the neural network.

#### Processing options example

```
.../autoquality:ml:0.02:70:80:0.001/...
```

### Neural networks format

Neural networks should fit the following requirements:
* Tensorflow frozen graph format
* Input layer size is 416x416
* Output layer size is 1x100
* Output layer values are logits of quality probabilities

If you're an imgproxy Pro user and you want to train your own network but you don't know how, feel free to contact the imgproxy team for instructions.
