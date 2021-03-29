#include "vips.h"
#include <string.h>

#define VIPS_SUPPORT_ARRAY_HEADERS \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 9))

#define VIPS_SUPPORT_AVIF \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 9))

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
  switch (imgtype)
  {
  case (JPEG):
    return vips_type_find("VipsOperation", "jpegload_buffer");
  case (PNG):
    return vips_type_find("VipsOperation", "pngload_buffer");
  case (WEBP):
    return vips_type_find("VipsOperation", "webpload_buffer");
  case (GIF):
    return vips_type_find("VipsOperation", "gifload_buffer");
  case (SVG):
    return vips_type_find("VipsOperation", "svgload_buffer");
  case (HEIC):
    return vips_type_find("VipsOperation", "heifload_buffer");
  case (AVIF):
    return vips_type_find("VipsOperation", "heifload_buffer");
  case (BMP):
    return vips_type_find("VipsOperation", "magickload_buffer");
  case (TIFF):
    return vips_type_find("VipsOperation", "tiffload_buffer");
  }
  return 0;
}

int
vips_type_find_save_go(int imgtype) {
  switch (imgtype)
  {
  case (JPEG):
    return vips_type_find("VipsOperation", "jpegsave_buffer");
  case (PNG):
    return vips_type_find("VipsOperation", "pngsave_buffer");
  case (WEBP):
    return vips_type_find("VipsOperation", "webpsave_buffer");
  case (GIF):
    return vips_type_find("VipsOperation", "magicksave_buffer");
#if VIPS_SUPPORT_AVIF
  case (AVIF):
    return vips_type_find("VipsOperation", "heifsave_buffer");
#endif
  case (ICO):
    return vips_type_find("VipsOperation", "pngsave_buffer");
  case (BMP):
    return vips_type_find("VipsOperation", "magicksave_buffer");
  case (TIFF):
    return vips_type_find("VipsOperation", "tiffsave_buffer");
  }

  return 0;
}

int
vips_jpegload_go(void *buf, size_t len, int shrink, VipsImage **out) {
  if (shrink > 1)
    return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "shrink", shrink, NULL);

  return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_pngload_go(void *buf, size_t len, VipsImage **out) {
  return vips_pngload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_webpload_go(void *buf, size_t len, double scale, int pages, VipsImage **out) {
  return vips_webpload_buffer(
    buf, len, out,
    "access", VIPS_ACCESS_SEQUENTIAL,
    "scale", scale,
    "n", pages,
    NULL
  );
}

int
vips_gifload_go(void *buf, size_t len, int pages, VipsImage **out) {
  return vips_gifload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "n", pages, NULL);
}

int
vips_svgload_go(void *buf, size_t len, double scale, VipsImage **out) {
  // libvips limits the minimal scale to 0.001, so we have to scale down dpi
  // for lower scale values
  double dpi = 72.0;
  if (scale < 0.001) {
    dpi *= VIPS_MAX(scale / 0.001, 0.001);
    scale = 0.001;
  }

  return vips_svgload_buffer(
    buf, len, out,
    "access", VIPS_ACCESS_SEQUENTIAL,
    "scale", scale,
    "dpi", dpi,
    NULL
  );
}

