# About the processing pipeline

imgproxy has a specific processing pipeline tuned for maximum performance. When you process an image with imgproxy, it does the following:

* If the source image format allows shrink-on-load, imgproxy uses it to quickly resize the image to the size that is closest to desired.
* If it is needed to resize an image with an alpha-channel, imgproxy premultiplies one to handle alpha correctly.
* imgproxy resizes the image to the desired size.
* If the image colorspace need to be fixed, imgproxy fixes it.
* imgproxy rotates/flip the image according to EXIF metadata.
* imgproxy crops the image using the specified gravity.
* imgproxy fills the image background if the background color was specified.
* imgproxy applies gaussian blur and sharpen filters.
* imgproxy adds a watermark if one was specified.
* And finally, imgproxy saves the image to the desired format.

This pipeline, using sequential access to source image data, allows for significantly reduced memory and CPU usage — one of the reasons imgproxy is so performant.
