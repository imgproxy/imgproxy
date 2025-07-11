// BMP loader
//
// See: https://en.wikipedia.org/wiki/BMP_file_format

#include "vips.h"

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

#define BMP_ROW_ADDR(BMP, OR, R, Y) \
  VIPS_REGION_ADDR(OR, 0, R->top + (BMP->top_down ? Y : R->height - 1 - Y))

/**
 * BMP ForeignLoad VIPS class implementation (generic)
 */
typedef struct _VipsForeignLoadBmp {
  VipsForeignLoad parent_object;
  VipsSource *source;

  int32_t width;
  int32_t height;
  uint16_t planes;
  uint16_t bpp;
  uint16_t compression;
  uint16_t offset;

  uint32_t rmask;
  uint32_t gmask;
  uint32_t bmask;
  uint32_t amask;

  uint32_t num_colors;

  int bands;           // target image bands
  int bytes_per_pixel; // source image bytes per pixel (not used when bpp<8)

  bool top_down; // true if image is vertically reversed
  bool rle;      // true if image is compressed with RLE
  bool bmp565;   // 16-bit

  uint32_t *palette; // palette for 1, 2, 4 or 8 bits per pixel BMP palette

  int y_pos; // current position in the image, used when sequential access is possible

  int dy; // in RLE mode this indicates how many lines to skip
  int dx; // in RLE mode this indicates start pixel X

  VipsPel *row_buffer; // buffer for the current row, long enough to hold the whole 32-bit row+padding
} VipsForeignLoadBmp;

typedef VipsForeignLoadClass VipsForeignLoadBmpClass;

G_DEFINE_ABSTRACT_TYPE(VipsForeignLoadBmp, vips_foreign_load_bmp,
    VIPS_TYPE_FOREIGN_LOAD);

static void
vips_foreign_load_bmp_dispose(GObject *gobject)
{
  VipsForeignLoadBmp *bmp = (VipsForeignLoadBmp *) gobject;

  VIPS_UNREF(bmp->source);

  G_OBJECT_CLASS(vips_foreign_load_bmp_parent_class)->dispose(gobject);
}

static int
vips_foreign_load_bmp_build(VipsObject *object)
{
  VipsForeignLoadBmp *bmp = (VipsForeignLoadBmp *) object;

  return VIPS_OBJECT_CLASS(vips_foreign_load_bmp_parent_class)
      ->build(object);
}

static VipsForeignFlags
vips_foreign_load_bmp_get_flags(VipsForeignLoad *load)
{
  return VIPS_FOREIGN_SEQUENTIAL;
}

/**
 * Sets the image header for the output image
 */
static int
vips_foreign_load_bmp_set_image_header(VipsForeignLoadBmp *bmp, VipsImage *out)
{
  vips_image_init_fields(
      out,
      bmp->width,
      bmp->height,
      bmp->bands,
      VIPS_FORMAT_UCHAR,
      VIPS_CODING_NONE,
      VIPS_INTERPRETATION_sRGB,
      1.0,
      1.0);

  if (bmp->palette != NULL) {
    int bd;

    if (bmp->num_colors > 16) {
      bd = 8; // 8-bit palette
    }
    else if (bmp->num_colors > 4) {
      bd = 4; // 4-bit palette
    }
    else if (bmp->num_colors > 2) {
      bd = 2; // 2-bit palette
    }
    else {
      bd = 1; // 1-bit palette
    }

    vips_image_set_int(out, "palette-bit-depth", bd);

#ifdef VIPS_META_BITS_PER_SAMPLE
    vips_image_set_int(out, VIPS_META_BITS_PER_SAMPLE, bd);
#endif

#ifdef VIPS_META_PALETTE
    vips_image_set_int(out, VIPS_META_PALETTE, TRUE);
#endif
  }

  if (vips_image_pipelinev(out, VIPS_DEMAND_STYLE_THINSTRIP, NULL))
    return -1;

  return 0;
}

/**
 * Checks if the source is a BMP image
 */
