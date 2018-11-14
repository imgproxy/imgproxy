#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>
#include "image_types.h"

#define VIPS_SUPPORT_SMARTCROP \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 5))

#define VIPS_SUPPORT_HASALPHA \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 5))

#define VIPS_SUPPORT_GIF \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 3))

#define VIPS_SUPPORT_MAGICK \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 7))

#define EXIF_ORIENTATION "exif-ifd0-Orientation"

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
    return vips_type_find("VipsOperation", "jpegload_buffer");
  }
  if (imgtype == PNG) {
    return vips_type_find("VipsOperation", "pngload_buffer");
  }
  if (imgtype == WEBP) {
    return vips_type_find("VipsOperation", "webpload_buffer");
  }
  if (imgtype == GIF) {
    return vips_type_find("VipsOperation", "gifload_buffer");
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
  if (imgtype == GIF) {
    return vips_type_find("VipsOperation", "magicksave_buffer");
  }
  return 0;
}

int
vips_jpegload_go(void *buf, size_t len, int shrink, VipsImage **out) {
  if (shrink > 1) {
    return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "shrink", shrink, NULL);
  }
  return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_pngload_go(void *buf, size_t len, VipsImage **out) {
  return vips_pngload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_webpload_go(void *buf, size_t len, int shrink, VipsImage **out) {
  if (shrink > 1) {
    return vips_webpload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "shrink", shrink, NULL);
  }
  return vips_webpload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_gifload_go(void *buf, size_t len, int pages, VipsImage **out) {
  #if VIPS_SUPPORT_GIF
    return vips_gifload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "n", pages, NULL);
  #else
    vips_error("vips_gifload_go", "Loading GIF is not supported");
    return 1;
  #endif
}

int
vips_get_exif_orientation(VipsImage *image) {
	const char *orientation;

	if (
		vips_image_get_typeof(image, EXIF_ORIENTATION) != 0 &&
		!vips_image_get_string(image, EXIF_ORIENTATION, &orientation)
	) return atoi(&orientation[0]);

  vips_error("vips_get_exif_orientation", "Can't get EXIF orientation");
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

VipsBandFormat
vips_band_format(VipsImage *in) {
  return in->BandFmt;
}

gboolean
vips_image_hasalpha_go(VipsImage * in) {
#if VIPS_SUPPORT_HASALPHA
  return vips_image_hasalpha(in);
#else
  return( in->Bands == 2 ||
		      (in->Bands == 4 && in->Type != VIPS_INTERPRETATION_CMYK) ||
		      in->Bands > 4 );
#endif
}

int
vips_premultiply_go(VipsImage *in, VipsImage **out) {
  return vips_premultiply(in, out, NULL);
}

int
vips_unpremultiply_go(VipsImage *in, VipsImage **out) {
  return vips_unpremultiply(in, out, NULL);
}

int
vips_cast_go(VipsImage *in, VipsImage **out, VipsBandFormat format) {
  return vips_cast(in, out, format, NULL);
}

int
vips_resize_go(VipsImage *in, VipsImage **out, double scale) {
  return vips_resize(in, out, scale, NULL);
}

int
vips_need_icc_import(VipsImage *in) {
  return in->Type == VIPS_INTERPRETATION_CMYK;
}

int
vips_icc_import_go(VipsImage *in, VipsImage **out, char *profile) {
  return vips_icc_import(in, out, "input_profile", profile, "embedded", TRUE, "pcs", VIPS_PCS_XYZ, NULL);
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
  vips_error("vips_smartcrop_go", "Smart crop is not supported");
  return 1;
#endif
}

int
vips_gaussblur_go(VipsImage *in, VipsImage **out, double sigma) {
  return vips_gaussblur(in, out, sigma, NULL);
}

int
vips_sharpen_go(VipsImage *in, VipsImage **out, double sigma) {
  return vips_sharpen(in, out, "sigma", sigma, NULL);
}

int
vips_flatten_go(VipsImage *in, VipsImage **out, double r, double g, double b) {
  VipsArrayDouble *bg = vips_array_double_newv(3, r, g, b);
  int res = vips_flatten(in, out, "background", bg, NULL);
  vips_area_unref((VipsArea *)bg);
  return res;
}

int
vips_extract_area_go(VipsImage *in, VipsImage **out, int left, int top, int width, int height) {
  return vips_extract_area(in, out, left, top, width, height, NULL);
}

int
vips_replicate_go(VipsImage *in, VipsImage **out, int across, int down) {
  return vips_replicate(in, out, across, down, NULL);
}

int
vips_embed_go(VipsImage *in, VipsImage **out, int x, int y, int width, int height) {
  return vips_embed(in, out, x, y, width, height, NULL);
}

int
vips_extract_band_go(VipsImage *in, VipsImage **out, int band, int band_num) {
  return vips_extract_band(in, out, band, "n", band_num, NULL);
}

int
vips_bandjoin_go (VipsImage *in1, VipsImage *in2, VipsImage **out) {
  return vips_bandjoin2(in1, in2, out, NULL);
}

int
vips_bandjoin_const_go (VipsImage *in, VipsImage **out, double c) {
  return vips_bandjoin_const1(in, out, c, NULL);
}

int
vips_linear_go (VipsImage *in, VipsImage **out, double a, double b) {
  return vips_linear1(in, out, a, b, NULL);
}

int
vips_ifthenelse_go(VipsImage *cond, VipsImage *in1, VipsImage *in2, VipsImage **out) {
  return vips_ifthenelse(cond, in1, in2, out, "blend", TRUE, NULL);
}

int
vips_arrayjoin_go(VipsImage **in, VipsImage **out, int n) {
  return vips_arrayjoin(in, out, n, "across", 1, NULL);
}

int
vips_jpegsave_go(VipsImage *in, void **buf, size_t *len, int strip, int quality, int interlace) {
  return vips_jpegsave_buffer(in, buf, len, "strip", strip, "Q", quality, "optimize_coding", TRUE, "interlace", interlace, NULL);
}

int
vips_pngsave_go(VipsImage *in, void **buf, size_t *len, int interlace) {
  return vips_pngsave_buffer(in, buf, len, "filter", VIPS_FOREIGN_PNG_FILTER_NONE, "interlace", interlace, NULL);
}

int
vips_webpsave_go(VipsImage *in, void **buf, size_t *len, int strip, int quality) {
  return vips_webpsave_buffer(in, buf, len, "strip", strip, "Q", quality, NULL);
}

int
vips_gifsave_go(VipsImage *in, void **buf, size_t *len) {
#if VIPS_SUPPORT_MAGICK
  return vips_magicksave_buffer(in, buf, len, "format", "gif", NULL);
#else
  vips_error("vips_gifsave_go", "Saving GIF is not supported");
  return 1;
#endif
}

void
vips_cleanup() {
  vips_thread_shutdown();
  vips_error_clear();
}
