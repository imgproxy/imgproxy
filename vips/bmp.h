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
