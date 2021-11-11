# Object detection<i class='badge badge-pro'></i><i class='badge badge-v3'></i>

imgproxy can detect objects on the image and use them for smart crop, bluring the detections, or drawing the detections.

For object detection purposes, imgproxy uses [Darknet YOLO](https://github.com/AlexeyAB/darknet) model. We provide Docker images with a model trained for face detection, but you can use any Darknet YOLO model found in the [zoo](https://github.com/AlexeyAB/darknet/wiki/YOLOv4-model-zoo) or you can train your own model following the [guide](https://github.com/AlexeyAB/darknet#how-to-train-to-detect-your-custom-objects).

## Configuration

You need to define four config variables to enable object detection:

* `IMGPROXY_OBJECT_DETECTION_CONFIG`: path to the neural network config.
* `IMGPROXY_OBJECT_DETECTION_WEIGHTS`: path to the neural network weights.
* `IMGPROXY_OBJECT_DETECTION_CLASSES`: path to the text file with the classes names, one by line.
* `IMGPROXY_OBJECT_DETECTION_NET_SIZE`: the size of the neural network input. The width and the heights of the inputs should be the same, so this config value should be a single number. Default: 416.

Read the [configuration](configuration.md#object-detection) guide for more config values info.

## Usage examples
### Object-oriented crop

You can [crop](https://docs.imgproxy.net/generating_the_url?id=crop) your images and keep objects of desired classes in frame:

```
.../crop:256:256/g:obj:face/...
```

### Bluring detections

You can [blur objects](https://docs.imgproxy.net/generating_the_url?id=blur-detections) of desired classes for anonymization or hiding NSFW content:

```
.../blur_detections:7:face/...
```

### Draw detections

You can make imgproxy [draw bounding boxes](https://docs.imgproxy.net/generating_the_url?id=draw-detections) of detected objects of desired classes (handy for testing your models):

```
.../draw_detections:1:face/...
```
