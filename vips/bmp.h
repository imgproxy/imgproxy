/*
 * BMP save/load
 */
#ifndef __BMP_H__
#define __BMP_H__

#define BMP_FILE_HEADER_LEN 14        // BMP header length
#define BMP_BITMAP_INFO_HEADER_LEN 40 // BITMAPINFOHEADER
#define BMP_V4_INFO_HEADER_LEN 108    // BITMAPV4HEADER
#define BMP_V5_INFO_HEADER_LEN 124    // BITMAPV5HEADER

#define COMPRESSION_BI_RGB 0             // no compression
#define COMPRESSION_BI_RLE8 1            // RLE8
#define COMPRESSION_BI_RLE4 2            // RLE4
#define COMPRESSION_BI_BITFIELDS 3       // Has RGB bit masks
#define COMPRESSION_BI_BITFIELDS_ALPHA 6 // Has RGBA bit masks

#define BMP_PALETTE_ITEM_SIZE 4 // 4 bytes per palette item (BGR + padding)

#define BMP_RLE_EOL 0     // end of line
#define BMP_RLE_EOF 1     // end of file
#define BMP_RLE_MOVE_TO 2 // move to position

// BMP file header
typedef struct __attribute__((packed)) _BmpFileHeader {
  uint8_t sig[2];           // "BM" identifier
  uint32_t size;            // size of the BMP file
  uint8_t reserved[4];      // 4 reserved bytes
  uint32_t offset;          // offset to start of pixel data
  uint32_t info_header_len; // length of the info header
} BmpFileHeader;

// BMP DIB header
typedef struct __attribute__((packed)) _BmpDibHeader {
  // uint32 info_header_len - already in file header

  // BITMAPINFOHEADER
  int32_t width;
  int32_t height;
  uint16_t planes;
  uint16_t bpp;
  uint32_t compression;
  uint32_t image_size;
  uint32_t x_ppm;
  uint32_t y_ppm;
  uint32_t num_colors;
  uint32_t num_important_colors;

  // BITMAPV4HEADER
  uint32_t rmask;
  uint32_t gmask;
  uint32_t bmask;
  uint32_t amask;
  uint8_t cs_type[4];
  uint8_t cs[36];
  uint32_t rgamma;
  uint32_t ggamma;
  uint32_t bgamma;

  // BITMAPV5HEADER
  uint32_t intent;
  uint32_t profile_data;
  uint32_t profile_size;
  uint32_t reserved_5;
} BmpDibHeader;

// defined in bmpload.c
VIPS_API
int vips_bmpload_source(VipsSource *source, VipsImage **out, ...)
    G_GNUC_NULL_TERMINATED;
int vips_bmpload_source_go(VipsImgproxySource *source, VipsImage **out, ImgproxyLoadOptions lo);
int
vips_bmpload_buffer(void *buf, size_t len, VipsImage **out, ...);

// defined in bmpsave.c
int vips_bmpsave_target_go(VipsImage *in, VipsTarget *target, ImgproxySaveOptions opts);

#endif
