#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

#define VIPS_SUPPORT_SMARTCROP \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 5))

#define VIPS_SUPPORT_GIF \
  VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 3)

#define EXIF_ORIENTATION "exif-ifd0-Orientation"

enum types {
  JPEG = 0,
  PNG,
  WEBP,
  GIF
};

int
vips_initialize() {
  return vips_init("imgproxy");
}

void
clear_image(VipsImage **in) {
  if (G_IS_OBJECT(*in)) g_clear_object(in);
}

void
g_free_go(void **buf) {
  g_free(*buf);
}

VipsAccess
access_mode(int random) {
  if (random > 0) return VIPS_ACCESS_RANDOM;
  return VIPS_ACCESS_SEQUENTIAL;
}

void
swap_and_clear(VipsImage **in, VipsImage *out) {
  clear_image(in);
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
vips_jpegload_buffer_go(void *buf, size_t len, VipsImage **out, int random) {
  return vips_jpegload_buffer(buf, len, out, "access", access_mode(random), NULL);
}

int
vips_pngload_buffer_go(void *buf, size_t len, VipsImage **out, int random) {
  return vips_pngload_buffer(buf, len, out, "access", access_mode(random), NULL);
}

int
vips_gifload_buffer_go(void *buf, size_t len, VipsImage **out, int random) {
#if VIPS_SUPPORT_GIF
  return vips_gifload_buffer(buf, len, out, "access", access_mode(random), NULL);
#else
  return 0;
#endif
}

int
vips_webpload_buffer_go(void *buf, size_t len, VipsImage **out, int random) {
  return vips_webpload_buffer(buf, len, out, "access", access_mode(random), NULL);
}

int
vips_get_exif_orientation(VipsImage *image) {
	const char *orientation;

	if (
		vips_image_get_typeof(image, EXIF_ORIENTATION) != 0 &&
		!vips_image_get_string(image, EXIF_ORIENTATION, &orientation)
	) return atoi(&orientation[0]);

	return 1;
}

int
vips_exif_rotate(VipsImage **img, int orientation) {
  int err;
  int angle = VIPS_ANGLE_D0;
  gboolean flip = FALSE;

  VipsImage *tmp;

  if (orientation == 3 || orientation == 4) angle = VIPS_ANGLE_D180;
  if (orientation == 5 || orientation == 6) angle = VIPS_ANGLE_D90;
  if (orientation == 7 || orientation == 8) angle = VIPS_ANGLE_D270;
  if (orientation == 2 || orientation == 4 || orientation == 5 || orientation == 7) {
    flip = TRUE;
  }

  err = vips_rot(*img, &tmp, angle, NULL);
  swap_and_clear(img, tmp);
  if (err > 0) { return err; }

  if (flip) {
    err = vips_flip(*img, &tmp, VIPS_DIRECTION_HORIZONTAL, NULL);
    swap_and_clear(img, tmp);
    if (err > 0) { return err; }
  }

  return 0;
}

int
vips_resize_go(VipsImage *in, VipsImage **out, double scale) {
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
vips_colourspace_go(VipsImage *in, VipsImage **out, VipsInterpretation space) {
  return vips_colourspace(in, out, space, NULL);
}

int
vips_extract_area_go(VipsImage *in, VipsImage **out, int left, int top, int width, int height) {
  return vips_extract_area(in, out, left, top, width, height, NULL);
}

int
vips_process_image(VipsImage **img, int resize, double scale, int crop, int smart, int left, int top, int width, int height) {
  VipsImage *tmp;
  int err;

  int exif_orientation = vips_get_exif_orientation(*img);
  if (exif_orientation > 1) {
    vips_exif_rotate(img, exif_orientation);
    if (err > 0) { return 1; }
  }

  if (resize > 0) {
    err = vips_resize_go(*img, &tmp, scale);
    swap_and_clear(img, tmp);
    if (err > 0) { return 1; }
  }

  if (crop > 0) {
    if (smart > 0) {
      err = vips_smartcrop_go(*img, &tmp, width, height);
      swap_and_clear(img, tmp);
      if (err > 0) { return 1; }
    } else {
      vips_extract_area_go(*img, &tmp, left, top, width, height);
      swap_and_clear(img, tmp);
      if (err > 0) { return 1; }
    }
  }

  if (vips_image_guess_interpretation(*img) != VIPS_INTERPRETATION_sRGB) {
    err = vips_colourspace_go(*img, &tmp, VIPS_INTERPRETATION_sRGB);
    swap_and_clear(img, tmp);
  }

  return err;
}

int
vips_jpegsave_go(VipsImage *in, void **buf, size_t *len, int strip, int quality, int interlace) {
  return vips_jpegsave_buffer(in, buf, len, "strip", strip, "Q", quality, "optimize_coding", TRUE, "interlace", interlace, NULL);
}

int
vips_pngsave_go(VipsImage *in, void **buf, size_t *len) {
  return vips_pngsave_buffer(in, buf, len, "filter", VIPS_FOREIGN_PNG_FILTER_NONE, NULL);
}

int
vips_webpsave_go(VipsImage *in, void **buf, size_t *len, int strip, int quality) {
  return vips_webpsave_buffer(in, buf, len, "strip", strip, "Q", quality, NULL);
}

void
vips_cleanup() {
  vips_thread_shutdown();
  vips_error_clear();
}
