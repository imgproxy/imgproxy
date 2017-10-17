#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>

#define VIPS_SUPPORT_SMARTCROP \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 5))

#define VIPS_SUPPORT_GIF \
  VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 3)

#define EXIF_ORIENTATION "exif-ifd0-Orientation"

enum types {
  UNKNOWN = 0,
  JPEG,
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
vips_load_buffer(void *buf, size_t len, int imgtype, int shrink, VipsImage **out) {
  switch (imgtype) {
    case JPEG:
      if (shrink > 1) {
        return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "shrink", shrink, NULL);
      }
      return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
    case PNG:
      return vips_pngload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
    case WEBP:
      if (shrink > 1) {
        return vips_webpload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "shrink", shrink, NULL);
      }
      return vips_webpload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
    #if VIPS_SUPPORT_GIF
    case GIF:
      return vips_gifload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
    #endif
  }
  return 1;
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
vips_support_smartcrop() {
#if VIPS_SUPPORT_SMARTCROP
  return 1;
#else
  return 0;
#endif
}

int
vips_resize_go(VipsImage *in, VipsImage **out, double scale) {
  return vips_resize(in, out, scale, NULL);
}

int
vips_colourspace_go(VipsImage *in, VipsImage **out, VipsInterpretation cs) {
  return vips_colourspace(in, out, cs, NULL);
}

int
vips_rot_go(VipsImage *in, VipsImage **out, VipsAngle angle) {
  return vips_rot(in, out, angle, NULL);
}

int
vips_flip_horizontal_go(VipsImage *in, VipsImage **out) {
  return vips_flip(in, out, VIPS_DIRECTION_HORIZONTAL, NULL);
}

int
vips_smartcrop_go(VipsImage *in, VipsImage **out, int width, int height) {
#if VIPS_SUPPORT_SMARTCROP
  return vips_smartcrop(in, out, width, height, NULL);
#else
  return 1;
#endif
}

int
vips_extract_area_go(VipsImage *in, VipsImage **out, int left, int top, int width, int height) {
  return vips_extract_area(in, out, left, top, width, height, NULL);
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
