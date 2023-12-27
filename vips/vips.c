#include "vips.h"
#include <string.h>

#define VIPS_SCRGB_ALPHA_FIXED \
  (VIPS_MAJOR_VERSION > 8 || (VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION >= 15))

#define VIPS_META_PALETTE_BITS_DEPTH "palette-bit-depth"

int
vips_initialize()
{
  return vips_init("imgproxy");
}

void
clear_image(VipsImage **in)
{
  if (G_IS_OBJECT(*in))
    g_clear_object(in);
}

void
g_free_go(void **buf)
{
  g_free(*buf);
}

void
swap_and_clear(VipsImage **in, VipsImage *out)
{
  clear_image(in);
  *in = out;
}

int
gif_resolution_limit()
{
  return INT_MAX / 4;
}

// Just create and destroy a tiny image to ensure vips is operational
int
vips_health()
{
  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 2);

  int res = vips_black(&t[0], 4, 4, "bands", 4, NULL) ||
      !(t[1] = vips_image_copy_memory(t[0]));

  clear_image(&base);

  return res;
}

int
vips_jpegload_go(void *buf, size_t len, int shrink, VipsImage **out)
{
  if (shrink > 1)
    return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "shrink", shrink,
        NULL);

  return vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_pngload_go(void *buf, size_t len, VipsImage **out)
{
  return vips_pngload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_webpload_go(void *buf, size_t len, double scale, int pages, VipsImage **out)
{
  return vips_webpload_buffer(
      buf, len, out,
      "access", VIPS_ACCESS_SEQUENTIAL,
      "scale", scale,
      "n", pages,
      NULL);
}

int
vips_gifload_go(void *buf, size_t len, int pages, VipsImage **out)
{
  return vips_gifload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, "n", pages, NULL);
}

int
vips_svgload_go(void *buf, size_t len, double scale, VipsImage **out)
{
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
      NULL);
}

int
vips_heifload_go(void *buf, size_t len, VipsImage **out, int thumbnail)
{
  return vips_heifload_buffer(
      buf, len, out,
      "access", VIPS_ACCESS_SEQUENTIAL,
      "thumbnail", thumbnail,
      NULL);
}

int
vips_tiffload_go(void *buf, size_t len, VipsImage **out)
{
  return vips_tiffload_buffer(buf, len, out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

int
vips_black_go(VipsImage **out, int width, int height, int bands)
{
  VipsImage *tmp;

  int res = vips_black(&tmp, width, height, "bands", bands, NULL) ||
      vips_copy(tmp, out, "interpretation", VIPS_INTERPRETATION_sRGB, NULL);

  clear_image(&tmp);

  return res;
}

int
vips_fix_scRGB_alpha_tiff(VipsImage *in, VipsImage **out)
{
#if VIPS_SCRGB_ALPHA_FIXED
  /* Vips 8.15+ uses 0.0-1.0 range for linear alpha, so we don't need a fix.
   */
  return vips_copy(in, out, NULL);
#else
  /* Vips prior to 8.14 loads linear alpha in the 0.0-1.0 range but uses the 0.0-255.0 range.
   */
  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 4);

  int res =
      vips_extract_band(in, &t[0], 0, "n", 3, NULL) ||
      vips_extract_band(in, &t[1], 3, "n", in->Bands - 3, NULL) ||
      vips_linear1(t[1], &t[2], 255.0, 0, NULL) ||
      vips_cast(t[2], &t[3], in->BandFmt, NULL) ||
      vips_bandjoin2(t[0], t[3], out, NULL);

  clear_image(&base);

  return res;
#endif
}

/* Vips loads linear BW TIFFs as VIPS_INTERPRETATION_B_W or VIPS_INTERPRETATION_GREY16
 * but these colourspaces are not linear. We should properly convert them to
 * VIPS_INTERPRETATION_GREY16
 */
int
vips_fix_BW_float_tiff(VipsImage *in, VipsImage **out)
{
  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 8);

  VipsImage *color = in;
  VipsImage *alpha = NULL;

  /* Extract and fix alpha. Float WB TIFF uses the 0.0-1.0 range but we need
   * the 0.0-65535.0 range
   */
  if (in->Bands > 1) {
    if (
        vips_extract_band(in, &t[0], 0, NULL) ||
        vips_extract_band(in, &t[1], 1, "n", in->Bands - 1, NULL) ||
        vips_linear1(t[1], &t[2], 65535.0, 0, NULL) ||
        vips_cast_ushort(t[2], &t[3], NULL) ||
        vips_copy(t[3], &t[4], "interpretation", VIPS_INTERPRETATION_GREY16, NULL)) {
      clear_image(&base);
      return 1;
    }

    color = t[0];
    alpha = t[4];
  }

  /* Craft an scRGB image and convert it back to GREY16 to apply a gamma
   * correction
   */
  VipsImage *rgb[3] = { color, color, color };
  if (
      vips_bandjoin(rgb, &t[5], 3, NULL) ||
      vips_colourspace(t[5], &t[6], VIPS_INTERPRETATION_GREY16,
          "source_space", VIPS_INTERPRETATION_scRGB, NULL)) {
    clear_image(&base);
    return 1;
  }

  int res;

  if (alpha)
    res =
        vips_bandjoin2(t[6], alpha, &t[7], NULL) ||
        vips_icc_remove(t[7], out);
  else
    res = vips_icc_remove(t[6], out);

  clear_image(&base);

  return res;
}

