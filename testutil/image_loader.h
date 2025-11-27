#include <stdlib.h>
#include <stdint.h> // uintptr_t

#include <vips/vips.h>

#ifndef TESTUTIL_IMAGE_LOADER_H
#define TESTUTIL_IMAGE_LOADER_H

// Function to read VipsImage as RGBA into memory buffer
int vips_image_read_from_to_memory(
    void *in, size_t in_size,       // inner raw buffer and its size
    void **out, size_t *out_size,   // out raw buffer an its size
    int *out_width, int *out_height // out image width and height
);

#endif
