#include <stdlib.h>

#include <vips/vips.h>
#include <vips/vips7compat.h>

enum ImgproxyImageTypes {
  UNKNOWN = 0,
  JPEG,
  PNG,
  WEBP,
  GIF,
  ICO,
  SVG
};

int vips_initialize();

void clear_image(VipsImage **in);
void g_free_go(void **buf);

void swap_and_clear(VipsImage **in, VipsImage *out);

int vips_type_find_load_go(int imgtype);
int vips_type_find_save_go(int imgtype);

int vips_jpegload_go(void *buf, size_t len, int shrink, VipsImage **out);
int vips_pngload_go(void *buf, size_t len, VipsImage **out);
int vips_webpload_go(void *buf, size_t len, int shrink, VipsImage **out);
int vips_gifload_go(void *buf, size_t len, int pages, VipsImage **out);
int vips_svgload_go(void *buf, size_t len, double scale, VipsImage **out);

int vips_get_exif_orientation(VipsImage *image);

int vips_support_smartcrop();

VipsBandFormat vips_band_format(VipsImage *in);

gboolean vips_is_animated_gif(VipsImage * in);
gboolean vips_image_hasalpha_go(VipsImage * in);

int vips_copy_go(VipsImage *in, VipsImage **out);

int vips_cast_go(VipsImage *in, VipsImage **out, VipsBandFormat format);

int vips_resize_go(VipsImage *in, VipsImage **out, double scale);
int vips_resize_with_premultiply(VipsImage *in, VipsImage **out, double scale);

int vips_need_icc_import(VipsImage *in);
int vips_icc_import_go(VipsImage *in, VipsImage **out, char *profile);
int vips_colourspace_go(VipsImage *in, VipsImage **out, VipsInterpretation cs);

int vips_rot_go(VipsImage *in, VipsImage **out, VipsAngle angle);
int vips_flip_horizontal_go(VipsImage *in, VipsImage **out);

int vips_extract_area_go(VipsImage *in, VipsImage **out, int left, int top, int width, int height);
int vips_smartcrop_go(VipsImage *in, VipsImage **out, int width, int height);

int vips_gaussblur_go(VipsImage *in, VipsImage **out, double sigma);
int vips_sharpen_go(VipsImage *in, VipsImage **out, double sigma);

int vips_flatten_go(VipsImage *in, VipsImage **out, double r, double g, double b);

int vips_replicate_go(VipsImage *in, VipsImage **out, int across, int down);
int vips_embed_go(VipsImage *in, VipsImage **out, int x, int y, int width, int height);

int vips_ensure_alpha(VipsImage *in, VipsImage **out);
int vips_apply_opacity(VipsImage *in, VipsImage **out, double opacity);

int vips_apply_watermark(VipsImage *in, VipsImage *watermark, VipsImage **out, double opacity);

int vips_arrayjoin_go(VipsImage **in, VipsImage **out, int n);

int vips_jpegsave_go(VipsImage *in, void **buf, size_t *len, int strip, int quality, int interlace);
int vips_pngsave_go(VipsImage *in, void **buf, size_t *len, int interlace, int embed_profile);
int vips_webpsave_go(VipsImage *in, void **buf, size_t *len, int strip, int quality);
int vips_gifsave_go(VipsImage *in, void **buf, size_t *len);
int vips_icosave_go(VipsImage *in, void **buf, size_t *len);

void vips_cleanup();
