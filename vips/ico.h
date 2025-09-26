/*
 * ICO save/load. ICO is a container for one+ PNG/BMP images.
 */
#ifndef __ICO_H__
#define __ICO_H__

#include <stdint.h>

#define ICO_TYPE_ICO 1
#define ICO_TYPE_CURSOR 2

// ICO file header
typedef struct __attribute__((packed)) _ICONDIR_IcoHeader {
  uint16_t reserved;    // Reserved, always 0
  uint16_t type;        // 1 for ICO, 2 for CUR
  uint16_t image_count; // Number of images in the file
} ICONDIR_IcoHeader;

// ICO image header
typedef struct __attribute__((packed)) _ICONDIRENTRY_IcoHeader {
  uint8_t width;            // Width of the icon in pixels (0 for 256)
  uint8_t height;           // Height of the icon in pixels (0 for 256
  uint8_t number_of_colors; // Number of colors, not used in our case
  uint8_t reserved;         // Reserved, always 0
  uint16_t color_planes;    // Color planes, always 1
  uint16_t bpp;             // Bits per pixel
  uint32_t data_size;       // Image data size
  uint32_t data_offset;     // Image data offset, always 22
} ICONDIRENTRY_IcoHeader;

// defined in icosave.c
int vips_icosave_target_go(VipsImage *in, VipsTarget *target, ImgproxySaveOptions opts);

// defined in icoload.c
VIPS_API
int icoload_source(VipsSource *source, VipsImage **out, ...)
    G_GNUC_NULL_TERMINATED;
int vips_icoload_source_go(VipsImgproxySource *source, VipsImage **out, ImgproxyLoadOptions lo);

#endif
