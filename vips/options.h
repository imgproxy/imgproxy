#include <vips/vips.h>

typedef struct _ImgproxySaveOptions {
  gboolean JpegProgressive; // Whether to save JPEG as progressive.

  gboolean PngInterlaced;    // Whether to save PNG as interlaced.
  gboolean PngQuantize;      // Whether to quantize PNG (save with palette).
  int PngQuantizationColors; // Number of colors to use in PNG quantization.

  VipsForeignWebpPreset WebpPreset; // WebP preset to use.
  int WebpEffort;                   // WebP encoding effort level.

  int AvifSpeed; // AVIF encoding speed.

  int JxlEffort; // JPEG XL encoding effort.
} ImgproxySaveOptions;
