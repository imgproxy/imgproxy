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

// vips async source read function
int vips_jpegloadsource_go(VipsAsyncSource *source, int shrink, VipsImage **out);

// creates new vips async source from a reader handle
VipsAsyncSource *vips_new_async_source(uintptr_t readerHandle);

// unreferences the source, which leads to reader close
void unref_source(VipsAsyncSource *source);
