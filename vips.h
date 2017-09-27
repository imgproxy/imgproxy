#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

#define VIPS_SUPPORT_SMARTCROP \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 5))

#define VIPS_SUPPORT_GIF \
  VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 3)

enum types {
	JPEG = 0,
  PNG,
	WEBP,
	GIF
};

int
vips_initialize()
{
  return vips_init("imgproxy");
}

void
swap_and_clear(VipsImage **in, VipsImage *out) {
  g_object_unref((gpointer) *in);
  *in = out;
}

int
vips_type_find_load_go(int imgtype) {
  if (imgtype == JPEG) {
    return vips_type_find("VipsOperation", "jpegload");
  }
  if (imgtype == PNG) {
    return vips_type_find("VipsOperation", "pngload");
  }
  if (imgtype == WEBP) {
    return vips_type_find("VipsOperation", "webpload");
  }
	if (imgtype == GIF) {
		return vips_type_find("VipsOperation", "gifload");
	}
	return 0;
}

int
vips_type_find_save_go(int imgtype) {
  if (imgtype == JPEG) {
    return vips_type_find("VipsOperation", "jpegsave_buffer");
  }
  if (imgtype == PNG) {
    return vips_type_find("VipsOperation", "pngsave_buffer");
  }
	if (imgtype == WEBP) {
		return vips_type_find("VipsOperation", "webpsave_buffer");
	}
	return 0;
}

int
vips_jpegload_buffer_go(void *buf, size_t len, VipsImage **out)
{
  return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
}

int
vips_pngload_buffer_go(void *buf, size_t len, VipsImage **out)
{
  return vips_pngload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
}

int
vips_gifload_buffer_go(void *buf, size_t len, VipsImage **out)
{
#if VIPS_SUPPORT_GIF
  return vips_gifload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
#else
  return 0;
#endif
}

int
vips_webpload_buffer_go(void *buf, size_t len, VipsImage **out)
{
  return vips_webpload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_resize_go(VipsImage *in, VipsImage **out, double scale)
{
  return vips_resize(in, out, scale, NULL);
}

int
vips_support_smartcrop() {
#if VIPS_SUPPORT_SMARTCROP
	return 1;
#else
	return 0;
#endif
}

int
vips_smartcrop_go(VipsImage *in, VipsImage **out, int width, int height) {
#if VIPS_SUPPORT_SMARTCROP
	return vips_smartcrop(in, out, width, height, NULL);
#else
	return 0;
#endif
}

int
vips_colourspace_go(VipsImage *in, VipsImage **out, VipsInterpretation space)
{
  return vips_colourspace(in, out, space, NULL);
}

int
vips_extract_area_go(VipsImage *in, VipsImage **out, int left, int top, int width, int height)
{
  return vips_extract_area(in, out, left, top, width, height, NULL);
}

int
vips_process_image(VipsImage **img, int resize, double scale, int crop, int smart, int left, int top, int width, int height)
{
  VipsImage *tmp;
  int err;

  if (resize > 0) {
    err = vips_resize_go(*img, &tmp, scale);
    swap_and_clear(img, tmp);
    if (err> 0) { return 1; }
  }

  if (crop > 0) {
    if (smart > 0) {
      err = vips_smartcrop_go(*img, &tmp, width, height);
      swap_and_clear(img, tmp);
      if (err> 0) { return 1; }
    } else {
      vips_extract_area_go(*img, &tmp, left, top, width, height);
      swap_and_clear(img, tmp);
      if (err> 0) { return 1; }
    }
  }

  err = vips_colourspace_go(*img, &tmp, VIPS_INTERPRETATION_sRGB);
  swap_and_clear(img, tmp);

  return err;
}

int
vips_jpegsave_go(VipsImage *in, void **buf, size_t *len, int strip, int quality, int interlace)
{
  return vips_jpegsave_buffer(in, buf, len, "strip", strip, "Q", quality, "optimize_coding", TRUE, "interlace", interlace, NULL);
}

int
vips_pngsave_go(VipsImage *in, void **buf, size_t *len)
{
  return vips_pngsave_buffer(in, buf, len, "filter", VIPS_FOREIGN_PNG_FILTER_NONE, NULL);
}

int
vips_webpsave_go(VipsImage *in, void **buf, size_t *len, int strip, int quality)
{
	return vips_webpsave_buffer(in, buf, len, "strip", strip, "Q", quality, NULL);
}

void
vips_cleanup()
{
  vips_thread_shutdown();
  vips_error_clear();
}