static gboolean
vips_foreign_load_bmp_source_is_a_source(VipsSource *source)
{
  unsigned char *bbuf = vips_source_sniff(source, 2);
  if (!bbuf) {
    vips_error("vips_foreign_load_bmp_source_is_a_source", "unable to sniff source");
    return 0;
  }

  return bbuf[0] == 'B' &&
      bbuf[1] == 'M';
}

/**
 * Loads the header of the BMP image from the source.
 */
static int
vips_foreign_load_bmp_header(VipsForeignLoad *load)
{
  VipsForeignLoadBmp *bmp = (VipsForeignLoadBmp *) load;

  // Rewind the source to the beginning
  if (vips_source_rewind(bmp->source))
    return -1;

  VipsPel file_header_buf[BMP_FILE_HEADER_LEN + 4];

  // Read the header + the next uint32 after
  if (vips_foreign_load_read_full(bmp->source, &file_header_buf, BMP_FILE_HEADER_LEN + 4) <= 0) {
    vips_error("vips_foreign_load_bmp_header", "unable to read file header from the source");
    return -1;
  }

  // Check if the info header length is valid
  uint32_t offset = GUINT32_FROM_LE(*(uint32_t *) (file_header_buf + 10));
  uint32_t info_header_len = GUINT32_FROM_LE(*(uint32_t *) (file_header_buf + 14));

  if (
      (info_header_len != BMP_BITMAP_INFO_HEADER_LEN) &&
      (info_header_len != BMP_V4_INFO_HEADER_LEN) &&
      (info_header_len != BMP_V5_INFO_HEADER_LEN)) {
    vips_error("vips_foreign_load_bmp_header", "incorrect BMP header length");
    return -1;
  }

  // Now, read the info header. -4 bytes is because we've already read the first 4 bytes of
  // the header (info_header_len) at the previous step. Constants include those 4 bytes.
  VipsPel *info_header = VIPS_ARRAY(load, info_header_len - 4, VipsPel);

  if (vips_foreign_load_read_full(bmp->source, info_header, info_header_len - 4) <= 0) {
    vips_error("vips_foreign_load_bmp_header", "unable to read BMP info header");
    return -1;
  }

  int32_t width = GINT32_FROM_LE(*(int32_t *) (info_header));
  int32_t height = GINT32_FROM_LE(*(int32_t *) (info_header + 4));
  uint16_t planes = GUINT16_FROM_LE(*(uint16_t *) (info_header + 8));
  uint16_t bpp = GUINT16_FROM_LE(*(uint16_t *) (info_header + 10));
  uint32_t compression = GUINT32_FROM_LE(*(uint32_t *) (info_header + 12));
  uint32_t num_colors = GUINT32_FROM_LE(*(uint32_t *) (info_header + 28));
  bool top_down = FALSE;
  bool rle = FALSE;
  bool bmp565 = FALSE;
  uint32_t rmask = 0;
  uint32_t gmask = 0;
  uint32_t bmask = 0;
  uint32_t amask = 0;
  int bands = 3; // 3 bands by default (RGB)

  // Let's determine if the image has an alpha channel
  bool has_alpha = bpp == 32;

  // If the info header is V4 or V5, check for alpha channel mask explicitly.
  // If it's non-zero, then the target image should have an alpha channel.
  if ((has_alpha) && (info_header_len > BMP_BITMAP_INFO_HEADER_LEN)) {
    has_alpha = GUINT32_FROM_LE(*(uint32_t *) (info_header + 48)) != 0;
  }

  // Target image should have alpha channel only in case source image has alpha channel
  // AND source image alpha mask is not zero
  if (has_alpha) {
    bands = 4;
  }

  // Source image bytes per pixel. It does not depend on the alpha mask, just on the bpp.
  int bytes_per_pixel = bpp >= 8 ? bpp / 8 : 1; // bytes per pixel in the source image

  if (height < 0) {
    height = -height;
    top_down = TRUE;
  }

  if ((width <= 0) || (height <= 0)) {
    vips_error("vips_foreign_load_bmp_header", "unsupported BMP image dimensions");
    return -1;
  }

  // we only support 1 plane and 8, 24 or 32 bits per pixel
  if (planes != 1) {
    vips_error("vips_foreign_load_bmp_header", "unsupported BMP image: planes != 1");
    return -1;
  }

  if (compression == COMPRESSION_BI_RGB) {
    // go ahead: no compression
  }
  else if (
      ((compression == COMPRESSION_BI_RLE8) && (bpp == 8)) ||
      ((compression == COMPRESSION_BI_RLE4) && (bpp == 4))) {
    // rle compression is used for 8-bit or 4-bit images
    rle = TRUE;
  }
  else if (
      (compression == COMPRESSION_BI_BITFIELDS) ||
      (compression == COMPRESSION_BI_BITFIELDS_ALPHA)) {
    int color_mask_len = 3;

    if (bpp > 24) {
      color_mask_len = 4;
    }

    uint32_t color_mask_buf[4];
    uint32_t *color_mask;

    // for the non-v4/v5 bmp image we need to load color mask separately since
    // it is not included in the header
    if (info_header_len == BMP_BITMAP_INFO_HEADER_LEN) {
      // let's attach it to load itself so we won't care about conditionally freeing it
      if (vips_foreign_load_read_full(bmp->source, color_mask_buf, color_mask_len * sizeof(uint32_t)) <= 0) {
        vips_error("vips_foreign_load_bmp_header", "unable to read BMP color mask");
        return -1;
      }
      color_mask = color_mask_buf;
    }
    else {
      // In case of v4/v5 info header, the color mask is already included in the info header,
      // we just need to read it
      color_mask = (uint32_t *) ((VipsPel *) info_header + 36);
    }

    // Standard says that color masks are in BE order. However, we do all the
    // checks and calculations as like as masks were in LE order.
    rmask = GUINT32_FROM_LE(color_mask[0]);
    gmask = GUINT32_FROM_LE(color_mask[1]);
    bmask = GUINT32_FROM_LE(color_mask[2]);
    amask = 0; // default alpha mask is 0

    if (color_mask_len > 3) {
      amask = GUINT32_FROM_LE(color_mask[3]);
    }

    if ((bpp == 16) && (rmask = 0xF800) && (gmask == 0x7E0) && (bmask == 0x1F)) {
      bmp565 = TRUE;
    }
    else if ((bpp == 16) && (rmask == 0x7C00) && (gmask == 0x3E0) && (bmask == 0x1F)) {
      // Go ahead, it's a regular 16 bit image
    }
    else if ((bpp == 32) && (rmask == 0xff0000) && (gmask == 0xff00) && (bmask == 0xff) && (amask == 0xff000000)) {
      // Go ahead, it's a regular 32-bit image
    }
    else {
      vips_error("vips_foreign_load_bmp_header", "unsupported BMP image: unsupported color masks");
      return -1;
    }
  }
  else {
    vips_error("vips_foreign_load_bmp_header", "unsupported BMP image: compression not supported");
    return -1;
  }

  // BMP images with 1, 2, 4 or 8 bits per pixel
  if (bpp <= 8) {
    // They could not have num_colors == 0, this is a bug in the BMP file.
    if (num_colors == 0) {
      num_colors = 1 << bpp;
    }

    // Please note, that BMP images are stored in BGR order rather than RGB order.
    // Every 4th byte is padding.
    bmp->palette = VIPS_MALLOC(load, num_colors * BMP_PALETTE_ITEM_SIZE);

    if (vips_foreign_load_read_full(bmp->source, bmp->palette, num_colors * BMP_PALETTE_ITEM_SIZE) <= 0) {
      vips_error("vips_foreign_load_bmp_header", "unable to read BMP palette");
      return -1;
    }
  }

  bmp->width = width;
  bmp->height = height;
  bmp->planes = planes;
  bmp->bpp = bpp;
  bmp->compression = compression;
  bmp->offset = offset;
  bmp->top_down = top_down;
  bmp->rle = rle;
  bmp->bmp565 = bmp565;
  bmp->num_colors = num_colors;
  bmp->bands = bands;

  bmp->rmask = rmask;
  bmp->gmask = gmask;
  bmp->bmask = bmask;
  bmp->amask = amask;

  bmp->bytes_per_pixel = bytes_per_pixel;
  bmp->y_pos = 0;

  bmp->dx = 0; // In sequential access this indicates that we need to skip n lines
  bmp->dy = 0; // n pixels

  // set the image header of the out image
  if (vips_foreign_load_bmp_set_image_header(bmp, load->out)) {
    return -1;
  }

  // seek to the beginning of image data
  if (vips_source_seek(bmp->source, offset, SEEK_SET) < 0) {
    vips_error("vips_foreign_load_bmp_header", "unable to seek to BMP image data");
    return -1;
  }

  vips_source_minimise(bmp->source);

  return 0;
}

