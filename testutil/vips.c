#include "vips.h"

/**
 * vips_rgba_image_data: converts VipsImage to RGBA format and reads into memory buffer
 * @in: VipsImage to convert and read
 * @buf: pointer to buffer pointer (will be allocated)
 * @size: pointer to size_t to store the buffer size
 *
 * Converts the VipsImage to RGBA format using VIPS operations and reads the raw pixel data.
 * The caller is responsible for freeing the buffer using vips_memory_buffer_free().
 *
 * Returns: 0 on success, -1 on error
 */
int
vips_rgba_image_data(void *in_buf, size_t in_buf_size, void **out, size_t *out_size)
{
  if (!in_buf || !in_buf_size || !out || !out_size) {
    vips_error("vips_rgba_image_data", "invalid arguments");
    return -1;
  }

  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 4);

  VIPS_LOAD_IMAGE(t[0], in_buf, in_buf_size);
  if (t[0] == NULL) {
    return -1;
  }

  // NOTE: Looks like there is a vips bug: if there were errors during format detection,
  // they would be saved in vips error buffer. However, if format detection was successful,
  // we expect that error buffer should be empty.
  // vips_error_clear();

  // Initialize output parameters
  *out = NULL;
  *out_size = 0;

  VipsImage *in = t[0];

  // Convert to sRGB colorspace first if needed
  if (vips_colourspace(in, &t[1], VIPS_INTERPRETATION_sRGB, NULL) != 0) {
    VIPS_UNREF(base);
    vips_error("vips_rgba_image_data", "failed to convert to sRGB");
    return -1;
  }

  in = t[1];

  // Add alpha channel if not present (convert to RGBA)
  if (!vips_image_hasalpha(in)) {
    // Add alpha channel
    if (vips_addalpha(in, &t[2], NULL) != 0) {
      VIPS_UNREF(base);
      vips_error("vips_rgba_image_data", "failed to add alpha channel");
      return -1;
    }
    in = t[2];
  }

  // Get raw pixel data, width and height
  *out = vips_image_write_to_memory(in, out_size);

  // Dispose the image regardless of the result
  VIPS_UNREF(base);

  if (*out == NULL) {
    vips_error("vips_rgba_image_data", "failed to write image to memory");
    return -1;
  }

  return 0;
}
