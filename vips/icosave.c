// ICO saver

#include "vips.h"

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>

/* Save a bit of typing.
 */
#define UC VIPS_FORMAT_UCHAR

static VipsBandFormat bandfmt_ico[10] = {
  /* Band format:  UC  C   US  S   UI  I   F  X  D  DX */
  /* Promotion: */ UC, UC, UC, UC, UC, UC, UC, UC, UC, UC
};

static uint8_t icon_dir[6] = { 0, 0, 1, 0, 1, 0 }; // ICONDIR header

// ICO file header
typedef struct __attribute__((packed)) _IcoHeader {
  uint8_t dir[6];           // ICONDIR
  uint8_t width;            // Width of the icon in pixels (0 for 256)
  uint8_t height;           // Height of the icon in pixels (0 for 256
  uint8_t number_of_colors; // Number of colors, not used in our case
  uint8_t reserved;         // Reserved, always 0
  uint16_t color_planes;    // Color planes, always 1
  uint16_t bpp;             // Bits per pixel
  uint32_t data_size;       // Image data size
  uint32_t data_offset;     // Image data offset, always 22
} IcoHeader;

typedef struct _VipsForeignSaveIco {
  VipsForeignSave parent_object;

  VipsTarget *target;
  VipsPel *line_buffer;

  uint16_t bands;
  uint32_t line_size;
} VipsForeignSaveIco;

typedef VipsForeignSaveClass VipsForeignSaveIcoClass;

G_DEFINE_ABSTRACT_TYPE(VipsForeignSaveIco, vips_foreign_save_ico,
    VIPS_TYPE_FOREIGN_SAVE);

static void
vips_foreign_save_ico_dispose(GObject *gobject)
{
  VipsForeignSaveIco *ico = (VipsForeignSaveIco *) gobject;

  VIPS_UNREF(ico->target);

  G_OBJECT_CLASS(vips_foreign_save_ico_parent_class)->dispose(gobject);
}

static int
vips_foreign_save_ico_build(VipsObject *object)
{
  VipsForeignSave *save = (VipsForeignSave *) object;
  VipsForeignSaveIco *ico = (VipsForeignSaveIco *) object;

  VipsImage *in;

  if (VIPS_OBJECT_CLASS(vips_foreign_save_ico_parent_class)->build(object))
    return -1;

  in = save->ready; // shortcut

  if ((in->Xsize > 256) || (in->Ysize > 256)) {
    vips_error("vips_foreign_save_ico_build", "Image is too big. Max dimension size for ICO is 256");
    return -1;
  }

  // bands (3 or 4) * 8 bits
  int bands = vips_image_get_bands(in);

  if (bands > 4) {
    vips_error("vips_foreign_save_ico_build", "ICO source file must have 3 or 4 bands (RGB or RGBA)");
    return -1;
  }

  uint8_t width = in->Xsize % 256;  // 0 means 256
  uint8_t height = in->Ysize % 256; // 0 means 256
  uint16_t bpp = 24;

  if (bands > 3) {
    bpp = 32;
  }

  uint32_t data_offset = 22; // ICO header size + ICONDIRENTRY size

  IcoHeader header;

  memcpy(header.dir, icon_dir, sizeof(icon_dir));
  header.width = width;                            // Width of the icon in pixels (0 for 256)
  header.height = height;                          // Height of the icon in pixels (0 for 256)
  header.number_of_colors = 0;                     // Number of colors, not used in our case
  header.reserved = 0;                             // Reserved, always 0
  header.color_planes = 1;                         // Color planes, always 1
  header.bpp = GUINT16_TO_LE(bpp);                 // Bits per pixel
  header.data_offset = GUINT32_TO_LE(data_offset); // Image data offset, always 22

  void *buffer;
  size_t data_size;

  // Save PNG image to a buffer
  if (vips_pngsave_buffer(in, &buffer, &data_size, NULL)) {
    vips_error("vips_foreign_save_ico_build", "unable to save ICO image as PNG");
    return -1;
  }

  header.data_size = GUINT32_TO_LE(data_size); // Image data size

  // Write header
  if (vips_target_write(ico->target, &header, sizeof(header))) {
    g_free(buffer);
    vips_error("vips_foreign_save_ico_build", "unable to write ICO header to target");
    return -1;
  }

  // Write data
  if (vips_target_write(ico->target, buffer, data_size)) {
    g_free(buffer);
    vips_error("vips_foreign_save_ico_build", "unable to write ICO header to target");
    return -1;
  }

  g_free(buffer);

  if (vips_target_end(ico->target))
    return -1;

  return 0;
}