/**
 * Generates a strip for 24/32 bpp BMP image.
 */
static int
vips_foreign_load_bmp_24_32_generate_strip(VipsRect *r, VipsRegion *out_region, VipsForeignLoadBmp *bmp)
{
  // Align the row size to 4 bytes, as BMP rows are 4-byte aligned.
  int row_size = (bmp->bytes_per_pixel * r->width + 3) & (~3);

  VipsPel *src;
  VipsPel *dest;

  for (int y = 0; y < r->height; y++) {
    src = bmp->row_buffer;
    dest = BMP_ROW_ADDR(bmp, out_region, r, y);

    if (vips_foreign_load_read_full(bmp->source, src, row_size) <= 0) {
      vips_error("vips_foreign_load_bmp_24_32_generate_strip", "failed to read raw data");
      return -1;
    }

    for (int x = 0; x < r->width; x++) {
      dest[0] = src[2]; // B
      dest[1] = src[1]; // G
      dest[2] = src[0]; // R

      // if the image has alpha channel, copy it too
      if (bmp->bands == 4) {
        dest[3] = src[3]; // A
      }

      dest += bmp->bands;
      src += bmp->bytes_per_pixel;
    }

    bmp->y_pos += 1;
  }

  return 0;
}

/**
 * Generates a strip for 16 bpp BMP image.
 */