int
vips_fix_float_tiff(VipsImage *in, VipsImage **out)
{
  /* Vips loads linear alpha in the 0.0-1.0 range but uses the 0.0-255.0 range.
   * https://github.com/libvips/libvips/pull/3627 fixes this behavior
   */
  if (in->Type == VIPS_INTERPRETATION_scRGB && in->Bands > 3)
    return vips_fix_scRGB_alpha_tiff(in, out);

  /* Vips loads linear BW TIFFs as VIPS_INTERPRETATION_B_W or VIPS_INTERPRETATION_GREY16
   * but these colourspaces are not linear. We should properly convert them to
   * VIPS_INTERPRETATION_GREY16
   */
  if (
      (in->Type == VIPS_INTERPRETATION_B_W || in->Type == VIPS_INTERPRETATION_GREY16) &&
      (in->BandFmt == VIPS_FORMAT_FLOAT || in->BandFmt == VIPS_FORMAT_DOUBLE))
    return vips_fix_BW_float_tiff(in, out);

  return vips_copy(in, out, NULL);
}

int
vips_get_orientation(VipsImage *image)
{
  int orientation;

  if (
      vips_image_get_typeof(image, VIPS_META_ORIENTATION) == G_TYPE_INT &&
      vips_image_get_int(image, VIPS_META_ORIENTATION, &orientation) == 0)
    return orientation;

  return 1;
}

int
vips_get_palette_bit_depth(VipsImage *image)
{
  int palette, palette_bit_depth;

#ifdef VIPS_META_PALETTE
  if (vips_image_get_typeof(image, VIPS_META_PALETTE) == G_TYPE_INT &&
      vips_image_get_int(image, VIPS_META_PALETTE, &palette) == 0 &&
      palette) {

    if (vips_image_get_typeof(image, VIPS_META_BITS_PER_SAMPLE) == G_TYPE_INT &&
        vips_image_get_int(image, VIPS_META_BITS_PER_SAMPLE, &palette_bit_depth) == 0)
      return palette_bit_depth;

    else
      /* Image has palette but VIPS_META_BITS_PER_SAMPLE is not set.
       * It's very unlikely but we should handle this
       */
      return 8;
  }
#else
  if (vips_image_get_typeof(image, VIPS_META_PALETTE_BITS_DEPTH) == G_TYPE_INT &&
      vips_image_get_int(image, VIPS_META_PALETTE_BITS_DEPTH, &palette_bit_depth) == 0)
    return palette_bit_depth;
#endif

  return 0;
}

void
vips_remove_palette_bit_depth(VipsImage *image)
{
  vips_image_remove(image, VIPS_META_PALETTE_BITS_DEPTH);
#ifdef VIPS_META_PALETTE
  vips_image_remove(image, VIPS_META_PALETTE);
#endif
}

VipsBandFormat
vips_band_format(VipsImage *in)
{
  return in->BandFmt;
}

gboolean
vips_is_animated(VipsImage *in)
{
  int n_pages;

  return (vips_image_get_typeof(in, "delay") != G_TYPE_INVALID &&
      vips_image_get_typeof(in, "loop") != G_TYPE_INVALID &&
      vips_image_get_typeof(in, "page-height") == G_TYPE_INT &&
      vips_image_get_typeof(in, "n-pages") == G_TYPE_INT &&
      vips_image_get_int(in, "n-pages", &n_pages) == 0 &&
      n_pages > 1);
}

