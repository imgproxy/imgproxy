package imagetype

var (
	JPEG = RegisterType(&TypeDesc{
		String:                "jpeg",
		Ext:                   ".jpg",
		Mime:                  "image/jpeg",
		IsVector:              false,
		SupportsAlpha:         false,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	JXL = RegisterType(&TypeDesc{
		String:                "jxl",
		Ext:                   ".jxl",
		Mime:                  "image/jxl",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: true,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	PNG = RegisterType(&TypeDesc{
		String:                "png",
		Ext:                   ".png",
		Mime:                  "image/png",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       false,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	WEBP = RegisterType(&TypeDesc{
		String:                "webp",
		Ext:                   ".webp",
		Mime:                  "image/webp",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: true,
		SupportsAnimationSave: true,
		SupportsThumbnail:     false,
	})

	GIF = RegisterType(&TypeDesc{
		String:                "gif",
		Ext:                   ".gif",
		Mime:                  "image/gif",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: false,
		SupportsQuality:       false,
		SupportsAnimationLoad: true,
		SupportsAnimationSave: true,
		SupportsThumbnail:     false,
	})

	ICO = RegisterType(&TypeDesc{
		String:                "ico",
		Ext:                   ".ico",
		Mime:                  "image/x-icon",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: false,
		SupportsQuality:       false,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	SVG = RegisterType(&TypeDesc{
		String:                "svg",
		Ext:                   ".svg",
		Mime:                  "image/svg+xml",
		IsVector:              true,
		SupportsAlpha:         true,
		SupportsColourProfile: false,
		SupportsQuality:       false,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	HEIC = RegisterType(&TypeDesc{
		String:                "heic",
		Ext:                   ".heic",
		Mime:                  "image/heif",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     true,
	})

	AVIF = RegisterType(&TypeDesc{
		String:                "avif",
		Ext:                   ".avif",
		Mime:                  "image/avif",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     true,
	})

	BMP = RegisterType(&TypeDesc{
		String:                "bmp",
		Ext:                   ".bmp",
		Mime:                  "image/bmp",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: false,
		SupportsQuality:       false,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	TIFF = RegisterType(&TypeDesc{
		String:                "tiff",
		Ext:                   ".tiff",
		Mime:                  "image/tiff",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})
)

// init registers default magic bytes for common image formats
func init() {
	// NOTE: we cannot be 100% sure of image type until we fully decode it. This is especially true
	// for "naked" jxl (0xff 0x0a). There is no other way to ensure this is a JXL file, except to fully
	// decode it. Two bytes are too few to reliably identify the format. The same applies to ICO.

	// JPEG magic bytes
	RegisterMagicBytes(JPEG, []byte("\xff\xd8"))

	// JXL magic bytes
	RegisterMagicBytes(JXL, []byte{0xff, 0x0a})                                                             // JXL codestream (can't use string due to 0x0a)
	RegisterMagicBytes(JXL, []byte{0x00, 0x00, 0x00, 0x0C, 0x4A, 0x58, 0x4C, 0x20, 0x0D, 0x0A, 0x87, 0x0A}) // JXL container (has null bytes)

	// PNG magic bytes
	RegisterMagicBytes(PNG, []byte("\x89PNG\r\n\x1a\n"))

	// WEBP magic bytes (RIFF container with WEBP fourcc) - using wildcard for size
	RegisterMagicBytes(WEBP, []byte("RIFF????WEBP"))

	// GIF magic bytes
	RegisterMagicBytes(GIF, []byte("GIF8?a"))

	// ICO magic bytes
	RegisterMagicBytes(ICO, []byte{0, 0, 1, 0}) // ICO (has null bytes)

	// HEIC/HEIF magic bytes with wildcards for size
	RegisterMagicBytes(HEIC, []byte("????ftypheic"),
		[]byte("????ftypheix"),
		[]byte("????ftyphevc"),
		[]byte("????ftypheim"),
		[]byte("????ftypheis"),
		[]byte("????ftyphevm"),
		[]byte("????ftyphevs"),
		[]byte("????ftypmif1"))

	// AVIF magic bytes
	RegisterMagicBytes(AVIF, []byte("????ftypavif"))

	// BMP magic bytes
	RegisterMagicBytes(BMP, []byte("BM"))

	// TIFF magic bytes (little-endian and big-endian)
	RegisterMagicBytes(TIFF, []byte("II*\x00"), []byte("MM\x00*")) // Big-Endian, Little-endian
}
