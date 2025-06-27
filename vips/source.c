#include "source.h"

// --- async source ----------------------------------------------------------------------

// define glib subtype for vips async source
#define VIPS_TYPE_ASYNC_SOURCE (vips_async_source_get_type())
G_DEFINE_FINAL_TYPE(VipsAsyncSource, vips_async_source, VIPS_TYPE_SOURCE)

extern void closeAsyncReader(uintptr_t handle);
extern gint64 asyncReaderSeek(uintptr_t handle, gint64 offset, int whence);
extern gint64 asyncReaderRead(uintptr_t handle, gpointer buffer, gint64 size);

// loads jpeg from a source
int
vips_jpegload_source_go(VipsAsyncSource *source, int shrink, VipsImage **out)
{
  if (shrink > 1)
    return vips_jpegload_source(VIPS_SOURCE(source), out, "shrink", shrink,
        NULL);

  return vips_jpegload_source(VIPS_SOURCE(source), out, NULL);
}

// loads xjl from source
int
vips_jxlload_source_go(VipsAsyncSource *source, int pages, VipsImage **out)
{
  return vips_jxlload_source(VIPS_SOURCE(source), out, "n", pages, NULL);
}

int
vips_pngload_source_go(VipsAsyncSource *source, VipsImage **out, int unlimited)
{
  return vips_pngload_source(
      VIPS_SOURCE(source), out,
      "unlimited", unlimited,
      NULL);
}

int
vips_webpload_source_go(VipsAsyncSource *source, double scale, int pages, VipsImage **out)
{
  return vips_webpload_source(
      VIPS_SOURCE(source), out,
      "scale", scale,
      "n", pages,
      NULL);
}

int
vips_gifload_source_go(VipsAsyncSource *source, int pages, VipsImage **out)
{
  return vips_gifload_source(VIPS_SOURCE(source), out, "access", VIPS_ACCESS_SEQUENTIAL, "n", pages, NULL);
}

int
vips_svgload_source_go(VipsAsyncSource *source, double scale, VipsImage **out, int unlimited)
{
  // libvips limits the minimal scale to 0.001, so we have to scale down dpi
  // for lower scale values
  double dpi = 72.0;
  if (scale < 0.001) {
    dpi *= VIPS_MAX(scale / 0.001, 0.001);
    scale = 0.001;
  }

  return vips_svgload_source(
      VIPS_SOURCE(source), out,
      "scale", scale,
      "dpi", dpi,
      "unlimited", unlimited,
      NULL);
}

int
vips_heifload_source_go(VipsAsyncSource *source, VipsImage **out, int thumbnail)
{
  return vips_heifload_source(
      VIPS_SOURCE(source), out,
      "access", VIPS_ACCESS_SEQUENTIAL,
      "thumbnail", thumbnail,
      NULL);
}

int
vips_tiffload_source_go(VipsAsyncSource *source, VipsImage **out)
{
  return vips_tiffload_source(VIPS_SOURCE(source), out, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
}

// dereferences source
void
unref_source(VipsAsyncSource *source)
{
  VIPS_UNREF(source);
}

// read function for vips async source
static gint64
vips_async_source_read(VipsSource *source, void *buffer, size_t length)
{
  VipsAsyncSource *self = (VipsAsyncSource *) source;

  gint64 read_length = asyncReaderRead(self->readerHandle, buffer, length);
  if (read_length < 0) {
    vips_error("vips_async_source_read", "failed to read from async source");
  }
  return read_length;
}

// seek function for vips async source. whence can be SEEK_SET (0), SEEK_CUR (1), or SEEK_END (2).
static gint64
vips_async_source_seek(VipsSource *source, gint64 offset, int whence)
{
  VipsAsyncSource *self = (VipsAsyncSource *) source;

  gint64 actual_offset = asyncReaderSeek(self->readerHandle, offset, whence);

  if (actual_offset < 0) {
    vips_error("vips_async_source_seek", "failed to seek in async source");
  }

  return actual_offset;
}

static void
vips_async_source_dispose(GObject *gobject)
{
  VipsAsyncSource *source = (VipsAsyncSource *) gobject;

  closeAsyncReader(source->readerHandle);

  G_OBJECT_CLASS(vips_async_source_parent_class)->dispose(gobject);
}

// attaches seek/read handlers to the async source class
static void
vips_async_source_class_init(VipsAsyncSourceClass *klass)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(klass);
  VipsObjectClass *object_class = VIPS_OBJECT_CLASS(klass);
  VipsSourceClass *source_class = VIPS_SOURCE_CLASS(klass);

  object_class->nickname = "async_source";
  object_class->description = "async input source";

  gobject_class->dispose = vips_async_source_dispose;

  source_class->read = vips_async_source_read;
  source_class->seek = vips_async_source_seek;
}

// initializes the async source (nothing to do here yet)
static void
vips_async_source_init(VipsAsyncSource *source)
{
}

// creates a new async source with the given reader handle
VipsAsyncSource *
vips_new_async_source(uintptr_t readerHandle)
{
  VipsAsyncSource *source = g_object_new(VIPS_TYPE_ASYNC_SOURCE, NULL);
  source->readerHandle = readerHandle;
  return source;
}
