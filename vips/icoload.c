// ICO loader
//
// See: https://en.wikipedia.org/wiki/ICO_(file_format)

#include "vips.h"

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

/**
 * ICO ForeignLoad VIPS class implementation (generic)
 */
typedef struct _VipsForeignLoadIco {
  VipsForeignLoad parent_object;
  VipsSource *source;

  VipsImage **internal; // internal image
} VipsForeignLoadIco;

typedef VipsForeignLoadClass VipsForeignLoadIcoClass;

G_DEFINE_ABSTRACT_TYPE(VipsForeignLoadIco, vips_foreign_load_ico,
    VIPS_TYPE_FOREIGN_LOAD);

static void
vips_foreign_load_ico_dispose(GObject *gobject)
{
  VipsForeignLoadIco *ico = (VipsForeignLoadIco *) gobject;

  VIPS_UNREF(ico->source);

  G_OBJECT_CLASS(vips_foreign_load_ico_parent_class)->dispose(gobject);
}

static int
vips_foreign_load_ico_build(VipsObject *object)
{
  VipsForeignLoadIco *ico = (VipsForeignLoadIco *) object;

  return VIPS_OBJECT_CLASS(vips_foreign_load_ico_parent_class)
      ->build(object);
}

static VipsForeignFlags
vips_foreign_load_ico_get_flags(VipsForeignLoad *load)
{
  return VIPS_FOREIGN_SEQUENTIAL;
}

/**
 * Checks if the source is a ICO image
 */
static gboolean
vips_foreign_load_ico_source_is_a_source(VipsSource *source)
{
  // There is no way of detecting ICO files (it has no signature).
  // However, vips requires this method to be present.
  unsigned char *buf = vips_source_sniff(source, 4);
  if (!buf) {
    vips_error("vips_foreign_load_bmp_source_is_a_source", "unable to sniff source");
    return 0;
  }

  return buf[0] == 0 && buf[1] == 0 && buf[2] == 1 && buf[3] == 0;
}

/**
 * Checks if the ICO image is a PNG image.
 */
static bool
vips_foreign_load_ico_is_png(VipsForeignLoadIco *ico, VipsPel *data, uint32_t data_size)
{
  // Check if the ICO data is PNG
  // ICO files can contain PNG images, so we need to check the magic bytes
  if (data_size < 8) {
    return false; // Not enough data to be a PNG
  }

  // Check the PNG signature
  return (data[0] == 137 && data[1] == 'P' && data[2] == 'N' &&
      data[3] == 'G' && data[4] == '\r' && data[5] == '\n' &&
      data[6] == 26 && data[7] == '\n');
}

void
vips_foreign_load_ico_free_buffer(
    VipsObject *self,
    gpointer user_data)
{
  VIPS_FREE(user_data);
}

/**
 * Loads the header of the ICO image from the source.
 */
