#include <stdlib.h>
#include <stdint.h> // uintptr_t

#include <vips/vips.h>

#ifndef TESTUTIL_IMAGE_HASH_H
#define TESTUTIL_IMAGE_HASH_H

// Function to read VipsImage as RGBA into memory buffer
int vips_image_read_to_memory(VipsImage *in, void **buf, size_t *size);

// Function to free/discard the memory buffer
void vips_memory_buffer_free(void *buf);

#endif