static int
vips_foreign_load_bmp_16_generate_strip(VipsRect *r, VipsRegion *out_region, VipsForeignLoadBmp *bmp)
{
  // Align the row size to 4 bytes, as BMP rows are 4-byte aligned, 16 bpp = 2 bytes per pixel
  int row_size = (bmp->bytes_per_pixel * r->width + 3) & (~3);

  VipsPel *src;
  VipsPel *dest;

  for (int y = 0; y < r->height; y++) {
    src = bmp->row_buffer;
    dest = BMP_ROW_ADDR(bmp, out_region, r, y);

    if (vips_foreign_load_read_full(bmp->source, src, row_size) <= 0) {
      vips_error("vips_foreign_load_bmp_16_generate_strip", "failed to read raw data");
      return -1;
    }

    for (int x = 0; x < r->width; x++) {
      uint16_t pixel = GUINT16_FROM_LE(*(uint16_t *) src);

      // 565 and non-565 formats both are handled here: they differ by the masks
      dest[0] = (uint8_t) ((pixel & bmp->rmask) >> 11) << 3;
      dest[1] = (uint8_t) ((pixel & bmp->gmask) >> 5) << 2;
      dest[2] = (uint8_t) (pixel & bmp->bmask) << 3;

      dest += bmp->bands;
      src += bmp->bytes_per_pixel; // 2 bytes per pixel for 16 bpp
    }

    bmp->y_pos += 1;
  }

  return 0;
}

/**
 * Writes pixels for 1/2/4/8 bpp BMP image using palette. Pixels are taken from the src (if present), or src_byte (RLE case).
 */
void
vips_foreign_load_bpp_1_8_write_pixels_palette(VipsForeignLoadBmp *bmp, VipsPel *dest, VipsPel *src, int width, VipsPel src_byte)
{
  int bit = 8 - bmp->bpp;
  int src_offset = 0;

  for (int x = 0; x < width; x++) {
    // Read the palette index from the source
    int pixel;

    if (src != NULL) {
      pixel = (int) src[src_offset] >> bit;
    }
    else {
      pixel = src_byte >> bit;
    }

    int mask = (1 << bmp->bpp) - 1;
    int palette_index = pixel & mask;

    int dest_offset = x * bmp->bands;

    if (bit == 0) {
      bit = 8 - bmp->bpp;
      src_offset++;
    }
    else {
      bit -= bmp->bpp;
    }

    VipsPel *color = (VipsPel *) &bmp->palette[palette_index];

    dest[dest_offset + 0] = color[2]; // BGR, reversed
    dest[dest_offset + 1] = color[1];
    dest[dest_offset + 2] = color[0];
  }
}

