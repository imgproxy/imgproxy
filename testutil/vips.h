#pragma once

#include <vips/vips.h>

/* VIPS_LOAD_IMAGE is a helper macros to load a VipsImage from memory buffer.
 * If the image is SVG, it reloads it with the correct DRI.
 */
#define VIPS_LOAD_IMAGE(IM, BUF, BUF_LEN) \
  { \
    IM = vips_image_new_from_buffer(BUF, BUF_LEN, "", NULL); \
\
    const char *loader; \
\
    if (IM && \
        vips_image_get_typeof(IM, VIPS_META_LOADER) == G_TYPE_STRING && \
        !vips_image_get_string(IM, VIPS_META_LOADER, &loader) && \
        !strcmp(loader, "svgload")) { \
      VIPS_UNREF(IM); \
      IM = vips_image_new_from_buffer(BUF, BUF_LEN, "[dpi=96.0,scale=0.75]", NULL); \
    } \
  }

/* vips_rgba_image_data reads the image from the input buffer, converts it to RGBA format
 * and returns the raw pixel data in the output buffer.
 */
int
vips_rgba_image_data(void *in_buf, size_t in_buf_size, void **out, size_t *out_size);
