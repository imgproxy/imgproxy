#include "image_hash.h"

/**
 * vips_image_read_to_memory: converts VipsImage to RGBA format and reads into memory buffer
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
vips_image_read_to_memory(VipsImage *in, void **buf, size_t *size)
{
  VipsImage *rgba_image = NULL;

  if (!in || !buf || !size) {
    vips_error("vips_image_read_to_memory", "invalid arguments");
    return -1;
  }

  // Initialize output parameters
  *buf = NULL;
  *size = 0;

  // Convert to sRGB colorspace first if needed
  if (vips_colourspace(in, &rgba_image, VIPS_INTERPRETATION_sRGB, NULL) != 0) {
    vips_error("vips_image_read_to_memory", "failed to convert to sRGB");
    return -1;
  }

  // Add alpha channel if not present (convert to RGBA)
  VipsImage *with_alpha = NULL;
  if (vips_image_hasalpha(rgba_image)) {
    // Already has alpha, just reference it
    with_alpha = rgba_image;
    g_object_ref(with_alpha);
  }
  else {
    // Add alpha channel
    if (vips_addalpha(rgba_image, &with_alpha, NULL) != 0) {
      g_object_unref(rgba_image);
      vips_error("vips_image_read_to_memory", "failed to add alpha channel");
      return -1;
    }
  }
  g_object_unref(rgba_image);

  // Get raw pixel data
  *buf = vips_image_write_to_memory(with_alpha, size);
  g_object_unref(with_alpha);

  if (*buf == NULL) {
    vips_error("vips_image_read_to_memory", "failed to write image to memory");
    return -1;
  }

  return 0;
}

/**
 * vips_memory_buffer_free: frees memory buffer allocated by vips_image_write_to_memory
 * @buf: buffer pointer to free
 *
 * Frees the memory buffer allocated by vips_image_write_to_memory.
 * Safe to call with NULL pointer.
 */
void
vips_memory_buffer_free(void *buf)
{
  if (buf) {
    g_free(buf);
  }
}
