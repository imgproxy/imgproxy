#include "vips.h"

#define VIPS_SUPPORT_SMARTCROP \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 5))

#define VIPS_SUPPORT_HASALPHA \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 5))

#define VIPS_SUPPORT_GIF \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 3))

#define VIPS_SUPPORT_SVG \
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
  case (ICO):
    return vips_type_find("VipsOperation", "magicksave_buffer");
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
vips_webpload_go(void *buf, size_t len, int shrink, VipsImage **out) {
  if (shrink > 1)
    return vips_webpload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "shrink", shrink, NULL);

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
vips_svgload_go(void *buf, size_t len, double scale, VipsImage **out) {
  #if VIPS_SUPPORT_SVG
    return vips_svgload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "scale", scale, NULL);
  #else
    vips_error("vips_svgload_go", "Loading SVG is not supported");
    return 1;
  #endif
}

int
vips_get_exif_orientation(VipsImage *image) {
	const char *orientation;

	if (
		vips_image_get_typeof(image, EXIF_ORIENTATION) != G_TYPE_INVALID &&
		vips_image_get_string(image, EXIF_ORIENTATION, &orientation) == 0
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

VipsBandFormat
vips_band_format(VipsImage *in) {
  return in->BandFmt;
}

gboolean
vips_is_animated_gif(VipsImage * in) {
  return( vips_image_get_typeof(in, "page-height") != G_TYPE_INVALID &&
          vips_image_get_typeof(in, "gif-delay") != G_TYPE_INVALID &&
          vips_image_get_typeof(in, "gif-loop") != G_TYPE_INVALID );
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
vips_copy_go(VipsImage *in, VipsImage **out) {
  return vips_copy(in, out, NULL);
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
vips_resize_with_premultiply(VipsImage *in, VipsImage **out, double scale) {
	VipsBandFormat format;
  VipsImage *tmp1, *tmp2;

  format = vips_band_format(in);

  if (vips_premultiply(in, &tmp1, NULL))
    return 1;

	if (vips_resize(tmp1, &tmp2, scale, NULL)) {
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
vips_embed_go(VipsImage *in, VipsImage **out, int x, int y, int width, int height) {
  return vips_embed(in, out, x, y, width, height, NULL);
}

int
vips_ensure_alpha(VipsImage *in, VipsImage **out) {
  if (vips_image_hasalpha_go(in)) {
    return vips_copy(in, out, NULL);
  }

  return vips_bandjoin_const1(in, out, 255, NULL);
}

int
vips_apply_opacity(VipsImage *in, VipsImage **out, double opacity){
  if (vips_image_hasalpha_go(in)) {
    if (opacity < 1) {
      VipsImage *img, *img_alpha, *tmp;

			if (vips_extract_band(in, &img, 0, "n", in->Bands - 1, NULL))
				return 1;

      if (vips_extract_band(in, &img_alpha, in->Bands - 1, "n", 1, NULL)) {
        clear_image(&img);
				return 1;
			}

			if (vips_linear1(img_alpha, &tmp, opacity, 0, NULL)) {
        clear_image(&img);
        clear_image(&img_alpha);
				return 1;
			}
			swap_and_clear(&img_alpha, tmp);

			if (vips_bandjoin2(img, img_alpha, out, NULL)) {
        clear_image(&img);
        clear_image(&img_alpha);
				return 1;
			}

      clear_image(&img);
      clear_image(&img_alpha);
    } else {
      if (vips_copy(in, out, NULL)) {
        return 1;
      }
    }
  } else {
    if (vips_bandjoin_const1(in, out, opacity * 255, NULL)) {
      return 1;
    }
  }

  return 0;
}

int
vips_apply_watermark(VipsImage *in, VipsImage *watermark, VipsImage **out, double opacity) {
  VipsImage *wm, *wm_alpha, *tmp;

	if (vips_extract_band(watermark, &wm, 0, "n", watermark->Bands - 1, NULL))
		return 1;

  if (vips_extract_band(watermark, &wm_alpha, watermark->Bands - 1, "n", 1, NULL)) {
    clear_image(&wm);
		return 1;
	}

	VipsInterpretation img_interpolation = vips_image_guess_interpretation(in);

	if (img_interpolation != vips_image_guess_interpretation(wm)) {
		if (vips_colourspace(wm, &tmp, img_interpolation, NULL)) {
      clear_image(&wm);
      clear_image(&wm_alpha);
			return 1;
		}
		swap_and_clear(&wm, tmp);
	}

	if (opacity < 1) {
		if (vips_linear1(wm_alpha, &tmp, opacity, 0, NULL)) {
      clear_image(&wm);
      clear_image(&wm_alpha);
			return 1;
		}

		swap_and_clear(&wm_alpha, tmp);
	}

	VipsBandFormat img_format;
	VipsImage *img, *img_alpha;

	img_format = vips_image_get_format(in);

  gboolean has_alpha = vips_image_hasalpha_go(in);

	if (has_alpha) {
		if (vips_extract_band(in, &img, 0, "n", in->Bands - 1, NULL)) {
      clear_image(&wm);
      clear_image(&wm_alpha);
			return 1;
		}

		if (vips_extract_band(in, &img_alpha, in->Bands - 1, "n", 1, NULL)) {
      clear_image(&wm);
      clear_image(&wm_alpha);
      clear_image(&img);
			return 1;
		}
	} else {
    if (vips_copy(in, &img, NULL)) {
      clear_image(&wm);
      clear_image(&wm_alpha);
			return 1;
    }
  }

	if (vips_ifthenelse(wm_alpha, wm, img, &tmp, "blend", TRUE, NULL)) {
    clear_image(&wm);
    clear_image(&wm_alpha);
    clear_image(&img);
    clear_image(&img_alpha);
		return 1;
	}

	swap_and_clear(&img, tmp);
  clear_image(&wm);
  clear_image(&wm_alpha);

	if (has_alpha) {
		if (vips_bandjoin2(img, img_alpha, &tmp, NULL)) {
      clear_image(&img);
      clear_image(&img_alpha);
			return 1;
		}

		swap_and_clear(&img, tmp);
    clear_image(&img_alpha);
	}

	if (img_format != vips_image_get_format(img)) {
		if (vips_cast(img, &tmp, img_format, NULL)) {
      clear_image(&img);
			return 1;
		}
		swap_and_clear(&img, tmp);
	}

  *out = img;

  return 0;
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
vips_pngsave_go(VipsImage *in, void **buf, size_t *len, int interlace, int embed_profile) {
  if (embed_profile)
    return vips_pngsave_buffer(in, buf, len, "filter", VIPS_FOREIGN_PNG_FILTER_NONE, "interlace", interlace, NULL);

  return vips_pngsave_buffer(in, buf, len, "profile", "none", "filter", VIPS_FOREIGN_PNG_FILTER_NONE, "interlace", interlace, NULL);
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

int
vips_icosave_go(VipsImage *in, void **buf, size_t *len) {
#if VIPS_SUPPORT_MAGICK
  return vips_magicksave_buffer(in, buf, len, "format", "ico", NULL);
#else
  vips_error("vips_icosave_go", "Saving ICO is not supported");
  return 1;
#endif
}

void
vips_cleanup() {
  vips_error_clear();
  vips_thread_shutdown();
}
