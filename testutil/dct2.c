#include <vips/vips.h>

#define DCT_SIZE 8
#define TARGET_SIZE 64

// Calculate DCT II for every pixel in the image (all channels)
// Extract low-frequency DCT coefficients (DCT_SIZE x DCT_SIZE) from the full image
// <https://en.wikipedia.org/wiki/Discrete_cosine_transform>
void
calc_raw_dct2(float *in, float *out, int width, int height, int channels)
{
  for (int c = 0; c < channels; c++) {
    for (int v = 0; v < DCT_SIZE; v++) {
      for (int u = 0; u < DCT_SIZE; u++) {
        double sum = 0.0;

        for (int y = 0; y < height; y++) {
          for (int x = 0; x < width; x++) {
            float pixel = in[(y * width + x) * channels + c];

            sum += pixel *
                cosf((2.0 * x + 1.0) * u * VIPS_PI / (2.0 * width)) *
                cosf((2.0 * y + 1.0) * v * VIPS_PI / (2.0 * height));
          }
        }

        double alpha_u = (u == 0) ? 1.0 / sqrt((double) width) : sqrt(2.0 / (double) width);
        double alpha_v = (v == 0) ? 1.0 / sqrt((double) height) : sqrt(2.0 / (double) height);

        out[c * DCT_SIZE * DCT_SIZE + v * DCT_SIZE + u] = (float) (alpha_u * alpha_v * sum);
      }
    }
  }
}

// Calculate DCT hash for an image buffer
int
vips_dct2_hash(void *in_buf, size_t in_buf_size, float **dct_array, size_t *length)
{
  VipsImage *base = vips_image_new();
  VipsImage **t = (VipsImage **) vips_object_local_array(VIPS_OBJECT(base), 6);

  // Load image from buffer
  t[0] = vips_image_new_from_buffer(in_buf, in_buf_size, "", NULL);
  if (t[0] == NULL) {
    VIPS_UNREF(base);
    return -1;
  }

  VipsImage *in = t[0];

  VipsInterpretation interp = vips_image_guess_interpretation(in);
  int is_bw = (interp == VIPS_INTERPRETATION_B_W || interp == VIPS_INTERPRETATION_GREY16);

  if (vips_colourspace(in, &t[1], VIPS_INTERPRETATION_scRGB, NULL)) {
    VIPS_UNREF(base);
    vips_error("vips_dct2_hash", "failed to convert to scRGB");
    return -1;
  }
  in = t[1];

  // Calculate resize proportions to fit TARGET_SIZE
  double wscale = VIPS_MIN((double) TARGET_SIZE / in->Xsize, 1.0);
  double hscale = VIPS_MIN((double) TARGET_SIZE / in->Ysize, 1.0);

  if (vips_resize(in, &t[2], wscale, "vscale", hscale, "kernel", VIPS_KERNEL_LANCZOS3, NULL)) {
    VIPS_UNREF(base);
    vips_error("vips_dct2_hash", "failed to resize image");
    return -1;
  }
  in = t[2];

  // Now, blend magenta background if alpha is present
  if (vips_image_hasalpha(in)) {
    VipsArrayDouble *bga = vips_array_double_newv(3, 1.0, 0.0, 1.0);

    int res = vips_flatten(in, &t[3], "background", bga, NULL);
    vips_area_unref((VipsArea *) bga);

    if (res) {
      VIPS_UNREF(base);
      return -1;
    }

    in = t[3];
  }

  // Convert to Lab color space for better perceptual hashing
  if (vips_colourspace(in, &t[4], VIPS_INTERPRETATION_LAB, NULL)) {
    VIPS_UNREF(base);
    vips_error("vips_dct2_hash", "failed to convert to Lab");
    return -1;
  }
  in = t[4];

  // For B/W images, we only need the lightness channel
  if (is_bw) {
    if (vips_extract_band(in, &t[5], 0, NULL)) {
      VIPS_UNREF(base);
      vips_error("vips_dct2_hash", "failed to extract band for B/W image");
      return -1;
    }
    in = t[5];
  }

  // Get image dimensions
  int width = in->Xsize;
  int height = in->Ysize;

  // Write image to memory buffer
  size_t size = 0;
  float *source = vips_image_write_to_memory(in, &size);
  if (size == 0) {
    vips_error("vips_dct2_hash", "unable to write image to memory");
    VIPS_UNREF(base);
    return -1;
  }

  // Allocate memory for DCT output (channels * 8x8)
  *dct_array = (float *) malloc(sizeof(float) * in->Bands * DCT_SIZE * DCT_SIZE);
  if (*dct_array == NULL) {
    VIPS_FREE(source);
    VIPS_UNREF(base);
    return -1;
  }

  // Calculate DCT II matrix for all channels
  calc_raw_dct2(source, *dct_array, width, height, in->Bands);

  VIPS_FREE(source);
  VIPS_UNREF(base);

  *length = in->Bands * DCT_SIZE * DCT_SIZE;

  return 0;
}