/**
 * Generates a strip for 1/2/4/8 bpp BMP image.
 */
static int
vips_foreign_load_bmp_1_8_generate_strip(VipsRect *r, VipsRegion *out_region, VipsForeignLoadBmp *bmp)
{
  // Align the row size to 4 bytes, as BMP rows are 4-byte aligned
  char cap = 8 / bmp->bpp;
  int row_size = ((r->width + cap - 1) / cap + 3) & (~3);

  VipsPel *src = bmp->row_buffer; // just a shortcut
  VipsPel *dest;

  for (int y = 0; y < r->height; y++) {
    if (vips_foreign_load_read_full(bmp->source, src, row_size) <= 0) {
      vips_error("vips_foreign_load_bmp_16_generate_strip", "failed to read raw data");
      return -1;
    }

    dest = BMP_ROW_ADDR(bmp, out_region, r, y);

    vips_foreign_load_bpp_1_8_write_pixels_palette(bmp, dest, src, r->width, -1);

    bmp->y_pos += 1;
  }

  return 0;
}

/**
 * Generates a strip for 1/2/4/8 bpp BMP image.
 *
 * BMP RLE is encoded per-line (so, each line has 0x00 00 - LE control byte at the end).
 */
static int
vips_foreign_load_bmp_rle_generate_strip(VipsRect *r, VipsRegion *out_region, VipsForeignLoadBmp *bmp)
{
  // Align the row size to 4 bytes, as BMP rows are 4-byte aligned
  char cap = 8 / bmp->bpp;

  VipsPel *src = bmp->row_buffer; // just a shortcut
  VipsPel *dest;
  VipsPel cmd[2];
  VipsPel dxdy[2];

  for (int y = 0; y < r->height; y++) {
    dest = BMP_ROW_ADDR(bmp, out_region, r, y);

    // fill the line with zeros (move to skips)
    memset(dest, 0, r->width * bmp->bands);

    // Skip lines if needed, this might be the whole region
    if (bmp->dy > 0) {
      bmp->dy--;
      bmp->y_pos += 1;
      continue;
    }

    int x = 0;

    if (bmp->dx > 0) {
      x = bmp->dx;
      bmp->dx = 0;
    }

    do {
      // Read next command
      //
      // NOTE: This might not be very efficient, unless underlying vips buffer is memory-buffered
      if (vips_foreign_load_read_full(bmp->source, &cmd, 2) <= 0) {
        vips_error("vips_foreign_load_bmp_rle_generate_strip", "failed to read next RLE command");
        return -1;
      }

      // Check control byte
      if (cmd[0] == 0) {
        if (cmd[1] == BMP_RLE_EOL) {
          break;
        }

        else if (cmd[1] == BMP_RLE_EOF) {
          bmp->dy = G_MAXUINT32; // set dy to max, so all the leftover lines will be skipped
          bmp->dx = 0;

          break; // exit the loop, we reached EOF
        }
        else if (cmd[1] == BMP_RLE_MOVE_TO) {
          if (vips_foreign_load_read_full(bmp->source, &dxdy, 2) <= 0) {
            vips_error("vips_foreign_load_bmp_rle_generate_strip", "failed to read RLE move command");
            return -1;
          }

          int dx = dxdy[0]; // relative X offset
          int dy = dxdy[1]; // relative Y offset

          // We treat movement by Y as EOL
          if (dy > 0) {
            bmp->dx = MIN(x + dx, r->width); // New X position must not exceed the width of the image
            bmp->dy = dy;                    // We do not care if Y pos is outside of the impage, it's a separate check

            break; // we need to skip lines, so we exit the loop
          } // Movement by X might not lead to EOL, so we continue
          else {
            bmp->dy = dy;              // 0
            x = MIN(x + dx, r->width); // Move to the desired pixel
          }
        }
        else { // Directly read next n bytes
          int pixels_count = cmd[1];
          int bytes_count = ((pixels_count + cap - 1) / cap + 1) & ~1;

          pixels_count = MIN(pixels_count, r->width - x);

          if (vips_foreign_load_read_full(bmp->source, src, bytes_count) <= 0) {
            vips_error("vips_foreign_load_bmp_rle_generate_strip", "failed to read RLE data");
            return -1;
          }

          vips_foreign_load_bpp_1_8_write_pixels_palette(bmp, dest + (x * bmp->bands), src, pixels_count, -1);

          x += pixels_count;
        }
      }
      else { // read RLE-encoded pixels
        VipsPel pixels_count = cmd[0];
        VipsPel pixel = cmd[1];

        pixels_count = MIN(pixels_count, r->width - x);

        vips_foreign_load_bpp_1_8_write_pixels_palette(bmp, dest + (x * bmp->bands), NULL, pixels_count, pixel);

        x += pixels_count;
      }
    } while (1);

    bmp->y_pos += 1; // Move to the next line
  }

  return 0;
}