static int
vips_foreign_load_ico_header(VipsForeignLoad *load)
{
  VipsForeignLoadIco *ico = (VipsForeignLoadIco *) load;

  // Rewind the source to the beginning
  if (vips_source_rewind(ico->source))
    return -1;

  ICONDIR_IcoHeader file_header;

  // Read the header
  if (vips_foreign_load_read_full(ico->source, &file_header, sizeof(file_header)) <= 0) {
    vips_error("vips_foreign_load_ico_header", "unable to read file header from the source");
    return -1;
  }

  // Get the image count from the file header
  uint16_t count = GUINT16_FROM_LE(file_header.image_count);

  // Now, let's find the largest image in the set
  ICONDIRENTRY_IcoHeader largest_image_header;
  memset(&largest_image_header, 0, sizeof(largest_image_header));

  for (int i = 0; i < count; i++) {
    ICONDIRENTRY_IcoHeader image_header;

    // Read the next header
    if (vips_foreign_load_read_full(ico->source, &image_header, sizeof(image_header)) <= 0) {
      vips_error("vips_foreign_load_ico_header", "unable to read file header from the source");
      return -1;
    }

    // this image width/height is greater than the largest image width/height
    // or this image width/height is 0 (which means 256)
    bool image_is_larger = ((image_header.width > largest_image_header.width) || (image_header.height > largest_image_header.height) || (image_header.width == 0) || (image_header.height == 0));

    // Update the largest image header
    if (image_is_larger) {
      memcpy(&largest_image_header, &image_header, sizeof(largest_image_header));
    }
  }

  // We failed to find any image which fits
  if (largest_image_header.data_offset == 0) {
    vips_error("vips_foreign_load_ico_header", "ICO file has no image which fits");
    return -1;
  }

  // Let's move to the ico image data offset.
  if (vips_source_seek(ico->source, GUINT32_FROM_LE(largest_image_header.data_offset), SEEK_SET) < 1) {
    vips_error("vips_foreign_load_ico_header", "unable to seek to ICO image data");
    return -1;
  }

  // Read the image into memory (otherwise, it would be too complex to handle). It's fine for ICO:
  // ICO files are usually small, and we can read them into memory without any issues.
  uint32_t data_size = GUINT32_FROM_LE(largest_image_header.data_size);

  // BMP file explicitly excludes BITMAPFILEHEADER, so we need to add it manually. We reserve
  // space for it at the beginning of the data buffer.
  VipsPel *data = (VipsPel *) VIPS_MALLOC(NULL, data_size + BMP_FILE_HEADER_LEN);
  void *actual_data = data + BMP_FILE_HEADER_LEN;

  if (vips_foreign_load_read_full(ico->source, actual_data, data_size) <= 0) {
    vips_error("vips_foreign_load_ico_header", "unable to read ICO image data from the source");
    return -1;
  }

  // Now, let's load the internal image
  ico->internal = (VipsImage **) vips_object_local_array(VIPS_OBJECT(load), 1);

  if (vips_foreign_load_ico_is_png(ico, actual_data, data_size)) {
    if (
        vips_pngload_buffer(
            actual_data, data_size,
            &ico->internal[0],
            "access", VIPS_ACCESS_SEQUENTIAL,
            NULL) < 0) {
      VIPS_FREE(data);
      vips_error("vips_foreign_load_ico_header", "unable to load ICO image as PNG");
      return -1;
    }
  }
  else {
    // Otherwise, we assume it's a BMP image.
    // According to ICO file format, it explicitly excludes BITMAPFILEHEADER (why???),
    // hence, we need to restore it to make bmp loader work.

    // Read num_colors and bpp from the BITMAPINFOHEADER
    uint32_t num_colors = GUINT32_FROM_LE(*(uint32_t *) (actual_data + 32));
    uint16_t bpp = GUINT16_FROM_LE(*(uint16_t *) (actual_data + 14));
    uint32_t pix_offset;

    if ((num_colors == 0) && (bpp <= 8)) {
      // If there are no colors and bpp is <= 8, we assume it's a palette image
      pix_offset = BMP_FILE_HEADER_LEN + BMP_BITMAP_INFO_HEADER_LEN + 4 * (1 << bpp);
    }
    else {
      // Otherwise, we use the number of colors
      pix_offset = BMP_FILE_HEADER_LEN + BMP_BITMAP_INFO_HEADER_LEN + 4 * num_colors;
    }

    // ICO file used to store alpha mask. By historical reasons, height of the ICO bmp
    // is still stored doubled to cover the alpha mask data even if they're zero
    // or not present.
    int32_t height = GINT32_FROM_LE(*(int32_t *) (actual_data + 8));
    height = height / 2;

    // Magic bytes
    data[0] = 'B';
    data[1] = 'M';

    // Size of the BMP file (data size + BMP file header length)
    (*(uint32_t *) (data + 2)) = GUINT32_TO_LE(data_size + BMP_FILE_HEADER_LEN);
    (*(uint32_t *) (data + 6)) = 0;                          // reserved
    (*(uint32_t *) (data + 10)) = GUINT32_TO_LE(pix_offset); // offset to the pixel data
    (*(int32_t *) (actual_data + 8)) = GINT32_TO_LE(height); // height

    if (
        vips_bmpload_buffer(
            data, data_size + BMP_FILE_HEADER_LEN,
            &ico->internal[0],
            "access", VIPS_ACCESS_SEQUENTIAL,
            NULL) < 0) {
      VIPS_FREE(data);
      vips_error("vips_foreign_load_ico_header", "unable to load ICO image as BMP");
      return -1;
    }
  }

  // It is recommended that we free the buffer in postclose callback
  g_signal_connect(
      ico->internal[0], "postclose",
      G_CALLBACK(vips_foreign_load_ico_free_buffer), data);

  // Copy the image metadata parameters to the load->out image.
  // This should be sufficient, as we do not care much about the rest of the
  // metadata inside .ICO files. At least, at this stage.
  vips_image_init_fields(
      load->out,
      vips_image_get_width(ico->internal[0]),
      vips_image_get_height(ico->internal[0]),
      vips_image_get_bands(ico->internal[0]),
      vips_image_get_format(ico->internal[0]),
      vips_image_get_coding(ico->internal[0]),
      vips_image_get_interpretation(ico->internal[0]),
      vips_image_get_xres(ico->internal[0]),
      vips_image_get_yres(ico->internal[0]));

  vips_source_minimise(ico->source);

  return 0;
}