int
vips_image_get_array_int_go(VipsImage *image, const char *name, int **out, int *n)
{
  return vips_image_get_array_int(image, name, out, n);
}

void
vips_image_set_array_int_go(VipsImage *image, const char *name, const int *array, int n)
{
  vips_image_set_array_int(image, name, array, n);
}

int
vips_addalpha_go(VipsImage *in, VipsImage **out)
{
  return vips_addalpha(in, out, NULL);
}

int
vips_copy_go(VipsImage *in, VipsImage **out)
{
  return vips_copy(in, out, NULL);
}

int
vips_cast_go(VipsImage *in, VipsImage **out, VipsBandFormat format)
{
  return vips_cast(in, out, format, NULL);
}

int
vips_rad2float_go(VipsImage *in, VipsImage **out)
{
  return vips_rad2float(in, out, NULL);
}

int
vips_resize_go(VipsImage *in, VipsImage **out, double wscale, double hscale)
{
  if (!vips_image_hasalpha(in))
    return vips_resize(in, out, wscale, "vscale", hscale, NULL);

  VipsBandFormat format = vips_band_format(in);

  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 4);

  int res =
      vips_premultiply(in, &t[0], NULL) ||
      vips_cast(t[0], &t[1], format, NULL) ||
      vips_resize(t[1], &t[2], wscale, "vscale", hscale, NULL) ||
      vips_unpremultiply(t[2], &t[3], NULL) ||
      vips_cast(t[3], out, format, NULL);

  clear_image(&base);

  return res;
}

int
vips_icc_is_srgb_iec61966(VipsImage *in)
{
  const void *data;
  size_t data_len;

  // 1998-12-01
  static char date[] = { 7, 206, 0, 2, 0, 9 };
  // 2.1
  static char version[] = { 2, 16, 0, 0 };

  if (vips_image_get_blob(in, VIPS_META_ICC_NAME, &data, &data_len))
    return FALSE;

  /* Less than header size
   */
  if (data_len < 128)
    return FALSE;

  /* Predict it is sRGB IEC61966 2.1 by checking some header fields
   */
  return ((memcmp(data + 48, "IEC ", 4) == 0) && // Device manufacturer
      (memcmp(data + 52, "sRGB", 4) == 0) &&     // Device model
      (memcmp(data + 80, "HP  ", 4) == 0) &&     // Profile creator
      (memcmp(data + 24, date, 6) == 0) &&       // Date of creation
      (memcmp(data + 8, version, 4) == 0));      // Version
}

int
vips_has_embedded_icc(VipsImage *in)
{
  return vips_image_get_typeof(in, VIPS_META_ICC_NAME) != 0;
}

int
vips_icc_import_go(VipsImage *in, VipsImage **out)
{
  return vips_icc_import(in, out, "embedded", TRUE, "pcs", VIPS_PCS_LAB, NULL);
}

int
vips_icc_export_go(VipsImage *in, VipsImage **out)
{
  return vips_icc_export(in, out, "pcs", VIPS_PCS_LAB, NULL);
}

int
vips_icc_export_srgb(VipsImage *in, VipsImage **out)
{
  return vips_icc_export(in, out, "output_profile", "sRGB", "pcs", VIPS_PCS_LAB, NULL);
}

int
vips_icc_transform_go(VipsImage *in, VipsImage **out)
{
  return vips_icc_transform(in, out, "sRGB", "embedded", TRUE, "pcs", VIPS_PCS_LAB, NULL);
}

int
vips_icc_remove(VipsImage *in, VipsImage **out)
{
  if (vips_copy(in, out, NULL))
    return 1;

  vips_image_remove(*out, VIPS_META_ICC_NAME);
  vips_image_remove(*out, "exif-ifd0-WhitePoint");
  vips_image_remove(*out, "exif-ifd0-PrimaryChromaticities");
  vips_image_remove(*out, "exif-ifd2-ColorSpace");

  return 0;
}

int
vips_colourspace_go(VipsImage *in, VipsImage **out, VipsInterpretation cs)
{
  return vips_colourspace(in, out, cs, NULL);
}

