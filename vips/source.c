#include "source.h"

// define glib subtype for vips async source
#define VIPS_TYPE_IMGPROXY_SOURCE (vips_imgproxy_source_get_type())
G_DEFINE_FINAL_TYPE(VipsImgproxySource, vips_imgproxy_source, VIPS_TYPE_SOURCE)

extern void closeImgproxyReader(uintptr_t handle);
extern gint64 imgproxyReaderSeek(uintptr_t handle, gint64 offset, int whence);
extern gint64 imgproxyReaderRead(uintptr_t handle, gpointer buffer, gint64 size);

// dereferences source
void
unref_imgproxy_source(VipsImgproxySource *source)
{
  VIPS_UNREF(source);
}

// read function for vips imgproxy source
static gint64
vips_imgproxy_source_read(VipsSource *source, void *buffer, size_t length)
{
  VipsImgproxySource *self = (VipsImgproxySource *) source;

  return imgproxyReaderRead(self->readerHandle, buffer, length);
}

// seek function for vips imgproxy source. whence can be SEEK_SET (0), SEEK_CUR (1), or SEEK_END (2).
static gint64
vips_imgproxy_source_seek(VipsSource *source, gint64 offset, int whence)
{
  VipsImgproxySource *self = (VipsImgproxySource *) source;

  return imgproxyReaderSeek(self->readerHandle, offset, whence);
}

static void
vips_imgproxy_source_dispose(GObject *gobject)
{
  VipsImgproxySource *source = (VipsImgproxySource *) gobject;

  closeImgproxyReader(source->readerHandle);

  G_OBJECT_CLASS(vips_imgproxy_source_parent_class)->dispose(gobject);
}

// attaches seek/read handlers to the imgproxy source class
static void
vips_imgproxy_source_class_init(VipsImgproxySourceClass *klass)
{
  GObjectClass *gobject_class = G_OBJECT_CLASS(klass);
  VipsObjectClass *object_class = VIPS_OBJECT_CLASS(klass);
  VipsSourceClass *source_class = VIPS_SOURCE_CLASS(klass);

  object_class->nickname = "imgproxy_source";
  object_class->description = "imgproxy input source";

  gobject_class->dispose = vips_imgproxy_source_dispose;

  source_class->read = vips_imgproxy_source_read;
  source_class->seek = vips_imgproxy_source_seek;
}

// initializes the imgproxy source (nothing to do here yet)
static void
vips_imgproxy_source_init(VipsImgproxySource *source)
{
}

// creates a new imgproxy source with the given reader handle
VipsImgproxySource *
vips_new_imgproxy_source(uintptr_t readerHandle)
{
  VipsImgproxySource *source = g_object_new(VIPS_TYPE_IMGPROXY_SOURCE, NULL);
  source->readerHandle = readerHandle;
  return source;
}