/**
 * Loads a strip of non-rle bmp image, access must be sequential, demand
 * style must be thinstrip + strip height set to max.
 */
static int
vips_foreign_load_bmp_rgb_generate(VipsRegion *out_region,
    void *seq, void *a, void *b, gboolean *stop)
{
  VipsForeignLoadBmp *bmp = (VipsForeignLoadBmp *) a;
  VipsRect *r = &out_region->valid;

  /**
   * Sanity checks which assure that the requested region has the correct shape.
   *
   * We use sequential access + thinstrip demand style which means that image would
   * be read in strips, where each strip represents image-wide set of rows.
   */
  g_assert(r->left == 0);                                 // Strip starts at the left edge of the image
  g_assert(r->width == out_region->im->Xsize);            // Has width of the image
  g_assert(VIPS_RECT_BOTTOM(r) <= out_region->im->Ysize); // Equals or less of image height

  // Equals to the maximum height of the strip or less (last strip)
  if (bmp->top_down)
    g_assert(r->height ==
        VIPS_MIN(VIPS__FATSTRIP_HEIGHT, out_region->im->Ysize - r->top));
  else
    g_assert(r->height == out_region->im->Ysize);

  // Check if the requested strip is in order
  if (r->top != bmp->y_pos) {
    vips_error("vips_foreign_load_bmp_generate", "out of order read at line %d", bmp->y_pos);
    return -1;
  }

  if (bmp->rle) {
    return vips_foreign_load_bmp_rle_generate_strip(r, out_region, bmp);
  }
  else if (bmp->bpp >= 24) {
    return vips_foreign_load_bmp_24_32_generate_strip(r, out_region, bmp);
  }
  else if (bmp->bpp == 16) {
    return vips_foreign_load_bmp_16_generate_strip(r, out_region, bmp);
  }

  return vips_foreign_load_bmp_1_8_generate_strip(r, out_region, bmp);
}

/**
 * Loads a BMP image from the source.
 */
static int
vips_foreign_load_bmp_load(VipsForeignLoad *load)
{
  VipsForeignLoadBmp *bmp = (VipsForeignLoadBmp *) load;

  // For a case when we encounter buggy BMP image which has RLE command to read next
  // 255 bytes, and our buffer is smaller than that, we need it to be at least 255 bytes.
  int row_buffer_length = (bmp->width * 4) + 4;
  if (row_buffer_length < 255) {
    row_buffer_length = 255;
  }

  // Allocate a row buffer for the current row in all generate* functions.
  // 4 * width + 4 is guaranteed to be enough for the longest (32-bit per pixel) row + padding.
  bmp->row_buffer = VIPS_ARRAY(load, row_buffer_length, VipsPel);

  VipsImage **t = (VipsImage **)
      vips_object_local_array(VIPS_OBJECT(load), 2);

  t[0] = vips_image_new();

  if (
      vips_foreign_load_bmp_set_image_header(bmp, t[0]) ||
      vips_image_generate(t[0],
          NULL, vips_foreign_load_bmp_rgb_generate, NULL,
          bmp, NULL) ||
      // For bottom-up BMP images we need to flip the image vertically.
      // We do this in vips_image_generate callback, so we need to be sure that
      // we generate regions size of the whole image.
      vips_sequential(t[0], &t[1],
          "tile_height", bmp->top_down ? VIPS__FATSTRIP_HEIGHT : t[0]->Ysize,
          NULL) ||
      vips_image_write(t[1], load->real) ||
      vips_source_decode(bmp->source)) {
    return -1;
  }

  return 0;
}

