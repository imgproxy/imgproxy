#pragma once

#include <stdlib.h>

int
vips_dct2_hash(void *in_buf, size_t in_buf_size, float **dct_array, size_t *length);
