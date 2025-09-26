#include <stdlib.h>
#include <stdint.h> // uintptr_t

#include <vips/vips.h>
#include <vips/vips7compat.h>
#include <vips/vector.h>
#include <vips/foreign.h>

#include "options.h"
#include "source.h"
#include "bmp.h"
#include "ico.h"

typedef struct _RGB {
  double r;
  double g;
  double b;
} RGB;

int vips_initialize();

void unref_image(VipsImage *in);
void g_free_go(void **buf);

int gif_resolution_limit();

int vips_health();

int vips_jpegload_source_go(VipsImgproxySource *source, int shrink, VipsImage **out);
int vips_jxlload_source_go(VipsImgproxySource *source, int pages, VipsImage **out);
int vips_pngload_source_go(VipsImgproxySource *source, VipsImage **out, int unlimited);
int vips_webpload_source_go(VipsImgproxySource *source, double scale, int pages, VipsImage **out);
int vips_gifload_source_go(VipsImgproxySource *source, int pages, VipsImage **out);
int vips_svgload_source_go(VipsImgproxySource *source, double scale, VipsImage **out, int unlimited);
int vips_heifload_source_go(VipsImgproxySource *source, VipsImage **out, int thumbnail);
int vips_tiffload_source_go(VipsImgproxySource *source, VipsImage **out);

int vips_black_go(VipsImage **out, int width, int height, int bands);

int vips_fix_float_tiff(VipsImage *in, VipsImage **out);

int vips_get_orientation(VipsImage *image);

VipsBandFormat vips_band_format(VipsImage *in);

gboolean vips_image_is_animated(VipsImage *in);
int vips_image_remove_animation(VipsImage *in, VipsImage **out);

int vips_image_get_array_int_go(VipsImage *image, const char *name, int **out, int *n);
void vips_image_set_array_int_go(VipsImage *image, const char *name, const int *array, int n);

int vips_addalpha_go(VipsImage *in, VipsImage **out);

int vips_copy_go(VipsImage *in, VipsImage **out);

int vips_cast_go(VipsImage *in, VipsImage **out, VipsBandFormat format);
int vips_rad2float_go(VipsImage *in, VipsImage **out);

int vips_resize_go(VipsImage *in, VipsImage **out, double wscale, double hscale);

int vips_icc_is_srgb_iec61966(VipsImage *in);
int vips_has_embedded_icc(VipsImage *in);
int vips_icc_backup(VipsImage *in, VipsImage **out);
int vips_icc_restore(VipsImage *in, VipsImage **out);
int vips_icc_import_go(VipsImage *in, VipsImage **out);
int vips_icc_export_go(VipsImage *in, VipsImage **out);
int vips_icc_export_srgb(VipsImage *in, VipsImage **out);
int vips_icc_transform_srgb(VipsImage *in, VipsImage **out);
int vips_icc_remove(VipsImage *in, VipsImage **out);
int vips_colourspace_go(VipsImage *in, VipsImage **out, VipsInterpretation cs);

int vips_rot_go(VipsImage *in, VipsImage **out, VipsAngle angle);
int vips_flip_horizontal_go(VipsImage *in, VipsImage **out);

int vips_extract_area_go(VipsImage *in, VipsImage **out, int left, int top, int width, int height);
int vips_smartcrop_go(VipsImage *in, VipsImage **out, int width, int height);
int vips_trim(VipsImage *in, VipsImage **out, double threshold, gboolean smart, RGB bg, gboolean equal_hor, gboolean equal_ver);

int vips_apply_filters(VipsImage *in, VipsImage **out, double blur_sigma, double sharp_sigma,
    int pixelate_pixels);

int vips_flatten_go(VipsImage *in, VipsImage **out, RGB bg);

int vips_replicate_go(VipsImage *in, VipsImage **out, int across, int down, int centered);
int vips_embed_go(VipsImage *in, VipsImage **out, int x, int y, int width, int height);

int vips_apply_watermark(VipsImage *in, VipsImage *watermark, VipsImage **out, int left, int top,
    double opacity);

int vips_linecache_seq(VipsImage *in, VipsImage **out, int tile_height);

int vips_arrayjoin_go(VipsImage **in, VipsImage **out, int n);

int vips_strip(VipsImage *in, VipsImage **out, int keep_exif_copyright);
int vips_strip_all(VipsImage *in, VipsImage **out);

int vips_jpegsave_go(VipsImage *in, VipsTarget *target, int quality, ImgproxySaveOptions opts);
int vips_jxlsave_go(VipsImage *in, VipsTarget *target, int quality, ImgproxySaveOptions opts);
int vips_pngsave_go(VipsImage *in, VipsTarget *target, ImgproxySaveOptions opts);
int vips_webpsave_go(VipsImage *in, VipsTarget *target, int quality, ImgproxySaveOptions opts);
int vips_gifsave_go(VipsImage *in, VipsTarget *target, ImgproxySaveOptions opts);
int vips_heifsave_go(VipsImage *in, VipsTarget *target, int quality, ImgproxySaveOptions opts);
int vips_avifsave_go(VipsImage *in, VipsTarget *target, int quality, ImgproxySaveOptions opts);
int vips_tiffsave_go(VipsImage *in, VipsTarget *target, int quality, ImgproxySaveOptions opts);

void vips_cleanup();

void vips_error_go(const char *function, const char *message);

int vips_foreign_load_read_full(VipsSource *source, void *buf, size_t len);
void vips_unref_target(VipsTarget *target);
