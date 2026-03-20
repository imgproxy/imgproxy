#pragma once

#include "vips.h"

/* vips_dct2_hash calculates the DCT hash of the image in the input buffer
 * and returns it as an array of floats.
 */
int
vips_dct2_hash(void *in_buf, size_t in_buf_size, float **dct_array, size_t *length);
