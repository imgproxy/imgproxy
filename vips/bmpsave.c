// BMP saver

#include "vips.h"

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

/* Save a bit of typing.
 */
#define UC VIPS_FORMAT_UCHAR

static VipsBandFormat bandfmt_bmp[10] = {
  /* Band format:  UC  C   US  S   UI  I   F  X  D  DX */
  /* Promotion: */ UC, UC, UC, UC, UC, UC, UC, UC, UC, UC
};

// BMP BITMAPINFOHEADERV5 file header struct, ((packed)) since we
// do not want any compiler-induced padding
typedef struct __attribute__((packed)) _BmpHeader {
  uint8_t sig[2];              // Signature 'BM'
  uint32_t file_size;          // File size in bytes
  uint16_t reserved[2];        // Reserved fields
  uint32_t pix_offset;         // Offset to pixel data
  uint32_t dib_header_size;    // DIB header size
  int32_t width;               // Image width
  int32_t height;              // Image height
  uint16_t color_plane;        // Number of color planes
  uint16_t bpp;                // Bits per pixel
  uint32_t compression;        // Compression method
  uint32_t image_size;         // Image size
  uint32_t x_pixels_per_meter; // Horizontal resolution
  uint32_t y_pixels_per_meter; // Vertical resolution
  uint32_t color_use;          // Number of colors in palette
  uint32_t color_important;    // Number of important colors
  uint32_t rmask;              // Red mask
  uint32_t gmask;              // Green mask
  uint32_t bmask;              // Blue mask
  uint32_t amask;              // Alpha mask (optional, only for 32 bpp BMP files)
  uint8_t cs_type[4];          // Color space type (B G R s)
  uint8_t cs[36];              // CIEXYZTRIPLE Color Space
  uint32_t red_gamma;          // Red gamma
  uint32_t green_gamma;        // Green gamma
  uint32_t blue_gamma;         // Blue gamma
  uint32_t intent;
  uint32_t profile_data; // Profile data (optional, only for 32 bpp BMP files)
  uint32_t profile_size;
  uint32_t reserved_5;
} BmpHeader;

typedef struct _VipsForeignSaveBmp {
  VipsForeignSave parent_object;

  VipsTarget *target;
  VipsPel *line_buffer;

  uint16_t bands;
  uint32_t line_size;
} VipsForeignSaveBmp;

typedef VipsForeignSaveClass VipsForeignSaveBmpClass;

G_DEFINE_ABSTRACT_TYPE(VipsForeignSaveBmp, vips_foreign_save_bmp,
    VIPS_TYPE_FOREIGN_SAVE);

static void
vips_foreign_save_bmp_dispose(GObject *gobject)
{
  VipsForeignSaveBmp *bmp = (VipsForeignSaveBmp *) gobject;

  VIPS_UNREF(bmp->target);

  G_OBJECT_CLASS(vips_foreign_save_bmp_parent_class)->dispose(gobject);
}

static int
vips_foreign_save_bmp_block(VipsRegion *region, VipsRect *area, void *a)
{
  VipsForeignSaveBmp *bmp = (VipsForeignSaveBmp *) a;
  VipsImage *image = region->im;

  // This is the position in the source image
  uint32_t source_row_size = region->im->Xsize * bmp->bands;

  for (int y = 0; y < area->height; y++) {
    VipsPel *src = VIPS_REGION_ADDR(region, 0, area->top + y);
    VipsPel *dst = bmp->line_buffer;

    for (int x = 0; x < source_row_size; x += bmp->bands) {
      dst[0] = src[2]; // B
      dst[1] = src[1]; // G
      dst[2] = src[0]; // R

      if (bmp->bands == 4) {
        dst[3] = src[3]; // A
      }

      dst += bmp->bands;
      src += bmp->bands;
    }

    if (vips_target_write(bmp->target, bmp->line_buffer, bmp->line_size) < 0) {
      vips_error("vips_foreign_save_bmp_build", "unable to write BMP pixel data to target");
      return -1;
    }
  }

  return 0;
}

