#include <stdlib.h>
#include <stdint.h> // uintptr_t

#include <vips/vips.h>
#include <vips/connection.h>

#ifndef VIPS_IMGPROXY_SOURCE_H
#define VIPS_IMGPROXY_SOURCE_H

// vips async source
typedef struct _VipsImgproxySource {
  VipsSource source;      // class designator
  uintptr_t readerHandle; // async reader handler
} VipsImgproxySource;

// glib class for vips async source
typedef struct _VipsImgproxySourceClass {
  VipsSourceClass parent_class;
} VipsImgproxySourceClass;

// creates new vips async source from a reader handle
VipsImgproxySource *vips_new_imgproxy_source(uintptr_t readerHandle);

#endif

// unreferences the source, which leads to reader close
void unref_imgproxy_source(VipsImgproxySource *source);