static void
vips_foreign_load_bmp_class_init(VipsForeignLoadBmpClass *class)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(class);
  VipsObjectClass *object_class = (VipsObjectClass *) class;
  VipsForeignClass *foreign_class = (VipsForeignClass *) class;
  VipsForeignLoadClass *load_class = (VipsForeignLoadClass *) class;

  gobject_class->dispose = vips_foreign_load_bmp_dispose;
  gobject_class->set_property = vips_object_set_property;
  gobject_class->get_property = vips_object_get_property;

  object_class->nickname = "bmpload_base";
  object_class->description = "load bmp image (internal format for video thumbs)";
  object_class->build = vips_foreign_load_bmp_build;

  load_class->get_flags = vips_foreign_load_bmp_get_flags;
  load_class->header = vips_foreign_load_bmp_header;
  load_class->load = vips_foreign_load_bmp_load;
}

static void
vips_foreign_load_bmp_init(VipsForeignLoadBmp *load)
{
  load->palette = NULL;
}

typedef struct _VipsForeignLoadBmpSource {
  VipsForeignLoadBmp parent_object;

  VipsSource *source;
} VipsForeignLoadBmpSource;

typedef VipsForeignLoadBmpClass VipsForeignLoadBmpSourceClass;

G_DEFINE_TYPE(VipsForeignLoadBmpSource, vips_foreign_load_bmp_source,
    vips_foreign_load_bmp_get_type());

static int
vips_foreign_load_bmp_source_build(VipsObject *object)
{
  VipsForeignLoadBmp *bmp = (VipsForeignLoadBmp *) object;
  VipsForeignLoadBmpSource *source =
      (VipsForeignLoadBmpSource *) object;

  if (source->source) {
    bmp->source = source->source;
    g_object_ref(bmp->source);
  }

  return VIPS_OBJECT_CLASS(vips_foreign_load_bmp_source_parent_class)
      ->build(object);
}

static void
vips_foreign_load_bmp_source_class_init(
    VipsForeignLoadBmpSourceClass *class)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(class);
  VipsObjectClass *object_class = (VipsObjectClass *) class;
  VipsOperationClass *operation_class = VIPS_OPERATION_CLASS(class);
  VipsForeignLoadClass *load_class = (VipsForeignLoadClass *) class;

  gobject_class->set_property = vips_object_set_property;
  gobject_class->get_property = vips_object_get_property;

  load_class->is_a_source = vips_foreign_load_bmp_source_is_a_source;

  object_class->nickname = "bmpload_source";
  object_class->description = "load image from bmp source";
  object_class->build = vips_foreign_load_bmp_source_build;

  VIPS_ARG_OBJECT(class, "source", 1,
      "Source",
      "Source to load from",
      VIPS_ARGUMENT_REQUIRED_INPUT,
      G_STRUCT_OFFSET(VipsForeignLoadBmpSource, source),
      VIPS_TYPE_SOURCE);
}

static void
vips_foreign_load_bmp_source_init(VipsForeignLoadBmpSource *source)
{
}

/**
 * vips_bmpload_source:
 * @source: source to load
 * @out: (out): image to write
 * @...: `NULL`-terminated list of optional named arguments
 *
 * Read a RAWRGB-formatted memory block into a VIPS image.
 *
 * Returns: 0 on success, -1 on error.
 */
int
vips_bmpload_source(VipsSource *source, VipsImage **out, ...)
{
  va_list ap;
  int result;

  va_start(ap, out);
  result = vips_call_split("bmpload_source", ap, source, out);
  va_end(ap);

  return result;
}

// wrapper function which hiders varargs (...) from CGo
int
vips_bmpload_source_go(VipsImgproxySource *source, VipsImage **out)
{
  return vips_bmpload_source(
      VIPS_SOURCE(source), out,
      "access", VIPS_ACCESS_SEQUENTIAL,
      NULL);
}