int
vips_rot_go(VipsImage *in, VipsImage **out, VipsAngle angle)
{
  return vips_rot(in, out, angle, NULL);
}

int
vips_flip_horizontal_go(VipsImage *in, VipsImage **out)
{
  return vips_flip(in, out, VIPS_DIRECTION_HORIZONTAL, NULL);
}

int
vips_smartcrop_go(VipsImage *in, VipsImage **out, int width, int height)
{
  return vips_smartcrop(in, out, width, height, NULL);
}

int
vips_apply_filters(VipsImage *in, VipsImage **out, double blur_sigma,
    double sharp_sigma, int pixelate_pixels)
{

  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 10);

  VipsInterpretation interpretation = in->Type;
  VipsBandFormat format = in->BandFmt;
  gboolean premultiplied = FALSE;

  if ((blur_sigma > 0 || sharp_sigma > 0) && vips_image_hasalpha(in)) {
    if (
        vips_premultiply(in, &t[0], NULL) ||
        vips_cast(t[0], &t[1], format, NULL)) {
      clear_image(&base);
      return 1;
    }

    in = t[0];
    premultiplied = TRUE;
  }

  if (blur_sigma > 0.0) {
    if (vips_gaussblur(in, &t[2], blur_sigma, NULL)) {
      clear_image(&base);
      return 1;
    }

    in = t[2];
  }

  if (sharp_sigma > 0.0) {
    if (vips_sharpen(in, &t[3], "sigma", sharp_sigma, NULL)) {
      clear_image(&base);
      return 1;
    }

    in = t[3];
  }

  pixelate_pixels = VIPS_MIN(pixelate_pixels, VIPS_MAX(in->Xsize, in->Ysize));

  if (pixelate_pixels > 1) {
    int w, h, tw, th;

    w = in->Xsize;
    h = in->Ysize;

    tw = (int) (VIPS_CEIL((double) w / pixelate_pixels)) * pixelate_pixels;
    th = (int) (VIPS_CEIL((double) h / pixelate_pixels)) * pixelate_pixels;

    if (tw > w || th > h) {
      if (vips_embed(in, &t[4], 0, 0, tw, th, "extend", VIPS_EXTEND_MIRROR, NULL)) {
        clear_image(&base);
        return 1;
      }

      in = t[4];
    }

    if (
        vips_shrink(in, &t[5], pixelate_pixels, pixelate_pixels, NULL) ||
        vips_zoom(t[5], &t[6], pixelate_pixels, pixelate_pixels, NULL)) {
      clear_image(&base);
      return 1;
    }

    in = t[6];

    if (tw > w || th > h) {
      if (vips_extract_area(in, &t[7], 0, 0, w, h, NULL)) {
        clear_image(&base);
        return 1;
      }

      in = t[7];
    }
  }

  if (premultiplied) {
    if (vips_unpremultiply(in, &t[8], NULL)) {
      clear_image(&base);
      return 1;
    }

    in = t[8];
  }

  int res =
      vips_colourspace(in, &t[9], interpretation, NULL) ||
      vips_cast(t[9], out, format, NULL);

  clear_image(&base);

  return res;
}

int
vips_flatten_go(VipsImage *in, VipsImage **out, double r, double g, double b)
{
  if (!vips_image_hasalpha(in))
    return vips_copy(in, out, NULL);

  VipsArrayDouble *bg = vips_array_double_newv(3, r, g, b);
  int res = vips_flatten(in, out, "background", bg, NULL);
  vips_area_unref((VipsArea *) bg);
  return res;
}

int
vips_extract_area_go(VipsImage *in, VipsImage **out, int left, int top, int width, int height)
{
  return vips_extract_area(in, out, left, top, width, height, NULL);
}