int
vips_heifload_go(void *buf, size_t len, VipsImage **out) {
  return vips_heifload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_bmpload_go(void *buf, size_t len, VipsImage **out) {
  return vips_magickload_buffer(buf, len, out, NULL);
}

int
vips_tiffload_go(void *buf, size_t len, VipsImage **out) {
  return vips_tiffload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_get_orientation(VipsImage *image) {
  int orientation;

	if (
    vips_image_get_typeof(image, VIPS_META_ORIENTATION) == G_TYPE_INT &&
    vips_image_get_int(image, VIPS_META_ORIENTATION, &orientation) == 0
  ) return orientation;

	return 1;
}

VipsBandFormat
vips_band_format(VipsImage *in) {
  return in->BandFmt;
}

gboolean
vips_is_animated(VipsImage * in) {
  return( vips_image_get_typeof(in, "page-height") != G_TYPE_INVALID &&
          vips_image_get_typeof(in, "gif-delay") != G_TYPE_INVALID &&
          vips_image_get_typeof(in, "gif-loop") != G_TYPE_INVALID );
}

int
vips_image_get_array_int_go(VipsImage *image, const char *name, int **out, int *n) {
#if VIPS_SUPPORT_ARRAY_HEADERS
  return vips_image_get_array_int(image, name, out, n);
#else
  vips_error("vips_image_get_array_int_go", "Array headers are not supported (libvips 8.9+ reuired)");
  return 1;
#endif
}

void
vips_image_set_array_int_go(VipsImage *image, const char *name, const int *array, int n) {
#if VIPS_SUPPORT_ARRAY_HEADERS
  vips_image_set_array_int(image, name, array, n);
#endif
}

gboolean
vips_image_hasalpha_go(VipsImage * in) {
  return vips_image_hasalpha(in);
}

int
vips_addalpha_go(VipsImage *in, VipsImage **out) {
  return vips_addalpha(in, out, NULL);
}

int
vips_copy_go(VipsImage *in, VipsImage **out) {
  return vips_copy(in, out, NULL);
}

int
vips_cast_go(VipsImage *in, VipsImage **out, VipsBandFormat format) {
  return vips_cast(in, out, format, NULL);
}

int
vips_rad2float_go(VipsImage *in, VipsImage **out) {
	return vips_rad2float(in, out, NULL);
}

int
vips_resize_go(VipsImage *in, VipsImage **out, double wscale, double hscale) {
  return vips_resize(in, out, wscale, "vscale", hscale, NULL);
}

int
vips_resize_with_premultiply(VipsImage *in, VipsImage **out, double wscale, double hscale) {
	VipsBandFormat format;
  VipsImage *tmp1, *tmp2;

  format = vips_band_format(in);

  if (vips_premultiply(in, &tmp1, NULL))
    return 1;

	if (vips_resize(tmp1, &tmp2, wscale, "vscale", hscale, NULL)) {
    clear_image(&tmp1);
		return 1;
  }
  swap_and_clear(&tmp1, tmp2);

  if (vips_unpremultiply(tmp1, &tmp2, NULL)) {
    clear_image(&tmp1);
		return 1;
  }
  swap_and_clear(&tmp1, tmp2);

  if (vips_cast(tmp1, out, format, NULL)) {
    clear_image(&tmp1);
		return 1;
  }

  clear_image(&tmp1);

  return 0;
}

int
vips_icc_is_srgb_iec61966(VipsImage *in) {
  const void *data;
  size_t data_len;

  // 1998-12-01
  static char date[] = { 7, 206, 0, 2, 0, 9 };
  // 2.1
  static char version[] = { 2, 16, 0, 0 };

  if (vips_image_get_blob(in, VIPS_META_ICC_NAME, &data, &data_len))
    return FALSE;

  // Less than header size
  if (data_len < 128)
    return FALSE;

  // Predict it is sRGB IEC61966 2.1 by checking some header fields
  return ((memcmp(data + 48, "IEC ",  4) == 0) && // Device manufacturer
          (memcmp(data + 52, "sRGB",  4) == 0) && // Device model
          (memcmp(data + 80, "HP  ",  4) == 0) && // Profile creator
          (memcmp(data + 24, date,    6) == 0) && // Date of creation
          (memcmp(data + 8,  version, 4) == 0));  // Version
}

int
vips_has_embedded_icc(VipsImage *in) {
  return vips_image_get_typeof(in, VIPS_META_ICC_NAME) != 0;
}

int
vips_icc_import_go(VipsImage *in, VipsImage **out) {
  return vips_icc_import(in, out, "embedded", TRUE, "pcs", VIPS_PCS_XYZ, NULL);
}

int
vips_icc_export_go(VipsImage *in, VipsImage **out) {
  return vips_icc_export(in, out, NULL);
}

int
vips_icc_export_srgb(VipsImage *in, VipsImage **out) {
  return vips_icc_export(in, out, "output_profile", "sRGB", NULL);
}

int
vips_icc_transform_go(VipsImage *in, VipsImage **out) {
  return vips_icc_transform(in, out, "sRGB", "embedded", TRUE, "pcs", VIPS_PCS_XYZ, NULL);
}

int
vips_icc_remove(VipsImage *in, VipsImage **out) {
  if (vips_copy(in, out, NULL)) return 1;

  vips_image_remove(*out, VIPS_META_ICC_NAME);

  return 0;
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
  return vips_smartcrop(in, out, width, height, NULL);
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
vips_trim(VipsImage *in, VipsImage **out, double threshold,
          gboolean smart, double r, double g, double b,
          gboolean equal_hor, gboolean equal_ver) {
  VipsImage *tmp;

  if (vips_image_hasalpha(in)) {
    if (vips_flatten(in, &tmp, NULL))
      return 1;
  } else {
    if (vips_copy(in, &tmp, NULL))
      return 1;
  }

  double *bg;
  int bgn;
  VipsArrayDouble *bga;

  if (smart) {
    if (vips_getpoint(tmp, &bg, &bgn, 0, 0, NULL)) {
      clear_image(&tmp);
      return 1;
    }
    bga = vips_array_double_new(bg, bgn);
  } else {
    bga = vips_array_double_newv(3, r, g, b);
    bg = 0;
  }

  int left, right, top, bot, width, height, diff;

  if (vips_find_trim(tmp, &left, &top, &width, &height, "background", bga, "threshold", threshold, NULL)) {
    clear_image(&tmp);
    vips_area_unref((VipsArea *)bga);
    g_free(bg);
    return 1;
  }

  if (equal_hor) {
    right = in->Xsize - left - width;
    diff = right - left;
    if (diff > 0) {
      width += diff;
    } else if (diff < 0) {
      left = right;
      width -= diff;
    }
  }

  if (equal_ver) {
    bot = in->Ysize - top - height;
    diff = bot - top;
    if (diff > 0) {
      height += diff;
    } else if (diff < 0) {
      top = bot;
      height -= diff;
    }
  }

  clear_image(&tmp);
  vips_area_unref((VipsArea *)bga);
  g_free(bg);

  if (width == 0 || height == 0) {
    return vips_copy(in, out, NULL);
  }

  return vips_extract_area(in, out, left, top, width, height, NULL);
}

int
vips_replicate_go(VipsImage *in, VipsImage **out, int width, int height) {
  VipsImage *tmp;

  if (vips_replicate(in, &tmp, 1 + width / in->Xsize, 1 + height / in->Ysize, NULL))
    return 1;

	if (vips_extract_area(tmp, out, 0, 0, width, height, NULL)) {
    clear_image(&tmp);
		return 1;
  }

  clear_image(&tmp);

  return 0;
}

int
vips_embed_go(VipsImage *in, VipsImage **out, int x, int y, int width, int height, double *bg, int bgn) {
  VipsArrayDouble *bga = vips_array_double_new(bg, bgn);
  int ret = vips_embed(
    in, out, x, y, width, height,
    "extend", VIPS_EXTEND_BACKGROUND,
    "background", bga,
    NULL
  );
  vips_area_unref((VipsArea *)bga);
  return ret;
}

int
vips_ensure_alpha(VipsImage *in, VipsImage **out) {
  if (vips_image_hasalpha_go(in)) {
    return vips_copy(in, out, NULL);
  }

  return vips_bandjoin_const1(in, out, 255, NULL);
}

int
vips_apply_watermark(VipsImage *in, VipsImage *watermark, VipsImage **out, double opacity) {
  VipsImage *base = vips_image_new();
	VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 5);

	if (opacity < 1) {
    if (
      vips_extract_band(watermark, &t[0], 0, "n", watermark->Bands - 1, NULL) ||
      vips_extract_band(watermark, &t[1], watermark->Bands - 1, "n", 1, NULL) ||
		  vips_linear1(t[1], &t[2], opacity, 0, NULL) ||
      vips_bandjoin2(t[0], t[2], &t[3], NULL)
    ) {
      clear_image(&base);
			return 1;
		}
	} else {
    if (vips_copy(watermark, &t[3], NULL)) {
      clear_image(&base);
      return 1;
    }
  }

  int res =
    vips_composite2(in, t[3], &t[4], VIPS_BLEND_MODE_OVER, "compositing_space", in->Type, NULL) ||
    vips_cast(t[4], out, vips_image_get_format(in), NULL);

  clear_image(&base);

  return res;
}

int
vips_arrayjoin_go(VipsImage **in, VipsImage **out, int n) {
  return vips_arrayjoin(in, out, n, "across", 1, NULL);
}

int
vips_strip(VipsImage *in, VipsImage **out) {
  if (vips_copy(in, out, NULL)) return 1;

  gchar **fields = vips_image_get_fields(in);

  for (int i = 0; fields[i] != NULL; i++) {
    gchar *name = fields[i];

    if (strcmp(name, VIPS_META_ICC_NAME) == 0) continue;

    vips_image_remove(*out, name);
  }

  g_strfreev(fields);

  return 0;
}

int
vips_jpegsave_go(VipsImage *in, void **buf, size_t *len, int quality, int interlace) {
  return vips_jpegsave_buffer(
    in, buf, len,
    "Q", quality,
    "optimize_coding", TRUE,
    "interlace", interlace,
    NULL
  );
}

int
vips_pngsave_go(VipsImage *in, void **buf, size_t *len, int interlace, int quantize, int colors) {
  return vips_pngsave_buffer(
    in, buf, len,
    "filter", VIPS_FOREIGN_PNG_FILTER_NONE,
    "interlace", interlace,
    "palette", quantize,
    "colours", colors,
    NULL);
}

int
vips_webpsave_go(VipsImage *in, void **buf, size_t *len, int quality) {
  return vips_webpsave_buffer(
    in, buf, len,
    "Q", quality,
    NULL
  );
}

int
vips_gifsave_go(VipsImage *in, void **buf, size_t *len) {
  return vips_magicksave_buffer(in, buf, len, "format", "gif", NULL);
}

int
vips_tiffsave_go(VipsImage *in, void **buf, size_t *len, int quality) {
  return vips_tiffsave_buffer(in, buf, len, "Q", quality, NULL);
}

int
vips_avifsave_go(VipsImage *in, void **buf, size_t *len, int quality) {
#if VIPS_SUPPORT_AVIF
  return vips_heifsave_buffer(in, buf, len, "Q", quality, "compression", VIPS_FOREIGN_HEIF_COMPRESSION_AV1, NULL);
#else
  vips_error("vips_avifsave_go", "Saving AVIF is not supported (libvips 8.9+ reuired)");
  return 1;
#endif
}

int
vips_bmpsave_go(VipsImage *in, void **buf, size_t *len) {
  return vips_magicksave_buffer(in, buf, len, "format", "bmp", NULL);
}

void
vips_cleanup() {
  vips_error_clear();
  vips_thread_shutdown();
}
