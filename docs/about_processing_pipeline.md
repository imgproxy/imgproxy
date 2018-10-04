# About processing pipeline

imgproxy has a fixed processing pipeline that tuned for maximum performance. When you process an image with imgproxy, it does the following things:

* If source image format allows shrink-on-load, imgproxy uses it to quickly resize image to the size closest to desired;
* If it's needed to resize an image with alpha-channel, imgproxy premultiplies one to handle alpha correctly;
* Resize image to desired size;
* If image colorspace need to be fixed, imgproxy does this;
* Rotate/flip image according to EXIF metadata;
* Crop image using specified gravity;
* Fill image background if some background color was specified;
* Apply gaussian blur and sharpen filters;
* And finally save the image to the desired format.

This pipeline with using a sequential access to source image data allows to highly reduce memory and CPU usage, that makes imgproxy so awesome.
