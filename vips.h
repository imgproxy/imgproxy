#include <stdlib.h>

#include <vips/vips.h>
#include <vips/vips7compat.h>
#include <vips/vector.h>

enum ImgproxyImageTypes {
  UNKNOWN = 0,
  JPEG,
  PNG,
  WEBP,
  GIF,
  ICO,
  SVG,
  HEIC,
  BMP,
  TIFF
};

int vips_initialize();

void clear_image(VipsImage **in);
void g_free_go(void **buf);

void swap_and_clear(VipsImage **in, VipsImage *out);

int vips_type_find_load_go(int imgtype);
int vips_type_find_save_go(int imgtype);

int vips_jpegload_go(void *buf, size_t len, int shrink, VipsImage **out);
int vips_pngload_go(void *buf, size_t len, VipsImage **out);
int vips_webpload_go(void *buf, size_t len, double scale, int pages, VipsImage **out);
int vips_gifload_go(void *buf, size_t len, int pages, VipsImage **out);
int vips_svgload_go(void *buf, size_t len, double scale, VipsImage **out);
int vips_heifload_go(void *buf, size_t len, VipsImage **out);
int vips_bmpload_go(void *buf, size_t len, VipsImage **out);
int vips_tiffload_go(void *buf, size_t len, VipsImage **out);

int vips_get_orientation(VipsImage *image);
void vips_strip_meta(VipsImage *image);

int vips_support_smartcrop();

VipsBandFormat vips_band_format(VipsImage *in);

gboolean vips_support_webp_animation();
gboolean vips_is_animated(VipsImage * in);
gboolean vips_image_hasalpha_go(VipsImage * in);

int vips_copy_go(VipsImage *in, VipsImage **out);

int vips_cast_go(VipsImage *in, VipsImage **out, VipsBandFormat format);
int vips_rad2float_go(VipsImage *in, VipsImage **out);

int vips_resize_go(VipsImage *in, VipsImage **out, double scale);
int vips_resize_with_premultiply(VipsImage *in, VipsImage **out, double scale);

int vips_icc_is_srgb_iec61966(VipsImage *in);
int vips_has_embedded_icc(VipsImage *in);
int vips_support_builtin_icc();
int vips_icc_import_go(VipsImage *in, VipsImage **out, char *profile);
int vips_colourspace_go(VipsImage *in, VipsImage **out, VipsInterpretation cs);

int vips_rot_go(VipsImage *in, VipsImage **out, VipsAngle angle);
int vips_flip_horizontal_go(VipsImage *in, VipsImage **out);

int vips_extract_area_go(VipsImage *in, VipsImage **out, int left, int top, int width, int height);
int vips_smartcrop_go(VipsImage *in, VipsImage **out, int width, int height);
int vips_trim(VipsImage *in, VipsImage **out, double threshold,
              gboolean smart, double r, double g, double b,
              gboolean equal_hor, gboolean equal_ver);

int vips_gaussblur_go(VipsImage *in, VipsImage **out, double sigma);
int vips_sharpen_go(VipsImage *in, VipsImage **out, double sigma);

int vips_flatten_go(VipsImage *in, VipsImage **out, double r, double g, double b);

int vips_replicate_go(VipsImage *in, VipsImage **out, int across, int down);
int vips_embed_go(VipsImage *in, VipsImage **out, int x, int y, int width, int height, double *bg, int bgn);

int vips_ensure_alpha(VipsImage *in, VipsImage **out);

int vips_apply_watermark(VipsImage *in, VipsImage *watermark, VipsImage **out, double opacity);

int vips_arrayjoin_go(VipsImage **in, VipsImage **out, int n);

int vips_jpegsave_go(VipsImage *in, void **buf, size_t *len, int quality, int interlace, gboolean strip);
int vips_pngsave_go(VipsImage *in, void **buf, size_t *len, int interlace, int quantize, int colors);
int vips_webpsave_go(VipsImage *in, void **buf, size_t *len, int quality, gboolean strip);
int vips_gifsave_go(VipsImage *in, void **buf, size_t *len);
int vips_icosave_go(VipsImage *in, void **buf, size_t *len);
int vips_bmpsave_go(VipsImage *in, void **buf, size_t *len);
int vips_tiffsave_go(VipsImage *in, void **buf, size_t *len, int quality);

void vips_cleanup();