static int
vips_foreign_save_bmp_build(VipsObject *object)
{
  VipsForeignSave *save = (VipsForeignSave *) object;
  VipsForeignSaveBmp *bmp = (VipsForeignSaveBmp *) object;

  VipsImage *in;

  if (VIPS_OBJECT_CLASS(vips_foreign_save_bmp_parent_class)->build(object))
    return -1;

  in = save->ready; // shortcut

  // bands (3 or 4) * 8 bits
  int bands = vips_image_get_bands(in);

  if ((bands < 3) || (bands > 4)) {
    vips_error("vips_foreign_save_bmp_build", "BMP source file must have 3 or 4 bands (RGB or RGBA)");
    return -1;
  }

  int bpp = bands * 8;

  // Target image line size trimmed to 4 bytes.
  uint32_t line_size = (in->Xsize * bands + 3) & (~3);
  uint32_t image_size = in->Ysize * line_size;

  // pix_offset = header size + file size
  uint32_t pix_offset = BMP_FILE_HEADER_LEN + BMP_V5_INFO_HEADER_LEN;

  // Format BMP file header. We write 24/32 bpp BMP files only with no compression.
  BmpHeader header;

  header.sig[0] = 'B';
  header.sig[1] = 'M';
  header.file_size = GUINT32_TO_LE(pix_offset + image_size);
  header.reserved[0] = 0;
  header.reserved[1] = 0;
  header.pix_offset = GUINT32_TO_LE(pix_offset);
  header.dib_header_size = GUINT32_TO_LE(BMP_V5_INFO_HEADER_LEN);
  header.width = GINT32_TO_LE(in->Xsize);
  header.height = GINT32_TO_LE(-in->Ysize);
  header.color_plane = GUINT16_TO_LE(1);
  header.bpp = GUINT16_TO_LE(bpp);
  header.compression = COMPRESSION_BI_RGB;
  header.image_size = GUINT32_TO_LE(image_size);
  header.x_pixels_per_meter = 0; // GUINT32_TO_LE(2835);
  header.y_pixels_per_meter = 0; // GUINT32_TO_LE(2835);
  header.color_use = 0;
  header.color_important = 0;
  header.rmask = GUINT32_TO_LE(0x00FF0000); // Standard says that masks are in BE order
  header.gmask = GUINT32_TO_LE(0x0000FF00);
  header.bmask = GUINT32_TO_LE(0x000000FF);
  header.amask = GUINT32_TO_LE(0xFF000000);
  header.cs_type[0] = 'B'; // Image color profile
  header.cs_type[1] = 'G';
  header.cs_type[2] = 'R';
  header.cs_type[3] = 's';
  memset(header.cs, 0, sizeof(header.cs)); // CIEXYZTRIPLE Color Space
  header.red_gamma = 0;
  header.green_gamma = 0;
  header.blue_gamma = 0;
  header.intent = GUINT32_TO_LE(4); // IMAGES intent, must be 4
  header.profile_data = 0;
  header.profile_size = 0;
  header.reserved_5 = 0;

  if (vips_target_write(bmp->target, &header, sizeof(header)) < 0) {
    vips_error("vips_foreign_save_bmp_build", "unable to write BMP header to target");
    return -1;
  }

  // Allocate a line buffer for the target image
  bmp->line_buffer = VIPS_MALLOC(save, line_size);
  bmp->bands = bands;
  bmp->line_size = line_size;

  // save image async
  if (vips_sink_disc(in, vips_foreign_save_bmp_block, bmp))
    return -1;

  if (vips_target_end(bmp->target))
    return -1;

  return 0;
}

static void
vips_foreign_save_bmp_class_init(VipsForeignSaveBmpClass *class)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(class);
  VipsObjectClass *object_class = (VipsObjectClass *) class;
  VipsForeignClass *foreign_class = (VipsForeignClass *) class;
  VipsForeignSaveClass *save_class = (VipsForeignSaveClass *) class;

  gobject_class->dispose = vips_foreign_save_bmp_dispose;
  gobject_class->set_property = vips_object_set_property;
  gobject_class->get_property = vips_object_get_property;

  object_class->nickname = "bmpsave_base";
  object_class->description = "save bmp";
  object_class->build = vips_foreign_save_bmp_build;

  // We do not support saving monochrome images yet (VIPS_FOREIGN_SAVEABLE_MONO)
  // In v4 we will support it, so we leave it here commented out
  save_class->saveable =
      VIPS_SAVEABLE_RGB | // latest vips: VIPS_FOREIGN_SAVEABLE_RGB
      VIPS_SAVEABLE_RGBA; // latest vips: VIPS_FOREIGN_SAVEABLE_ALPHA

  save_class->format_table = bandfmt_bmp;
}

static void
vips_foreign_save_bmp_init(VipsForeignSaveBmp *bmp)
{
}

typedef struct _VipsForeignSaveBmpTarget {
  VipsForeignSaveBmp parent_object;

  VipsTarget *target;
} VipsForeignSaveBmpTarget;

typedef VipsForeignSaveBmpClass VipsForeignSaveBmpTargetClass;

G_DEFINE_TYPE(VipsForeignSaveBmpTarget, vips_foreign_save_bmp_target,
    vips_foreign_save_bmp_get_type());

static int
vips_foreign_save_bmp_target_build(VipsObject *object)
{
  VipsForeignSaveBmp *bmp = (VipsForeignSaveBmp *) object;
  VipsForeignSaveBmpTarget *target = (VipsForeignSaveBmpTarget *) object;

  bmp->target = target->target;
  g_object_ref(bmp->target);

  return VIPS_OBJECT_CLASS(vips_foreign_save_bmp_target_parent_class)
      ->build(object);
}

static void
vips_foreign_save_bmp_target_class_init(VipsForeignSaveBmpTargetClass *class)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(class);
  VipsObjectClass *object_class = (VipsObjectClass *) class;

  gobject_class->set_property = vips_object_set_property;
  gobject_class->get_property = vips_object_get_property;

  object_class->nickname = "bmpsave_target";
  object_class->description = "save image to target as PNG";
  object_class->build = vips_foreign_save_bmp_target_build;

  VIPS_ARG_OBJECT(class, "target", 1,
      "Target",
      "Target to save to",
      VIPS_ARGUMENT_REQUIRED_INPUT,
      G_STRUCT_OFFSET(VipsForeignSaveBmpTarget, target),
      VIPS_TYPE_TARGET);
}

static void
vips_foreign_save_bmp_target_init(VipsForeignSaveBmpTarget *target)
{
}

/**
 * vips_bmpsave_target: (method)
 * @in: image to save
 * @target: save image to this target
 * @...: `NULL`-terminated list of optional named arguments
 *
 * As [method@Image.bmpsave], but save to a target.
 *
 * ::: seealso
 *     [method@Image.bmpsave], [method@Image.write_to_target].
 *
 * Returns: 0 on success, -1 on error.
 */
int
vips_bmpsave_target(VipsImage *in, VipsTarget *target, ...)
{
  va_list ap;
  int result;

  va_start(ap, target);
  result = vips_call_split("bmpsave_target", ap, in, target);
  va_end(ap);

  return result;
}

// wrapper function which hides varargs (...) from CGo
int
vips_bmpsave_target_go(VipsImage *in, VipsTarget *target)
{
  return vips_bmpsave_target(in, VIPS_TARGET(target), NULL);
}
