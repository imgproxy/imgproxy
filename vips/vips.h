#include <stdlib.h>
#include <stdint.h> // uintptr_t

#include <vips/vips.h>
#include <vips/connection.h>
#include <vips/vips7compat.h>
#include <vips/vector.h>

typedef struct _RGB {
  double r;
  double g;
  double b;
} RGB;

// vips async source
typedef struct _VipsAsyncSource {
  VipsSourceCustom source; // class designator
  uintptr_t readerHandle;  // async reader handler
} VipsAsyncSource;

// glib class for vips async source
typedef struct _VipsAsyncSourceClass {
  VipsSourceCustomClass parent_class;
} VipsAsyncSourceClass;

int vips_initialize();

void clear_image(VipsImage **in);
void g_free_go(void **buf);

void swap_and_clear(VipsImage **in, VipsImage *out);

int gif_resolution_limit();

int vips_health();

int vips_jpegload_go(void *buf, size_t len, int shrink, VipsImage **out);
int vips_jxlload_go(void *buf, size_t len, int pages, VipsImage **out);
int vips_pngload_go(void *buf, size_t len, VipsImage **out, int unlimited);
int vips_webpload_go(void *buf, size_t len, double scale, int pages, VipsImage **out);
int vips_gifload_go(void *buf, size_t len, int pages, VipsImage **out);
int vips_svgload_go(void *buf, size_t len, double scale, VipsImage **out, int unlimited);
int vips_heifload_go(void *buf, size_t len, VipsImage **out, int thumbnail);
int vips_tiffload_go(void *buf, size_t len, VipsImage **out);

int vips_black_go(VipsImage **out, int width, int height, int bands);

int vips_fix_float_tiff(VipsImage *in, VipsImage **out);

int vips_get_orientation(VipsImage *image);

VipsBandFormat vips_band_format(VipsImage *in);

gboolean vips_is_animated(VipsImage *in);

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
int vips_icc_transform_go(VipsImage *in, VipsImage **out);
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

int vips_jpegsave_go(VipsImage *in, void **buf, size_t *len, int quality, int interlace);
int vips_jxlsave_go(VipsImage *in, void **buf, size_t *len, int quality, int effort);
int vips_pngsave_go(VipsImage *in, void **buf, size_t *len, int interlace, int quantize,
    int colors);
int vips_webpsave_go(VipsImage *in, void **buf, size_t *len, int quality);
int vips_gifsave_go(VipsImage *in, void **buf, size_t *len);
int vips_heifsave_go(VipsImage *in, void **buf, size_t *len, int quality);
int vips_avifsave_go(VipsImage *in, void **buf, size_t *len, int quality, int speed);
int vips_tiffsave_go(VipsImage *in, void **buf, size_t *len, int quality);

void vips_cleanup();

// vips async source read function
int vips_jpegloadsource_go(VipsAsyncSource *source, int shrink, VipsImage **out);

// creates new vips async source from a reader handle
VipsAsyncSource *vips_new_async_source(uintptr_t readerHandle);

// attaches "close" signal to the vips image: closes reader and unrefs vips source
void vips_attach_image_close_signals(VipsImage **in, uintptr_t handle, VipsAsyncSource *source);

// closes source and corresponding reader
void close_source(VipsImage **in, VipsAsyncSource *source);
