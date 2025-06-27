#include <stdlib.h>
#include <stdint.h> // uintptr_t

#include <vips/vips.h>
#include <vips/connection.h>

// vips async source
typedef struct _VipsAsyncSource {
  VipsSource source;      // class designator
  uintptr_t readerHandle; // async reader handler
} VipsAsyncSource;

// glib class for vips async source
typedef struct _VipsAsyncSourceClass {
  VipsSourceClass parent_class;
} VipsAsyncSourceClass;

// vips async source read functions
int vips_jpegload_source_go(VipsAsyncSource *source, int shrink, VipsImage **out);
int vips_jxlload_source_go(VipsAsyncSource *source, int pages, VipsImage **out);
int vips_pngload_source_go(VipsAsyncSource *source, VipsImage **out, int unlimited);
int vips_webpload_source_go(VipsAsyncSource *source, double scale, int pages, VipsImage **out);
int vips_gifload_source_go(VipsAsyncSource *source, int pages, VipsImage **out);
int vips_svgload_source_go(VipsAsyncSource *source, double scale, VipsImage **out, int unlimited);
int vips_heifload_source_go(VipsAsyncSource *source, VipsImage **out, int thumbnail);
int vips_tiffload_source_go(VipsAsyncSource *source, VipsImage **out);

// creates new vips async source from a reader handle
VipsAsyncSource *vips_new_async_source(uintptr_t readerHandle);

// unreferences the source, which leads to reader close
void unref_source(VipsAsyncSource *source);