/**
 * Loads a ICO image from the source.
 */
static int
vips_foreign_load_ico_load(VipsForeignLoad *load)
{
  VipsForeignLoadIco *ico = (VipsForeignLoadIco *) load;
  VipsImage *image = ico->internal[0];

  // Just copy the internal image to the output image
  if (vips_image_write(image, &load->real)) {
    vips_error("vips_foreign_load_ico_load", "unable to copy ICO image to output");
    return -1;
  }

  return 0;
}

static void
vips_foreign_load_ico_class_init(VipsForeignLoadIcoClass *class)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(class);
  VipsObjectClass *object_class = (VipsObjectClass *) class;
  VipsForeignClass *foreign_class = (VipsForeignClass *) class;
  VipsForeignLoadClass *load_class = (VipsForeignLoadClass *) class;

  gobject_class->dispose = vips_foreign_load_ico_dispose;
  gobject_class->set_property = vips_object_set_property;
  gobject_class->get_property = vips_object_get_property;

  object_class->nickname = "icoload_base";
  object_class->description = "load ico image (internal format for video thumbs)";
  object_class->build = vips_foreign_load_ico_build;

  load_class->get_flags = vips_foreign_load_ico_get_flags;
  load_class->header = vips_foreign_load_ico_header;
  load_class->load = vips_foreign_load_ico_load;
}

static void
vips_foreign_load_ico_init(VipsForeignLoadIco *load)
{
}

typedef struct _VipsForeignLoadIcoSource {
  VipsForeignLoadIco parent_object;

  VipsSource *source;
} VipsForeignLoadIcoSource;

typedef VipsForeignLoadIcoClass VipsForeignLoadIcoSourceClass;

G_DEFINE_TYPE(VipsForeignLoadIcoSource, vips_foreign_load_ico_source,
    vips_foreign_load_ico_get_type());

static int
vips_foreign_load_ico_source_build(VipsObject *object)
{
  VipsForeignLoadIco *ico = (VipsForeignLoadIco *) object;
  VipsForeignLoadIcoSource *source =
      (VipsForeignLoadIcoSource *) object;

  if (source->source) {
    ico->source = source->source;
    g_object_ref(ico->source);
  }

  return VIPS_OBJECT_CLASS(vips_foreign_load_ico_source_parent_class)
      ->build(object);
}

static void
vips_foreign_load_ico_source_class_init(
    VipsForeignLoadIcoSourceClass *class)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(class);
  VipsObjectClass *object_class = (VipsObjectClass *) class;
  VipsOperationClass *operation_class = VIPS_OPERATION_CLASS(class);
  VipsForeignLoadClass *load_class = (VipsForeignLoadClass *) class;

  gobject_class->set_property = vips_object_set_property;
  gobject_class->get_property = vips_object_get_property;

  object_class->nickname = "icoload_source";
  object_class->description = "load image from ico source";
  object_class->build = vips_foreign_load_ico_source_build;

  load_class->is_a_source = vips_foreign_load_ico_source_is_a_source;

  VIPS_ARG_OBJECT(class, "source", 1,
      "Source",
      "Source to load from",
      VIPS_ARGUMENT_REQUIRED_INPUT,
      G_STRUCT_OFFSET(VipsForeignLoadIcoSource, source),
      VIPS_TYPE_SOURCE);
}

static void
vips_foreign_load_ico_source_init(VipsForeignLoadIcoSource *source)
{
}

/**
 * vips_icoload_source:
 * @source: source to load
 * @out: (out): image to write
 * @...: `NULL`-terminated list of optional named arguments
 *
 * Read a RAWRGB-formatted memory block into a VIPS image.
 *
 * Returns: 0 on success, -1 on error.
 */
int
vips_icoload_source(VipsSource *source, VipsImage **out, ...)
{
  va_list ap;
  int result;

  va_start(ap, out);
  result = vips_call_split("icoload_source", ap, source, out);
  va_end(ap);

  return result;
}

// wrapper function which hiders varargs (...) from CGo
int
vips_icoload_source_go(VipsImgproxySource *source, VipsImage **out)
{
  return vips_icoload_source(VIPS_SOURCE(source), out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}