int
vips_trim(VipsImage *in, VipsImage **out, double threshold,
    gboolean smart, double r, double g, double b,
    gboolean equal_hor, gboolean equal_ver)
{

  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 2);

  VipsImage *tmp = in;

  if (vips_image_guess_interpretation(in) != VIPS_INTERPRETATION_sRGB) {
    if (vips_colourspace(in, &t[0], VIPS_INTERPRETATION_sRGB, NULL)) {
      clear_image(&base);
      return 1;
    }
    tmp = t[0];
  }

  if (vips_image_hasalpha(tmp)) {
    if (vips_flatten_go(tmp, &t[1], 255.0, 0, 255.0)) {
      clear_image(&base);
      return 1;
    }
    tmp = t[1];
  }

  double *bg = NULL;
  int bgn;
  VipsArrayDouble *bga;

  if (smart) {
    if (vips_getpoint(tmp, &bg, &bgn, 0, 0, NULL)) {
      clear_image(&base);
      return 1;
    }
    bga = vips_array_double_new(bg, bgn);
  }
  else {
    bga = vips_array_double_newv(3, r, g, b);
  }

  int left, right, top, bot, width, height, diff;
  int res = vips_find_trim(tmp, &left, &top, &width, &height, "background", bga, "threshold", threshold, NULL);

  clear_image(&base);
  vips_area_unref((VipsArea *) bga);
  g_free(bg);

  if (res) {
    return 1;
  }

  if (equal_hor) {
    right = in->Xsize - left - width;
    diff = right - left;
    if (diff > 0) {
      width += diff;
    }
    else if (diff < 0) {
      left = right;
      width -= diff;
    }
  }

  if (equal_ver) {
    bot = in->Ysize - top - height;
    diff = bot - top;
    if (diff > 0) {
      height += diff;
    }
    else if (diff < 0) {
      top = bot;
      height -= diff;
    }
  }

  if (width == 0 || height == 0) {
    return vips_copy(in, out, NULL);
  }

  return vips_extract_area(in, out, left, top, width, height, NULL);
}

int
vips_replicate_go(VipsImage *in, VipsImage **out, int width, int height)
{
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
vips_embed_go(VipsImage *in, VipsImage **out, int x, int y, int width, int height)
{
  VipsImage *tmp = NULL;

  if (!vips_image_hasalpha(in)) {
    if (vips_addalpha(in, &tmp, NULL))
      return 1;

    in = tmp;
  }

  int ret =
      vips_embed(in, out, x, y, width, height, "extend", VIPS_EXTEND_BLACK, NULL);

  if (tmp)
    clear_image(&tmp);

  return ret;
}

int
vips_apply_watermark(VipsImage *in, VipsImage *watermark, VipsImage **out, int left, int top, double opacity)
{
  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 7);

  if (!vips_image_hasalpha(watermark)) {
    if (vips_addalpha(watermark, &t[0], NULL))
      return 1;

    watermark = t[0];
  }

  if (opacity < 1) {
    if (
        vips_extract_band(watermark, &t[1], 0, "n", watermark->Bands - 1, NULL) ||
        vips_extract_band(watermark, &t[2], watermark->Bands - 1, "n", 1, NULL) ||
        vips_linear1(t[2], &t[3], opacity, 0, NULL) ||
        vips_bandjoin2(t[1], t[3], &t[4], NULL)) {
      clear_image(&base);
      return 1;
    }

    watermark = t[4];
  }

  int had_alpha = vips_image_hasalpha(in);

  if (
      vips_composite2(
          in, watermark, &t[5], VIPS_BLEND_MODE_OVER,
          "x", left, "y", top, "compositing_space", in->Type,
          NULL) ||
      vips_cast(t[5], &t[6], vips_image_get_format(in), NULL)) {
    clear_image(&base);
    return 1;
  }

  int res;

  if (!had_alpha && vips_image_hasalpha(t[6])) {
    res = vips_extract_band(t[6], out, 0, "n", t[6]->Bands - 1, NULL);
  }
  else {
    res = vips_copy(t[6], out, NULL);
  }

  clear_image(&base);

  return res;
}

int
vips_linecache_seq(VipsImage *in, VipsImage **out, int tile_height)
{
  return vips_linecache(in, out, "tile_height", tile_height, "access", VIPS_ACCESS_SEQUENTIAL,
      NULL);
}

int
vips_arrayjoin_go(VipsImage **in, VipsImage **out, int n)
{
  return vips_arrayjoin(in, out, n, "across", 1, NULL);
}