static void
vips_foreign_save_ico_class_init(VipsForeignSaveIcoClass *class)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(class);
  VipsObjectClass *object_class = (VipsObjectClass *) class;
  VipsForeignClass *foreign_class = (VipsForeignClass *) class;
  VipsForeignSaveClass *save_class = (VipsForeignSaveClass *) class;

  gobject_class->dispose = vips_foreign_save_ico_dispose;
  gobject_class->set_property = vips_object_set_property;
  gobject_class->get_property = vips_object_get_property;

  object_class->nickname = "icosave_base";
  object_class->description = "save ico";
  object_class->build = vips_foreign_save_ico_build;

  // We do not support saving monochrome images yet (VIPS_FOREIGN_SAVEABLE_MONO)
  // In v4 we will support it, so we leave it here commented out
  save_class->saveable =
      VIPS_SAVEABLE_RGB | // latest vips: VIPS_FOREIGN_SAVEABLE_RGB
      VIPS_SAVEABLE_RGBA; // latest vips: VIPS_FOREIGN_SAVEABLE_ALPHA

  save_class->format_table = bandfmt_ico;
}

static void
vips_foreign_save_ico_init(VipsForeignSaveIco *ico)
{
}

typedef struct _VipsForeignSaveIcoTarget {
  VipsForeignSaveIco parent_object;

  VipsTarget *target;
} VipsForeignSaveIcoTarget;

typedef VipsForeignSaveIcoClass VipsForeignSaveIcoTargetClass;

G_DEFINE_TYPE(VipsForeignSaveIcoTarget, vips_foreign_save_ico_target,
    vips_foreign_save_ico_get_type());

static int
vips_foreign_save_ico_target_build(VipsObject *object)
{
  VipsForeignSaveIco *ico = (VipsForeignSaveIco *) object;
  VipsForeignSaveIcoTarget *target = (VipsForeignSaveIcoTarget *) object;

  ico->target = target->target;
  g_object_ref(ico->target);

  return VIPS_OBJECT_CLASS(vips_foreign_save_ico_target_parent_class)
      ->build(object);
}

static void
vips_foreign_save_ico_target_class_init(VipsForeignSaveIcoTargetClass *class)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(class);
  VipsObjectClass *object_class = (VipsObjectClass *) class;

  gobject_class->set_property = vips_object_set_property;
  gobject_class->get_property = vips_object_get_property;

  object_class->nickname = "icosave_target";
  object_class->description = "save image to target as PNG";
  object_class->build = vips_foreign_save_ico_target_build;

  VIPS_ARG_OBJECT(class, "target", 1,
      "Target",
      "Target to save to",
      VIPS_ARGUMENT_REQUIRED_INPUT,
      G_STRUCT_OFFSET(VipsForeignSaveIcoTarget, target),
      VIPS_TYPE_TARGET);
}

static void
vips_foreign_save_ico_target_init(VipsForeignSaveIcoTarget *target)
{
}

/**
 * vips_icosave_target: (method)
 * @in: image to save
 * @target: save image to this target
 * @...: `NULL`-terminated list of optional named arguments
 *
 * As [method@Image.icosave], but save to a target.
 *
 * ::: seealso
 *     [method@Image.icosave], [method@Image.write_to_target].
 *
 * Returns: 0 on success, -1 on error.
 */
int
vips_icosave_target(VipsImage *in, VipsTarget *target, ...)
{
  va_list ap;
  int result;

  va_start(ap, target);
  result = vips_call_split("icosave_target", ap, in, target);
  va_end(ap);

  return result;
}

// wrapper function which hides varargs (...) from CGo
int
vips_icosave_target_go(VipsImage *in, VipsTarget *target, ImgproxySaveOptions opts)
{
  return vips_icosave_target(in, VIPS_TARGET(target), NULL);
}
