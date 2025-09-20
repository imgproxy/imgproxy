package keys

const (
	Width  = "width"
	Height = "height"

	MinWidth  = "min-width"
	MinHeight = "min-height"

	Enlarge = "enlarge"

	ExtendEnabled        = PrefixExtend + SuffixEnabled
	ExtendGravity        = PrefixExtend + SuffixGravity
	ExtendGravityType    = ExtendGravity + SuffixType
	ExtendGravityXOffset = ExtendGravity + SuffixXOffset
	ExtendGravityYOffset = ExtendGravity + SuffixYOffset

	ExtendAspectRatioEnabled        = PrefixExtendAspectRatio + SuffixEnabled
	ExtendAspectRatioGravity        = PrefixExtendAspectRatio + SuffixGravity
	ExtendAspectRatioGravityType    = ExtendAspectRatioGravity + SuffixType
	ExtendAspectRatioGravityXOffset = ExtendAspectRatioGravity + SuffixXOffset
	ExtendAspectRatioGravityYOffset = ExtendAspectRatioGravity + SuffixYOffset

	ResizingType = "resizing_type"

	ZoomWidth  = "zoom_width"
	ZoomHeight = "zoom_height"

	Dpr = "dpr"

	Gravity        = "gravity"
	GravityType    = Gravity + SuffixType
	GravityXOffset = Gravity + SuffixXOffset
	GravityYOffset = Gravity + SuffixYOffset

	CropWidth          = "crop.width"
	CropHeight         = "crop.height"
	CropGravity        = "crop" + SuffixGravity
	CropGravityType    = CropGravity + SuffixType
	CropGravityXOffset = CropGravity + SuffixXOffset
	CropGravityYOffset = CropGravity + SuffixYOffset

	PaddingEnabled = "padding._enabled"
	PaddingTop     = "padding.top"
	PaddingRight   = "padding.right"
	PaddingBottom  = "padding.bottom"
	PaddingLeft    = "padding.left"

	TrimEnabled   = "trim._enabled"
	TrimThreshold = "trim.threshold"
	TrimSmart     = "trim._smart"
	TrimColor     = "trim.color"
	TrimEqualHor  = "trim.equal_horizontal"
	TrimEqualVer  = "trim.equal_vertical"

	Rotate = "rotate"

	Quality       = "quality"
	FormatQuality = "format_quality"

	MaxBytes = "max_bytes"

	Flatten    = "_flatten"
	Background = "background"

	Blur     = "blur"
	Sharpen  = "sharpen"
	Pixelate = "pixelate"

	WatermarkEnabled  = "watermark._enabled"
	WatermarkOpacity  = "watermark.opacity"
	WatermarkPosition = "watermark.position"
	WatermarkXOffset  = "watermark" + SuffixXOffset
	WatermarkYOffset  = "watermark" + SuffixYOffset
	WatermarkScale    = "watermark.scale"

	Format = "format"

	CacheBuster = "cachebuster"

	SkipProcessing = "skip_processing"

	Raw = "raw"

	Filename = "filename"

	Expires = "expires"

	StripMetadata     = "strip_metadata"
	KeepCopyright     = "keep_copyright"
	StripColorProfile = "strip_color_profile"

	AutoRotate = "auto_rotate"

	EnforceThumbnail = "enforce_thumbnail"

	ReturnAttachment = "return_attachment"

	MaxSrcResolution            = "max_src_resolution"
	MaxSrcFileSize              = "max_src_file_size"
	MaxAnimationFrames          = "max_animation_frames"
	MaxAnimationFrameResolution = "max_animation_frame_resolution"
	MaxResultDimension          = "max_result_dimension"

	PreferWebP  = "prefer_webp"
	EnforceWebP = "enforce_webp"
	PreferAvif  = "prefer_avif"
	EnforceAvif = "enforce_avif"
	PreferJxl   = "prefer_jxl"
	EnforceJxl  = "enforce_jxl"

	UsedPresets = "used_presets"

	PrefixExtend            = "extend"
	PrefixExtendAspectRatio = "extend_aspect_ratio"

	SuffixEnabled = ".enabled"
	SuffixGravity = ".gravity"
	SuffixType    = ".type"
	SuffixXOffset = ".x_offset"
	SuffixYOffset = ".y_offset"
)