int
vips_strip(VipsImage *in, VipsImage **out, int keep_exif_copyright)
{
  static double default_resolution = 72.0 / 25.4;

  if (vips_copy(
          in, out,
          "xres", default_resolution,
          "yres", default_resolution,
          NULL))
    return 1;

  gchar **fields = vips_image_get_fields(in);

  for (int i = 0; fields[i] != NULL; i++) {
    gchar *name = fields[i];

    if (
        (strcmp(name, VIPS_META_ICC_NAME) == 0) ||
#ifdef VIPS_META_BITS_PER_SAMPLE
        (strcmp(name, VIPS_META_BITS_PER_SAMPLE) == 0) ||
#endif
        (strcmp(name, VIPS_META_PALETTE_BITS_DEPTH) == 0) ||
        (strcmp(name, "width") == 0) ||
        (strcmp(name, "height") == 0) ||
        (strcmp(name, "bands") == 0) ||
        (strcmp(name, "format") == 0) ||
        (strcmp(name, "coding") == 0) ||
        (strcmp(name, "interpretation") == 0) ||
        (strcmp(name, "xoffset") == 0) ||
        (strcmp(name, "yoffset") == 0) ||
        (strcmp(name, "xres") == 0) ||
        (strcmp(name, "yres") == 0) ||
        (strcmp(name, "vips-loader") == 0) ||
        (strcmp(name, "background") == 0) ||
        (strcmp(name, "vips-sequential") == 0) ||
        (strcmp(name, "imgproxy-dpr-scale") == 0))
      continue;

    if (keep_exif_copyright) {
      if (
          (strcmp(name, VIPS_META_EXIF_NAME) == 0) ||
          (strcmp(name, "exif-ifd0-Copyright") == 0) ||
          (strcmp(name, "exif-ifd0-Artist") == 0))
        continue;
    }

    vips_image_remove(*out, name);
  }

  g_strfreev(fields);

  return 0;
}

int
vips_jpegsave_go(VipsImage *in, void **buf, size_t *len, int quality, int interlace)
{
  return vips_jpegsave_buffer(
      in, buf, len,
      "Q", quality,
      "optimize_coding", TRUE,
      "interlace", interlace,
      NULL);
}

int
vips_pngsave_go(VipsImage *in, void **buf, size_t *len, int interlace, int quantize, int colors)
{
  int bitdepth;

  if (quantize) {
    bitdepth = 1;
    if (colors > 16)
      bitdepth = 8;
    else if (colors > 4)
      bitdepth = 4;
    else if (colors > 2)
      bitdepth = 2;
  }
  else {
    bitdepth = vips_get_palette_bit_depth(in);
    if (bitdepth && bitdepth <= 8) {
      if (bitdepth > 4)
        bitdepth = 8;
      else if (bitdepth > 2)
        bitdepth = 4;
      quantize = 1;
      colors = 1 << bitdepth;
    }
  }

  if (!quantize)
    return vips_pngsave_buffer(
        in, buf, len,
        "filter", VIPS_FOREIGN_PNG_FILTER_ALL,
        "interlace", interlace,
        NULL);

  return vips_pngsave_buffer(
      in, buf, len,
      "filter", VIPS_FOREIGN_PNG_FILTER_NONE,
      "interlace", interlace,
      "palette", quantize,
      "bitdepth", bitdepth,
      NULL);
}

int
vips_webpsave_go(VipsImage *in, void **buf, size_t *len, int quality)
{
  return vips_webpsave_buffer(
      in, buf, len,
      "Q", quality,
      NULL);
}

int
vips_gifsave_go(VipsImage *in, void **buf, size_t *len)
{
  int bitdepth = vips_get_palette_bit_depth(in);
  if (bitdepth <= 0 || bitdepth > 8)
    bitdepth = 8;
  return vips_gifsave_buffer(in, buf, len, "bitdepth", bitdepth, NULL);
}

int
vips_tiffsave_go(VipsImage *in, void **buf, size_t *len, int quality)
{
  return vips_tiffsave_buffer(in, buf, len, "Q", quality, NULL);
}

int
vips_heifsave_go(VipsImage *in, void **buf, size_t *len, int quality)
{
  return vips_heifsave_buffer(
      in, buf, len,
      "Q", quality,
      "compression", VIPS_FOREIGN_HEIF_COMPRESSION_HEVC,
      NULL);
}

int
vips_avifsave_go(VipsImage *in, void **buf, size_t *len, int quality, int speed)
{
  return vips_heifsave_buffer(
      in, buf, len,
      "Q", quality,
      "compression", VIPS_FOREIGN_HEIF_COMPRESSION_AV1,
      "effort", 9 - speed,
      NULL);
}

void
vips_cleanup()
{
  vips_error_clear();
  vips_thread_shutdown();
}
