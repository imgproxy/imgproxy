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

  void *data;         // pointer to the ICO image data in memory
  uint32_t data_size; // size of the desired picture in bytes

  VipsPel *row_buffer; // buffer for the current row, long enough to hold the whole 32-bit row+padding
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
 * Sets the image header for the output image
 */
static int
vips_foreign_load_ico_set_image_header(VipsForeignLoadIco *ico, VipsImage *out)
{
  //   vips_image_init_fields(
  //       out,
  //       ico->width,
  //       ico->height,
  //       ico->bands,
  //       VIPS_FORMAT_UCHAR,
  //       VIPS_CODING_NONE,
  //       VIPS_INTERPRETATION_sRGB,
  //       1.0,
  //       1.0);

  // // ICO files are mirrored vertically, so we need to set the orientation in reverse
  // #ifdef VIPS_META_ORIENTATION
  //   if (ico->top_down) {
  //     vips_image_set_int(out, VIPS_META_ORIENTATION, 1); // file stays top-down
  //   }
  //   else {
  //     vips_image_set_int(out, VIPS_META_ORIENTATION, 4); // top-down file is mirrored vertically
  //   }
  // #endif

  //   if (ico->palette != NULL) {
  //     int bd;

  //     if (ico->num_colors > 16) {
  //       bd = 8; // 8-bit palette
  //     }
  //     else if (ico->num_colors > 4) {
  //       bd = 4; // 4-bit palette
  //     }
  //     else if (ico->num_colors > 2) {
  //       bd = 2; // 2-bit palette
  //     }
  //     else {
  //       bd = 1; // 1-bit palette
  //     }

  //     vips_image_set_int(out, "palette-bit-depth", bd);

  // #ifdef VIPS_META_BITS_PER_SAMPLE
  //     vips_image_set_int(out, VIPS_META_BITS_PER_SAMPLE, bd);
  // #endif

  // #ifdef VIPS_META_PALETTE
  //     vips_image_set_int(out, VIPS_META_PALETTE, TRUE);
  // #endif
  //   }

  //   if (vips_image_pipelinev(out, VIPS_DEMAND_STYLE_THINSTRIP, NULL))
  //     return -1;

  return 0;
}

/**
 * Checks if the source is a ICO image
 */
static gboolean
vips_foreign_load_ico_source_is_a_source(VipsSource *source)
{
  // There is no way of detecting ICO files (it has no signature).
  // However, vips requires this method to be present.
  return FALSE;
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

    // Read the header
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

  // Read the image into memory (otherwise, it would be too complex to handle)
  ico->data_size = GUINT32_FROM_LE(largest_image_header.data_size);
  ico->data = VIPS_MALLOC(load, ico->data_size);

  if (vips_foreign_load_read_full(ico->source, ico->data, ico->data_size) <= 0) {
    vips_error("vips_foreign_load_ico_header", "unable to read ICO image data from the source");
    return -1;
  }
  // Determine the underlying format
  // Read underlying image header alone
  // Set output image parameters
  // Proceed to load the image data

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

  // Now, let's get the underlying file type. Unfortunately, we can not use vips_source_sniff
  // and all the methods which try to guess file type from the source since they all rewind source
  // to the beginning first.
  gint64 current_pos = vips_source_seek(ico->source, 0, SEEK_CUR);
  if (current_pos < 1) {
    vips_error("vips_foreign_load_ico_header", "unable to get current position in ICO source");
    return -1;
  }

  // Read the header which contains the file type
  VipsPel magic_buf[8];
  if (vips_foreign_load_read_full(ico->source, &magic_buf, sizeof(magic_buf)) <= 0) {
    vips_error("vips_foreign_load_ico_header", "unable to read file magic bytes from the source");
    return -1;
  }

  // Rewind back to the data beginning
  if (vips_source_seek(ico->source, current_pos, SEEK_SET) <= 0) {
    vips_error("vips_foreign_load_ico_header", "unable to seek to the original position in ICO source");
    return -1;
  }

  VipsPel *buffer = VIPS_MALLOC(load, ico->data_size);
  if (vips_foreign_load_read_full(ico->source, buffer, ico->data_size) <= 0) {
    vips_error("vips_foreign_load_ico_header", "unable to read ICO image data from the source");
    return -1;
  }

  // Now, let's check the magic bytes to determine the file type
  // https://www.libpng.org/pub/png/spec/1.2/PNG-Structure.html - PNG signature here
  if ((buffer[0] == 137) && (buffer[1] == 80) && (buffer[2] == 78) && (buffer[3] == 71) && (buffer[4] == 13) && (buffer[5] == 10) && (buffer[6] == 26) && (buffer[7] == 10)) {
    VipsSource *source = vips_source_new_from_memory((void *) buffer, ico->data_size);

    if (
        vips_pngload_source(
            VIPS_SOURCE(source),
            &load->real,
            "access", VIPS_ACCESS_SEQUENTIAL,
            "unlimited", 0,
            NULL) ||
        vips_source_decode(ico->source) < 0) {
      vips_error("vips_foreign_load_ico_header", "unable to load ICO image as PNG");
      return -1;
    }
  }
  else { // should be BMP otherwise
  }

  //   // For a case when we encounter buggy ICO image which has RLE command to read next
  //   // 255 bytes, and our buffer is smaller than that, we need it to be at least 255 bytes.
  //   int row_buffer_length = (ico->width * 4) + 4;
  //   if (row_buffer_length < 255) {
  //     row_buffer_length = 255;
  //   }

  //   // Allocate a row buffer for the current row in all generate* functions.
  //   // 4 * width + 4 is guaranteed to be enough for the longest (32-bit per pixel) row + padding.
  //   ico->row_buffer = VIPS_ARRAY(load, row_buffer_length, VipsPel);

  //   VipsImage **t = (VipsImage **)
  //       vips_object_local_array(VIPS_OBJECT(load), 2);

  //   t[0] = vips_image_new();

  //   if (
  //       vips_foreign_load_ico_set_image_header(ico, t[0]) ||
  //       vips_image_generate(t[0],
  //           NULL, vips_foreign_load_ico_rgb_generate, NULL,
  //           ico, NULL) ||
  //       vips_sequential(t[0], &t[1], "tile_height", VIPS__FATSTRIP_HEIGHT, NULL) ||
  //       vips_image_write(t[1], load->real) ||
  //       vips_source_decode(ico->source)) {
  //     return -1;
  //   }

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
