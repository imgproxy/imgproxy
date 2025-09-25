package processing

import (
	"slices"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips/color"
)

// ProcessingOptions is a thin wrapper around options.Options that provides
// helpers for image processing options.
type ProcessingOptions struct {
	*options.Options

	// Config that contains default values for some options.
	config *Config
}

func (p *Processor) NewProcessingOptions(o *options.Options) ProcessingOptions {
	return ProcessingOptions{
		Options: o,
		config:  p.config,
	}
}

func (po ProcessingOptions) Width() int {
	return po.GetInt(keys.Width, 0)
}

func (po ProcessingOptions) Height() int {
	return po.GetInt(keys.Height, 0)
}

func (po ProcessingOptions) MinWidth() int {
	return po.GetInt(keys.MinWidth, 0)
}

func (po ProcessingOptions) MinHeight() int {
	return po.GetInt(keys.MinHeight, 0)
}

func (po ProcessingOptions) ResizingType() ResizeType {
	return options.Get(po.Options, keys.ResizingType, ResizeFit)
}

func (po ProcessingOptions) ZoomWidth() float64 {
	return po.GetFloat(keys.ZoomWidth, 1.0)
}

func (po ProcessingOptions) ZoomHeight() float64 {
	return po.GetFloat(keys.ZoomHeight, 1.0)
}

func (po ProcessingOptions) DPR() float64 {
	return po.GetFloat(keys.Dpr, 1.0)
}

func (po ProcessingOptions) EnforceThumbnail() bool {
	return po.Main().GetBool(keys.EnforceThumbnail, po.config.EnforceThumbnail)
}

func (po ProcessingOptions) Enlarge() bool {
	return po.GetBool(keys.Enlarge, false)
}

func (po ProcessingOptions) Gravity() GravityOptions {
	return NewGravityOptions(po.Options, keys.Gravity, GravityCenter)
}

func (po ProcessingOptions) ExtendEnabled() bool {
	return po.GetBool(keys.ExtendEnabled, false)
}

func (po ProcessingOptions) ExtendGravity() GravityOptions {
	return NewGravityOptions(po.Options, keys.ExtendGravity, GravityCenter)
}

func (po ProcessingOptions) ExtendAspectRatioEnabled() bool {
	return po.GetBool(keys.ExtendAspectRatioEnabled, false)
}

func (po ProcessingOptions) ExtendAspectRatioGravity() GravityOptions {
	return NewGravityOptions(po.Options, keys.ExtendAspectRatioGravity, GravityCenter)
}

func (po ProcessingOptions) Rotate() int {
	return po.GetInt(keys.Rotate, 0)
}

func (po ProcessingOptions) AutoRotate() bool {
	return po.GetBool(keys.AutoRotate, po.config.AutoRotate)
}

func (po ProcessingOptions) CropWidth() float64 {
	return po.GetFloat(keys.CropWidth, 0.0)
}

func (po ProcessingOptions) CropHeight() float64 {
	return po.GetFloat(keys.CropHeight, 0.0)
}

func (po ProcessingOptions) CropGravity() GravityOptions {
	return NewGravityOptions(po.Options, keys.CropGravity, GravityUnknown)
}

func (po ProcessingOptions) Format() imagetype.Type {
	return options.Get(po.Main(), keys.Format, imagetype.Unknown)
}

func (po ProcessingOptions) SetFormat(format imagetype.Type) {
	po.Set(keys.Format, format)
}

func (po ProcessingOptions) ShouldSkipFormatProcessing(inFormat imagetype.Type) bool {
	return slices.Contains(po.config.SkipProcessingFormats, inFormat) ||
		options.SliceContains(po.Main(), keys.SkipProcessing, inFormat)
}

func (po ProcessingOptions) ShouldFlatten() bool {
	return po.Has(keys.Background)
}

func (po ProcessingOptions) Background() color.RGB {
	return options.Get(po.Options, keys.Background, color.RGB{R: 255, G: 255, B: 255})
}

func (po ProcessingOptions) PaddingEnabled() bool {
	return po.PaddingTop() != 0 ||
		po.PaddingRight() != 0 ||
		po.PaddingBottom() != 0 ||
		po.PaddingLeft() != 0
}

func (po ProcessingOptions) PaddingTop() int {
	return po.GetInt(keys.PaddingTop, 0)
}

func (po ProcessingOptions) PaddingRight() int {
	return po.GetInt(keys.PaddingRight, 0)
}

func (po ProcessingOptions) PaddingBottom() int {
	return po.GetInt(keys.PaddingBottom, 0)
}

func (po ProcessingOptions) PaddingLeft() int {
	return po.GetInt(keys.PaddingLeft, 0)
}

func (po ProcessingOptions) Blur() float64 {
	return po.GetFloat(keys.Blur, 0.0)
}

func (po ProcessingOptions) Sharpen() float64 {
	return po.GetFloat(keys.Sharpen, 0.0)
}

func (po ProcessingOptions) Pixelate() int {
	return po.GetInt(keys.Pixelate, 1)
}

func (po ProcessingOptions) PreferWebP() bool {
	return po.GetBool(keys.PreferWebP, false)
}

func (po ProcessingOptions) PreferAvif() bool {
	return po.GetBool(keys.PreferAvif, false)
}

func (po ProcessingOptions) PreferJxl() bool {
	return po.GetBool(keys.PreferJxl, false)
}

func (po ProcessingOptions) EnforceWebP() bool {
	return po.GetBool(keys.EnforceWebP, false)
}

func (po ProcessingOptions) EnforceAvif() bool {
	return po.GetBool(keys.EnforceAvif, false)
}

func (po ProcessingOptions) EnforceJxl() bool {
	return po.GetBool(keys.EnforceJxl, false)
}

func (po ProcessingOptions) TrimEnabled() bool {
	return po.Has(keys.TrimThreshold)
}

func (po ProcessingOptions) DisableTrim() {
	po.Delete(keys.TrimThreshold)
}

func (po ProcessingOptions) TrimThreshold() float64 {
	return po.GetFloat(keys.TrimThreshold, 10.0)
}

func (po ProcessingOptions) TrimSmart() bool {
	return !po.Has(keys.TrimColor)
}

func (po ProcessingOptions) TrimColor() color.RGB {
	return options.Get(po.Options, keys.TrimColor, color.RGB{})
}

func (po ProcessingOptions) TrimEqualHor() bool {
	return po.GetBool(keys.TrimEqualHor, false)
}

func (po ProcessingOptions) TrimEqualVer() bool {
	return po.GetBool(keys.TrimEqualVer, false)
}

func (po ProcessingOptions) WatermarkOpacity() float64 {
	return po.GetFloat(keys.WatermarkOpacity, 0.0)
}

func (po ProcessingOptions) SetWatermarkOpacity(opacity float64) {
	po.Set(keys.WatermarkOpacity, opacity)
}

func (po ProcessingOptions) DeleteWatermarkOpacity() {
	po.Delete(keys.WatermarkOpacity)
}

func (po ProcessingOptions) WatermarkPosition() GravityType {
	return options.Get(po.Options, keys.WatermarkPosition, GravityCenter)
}

func (po ProcessingOptions) WatermarkXOffset() float64 {
	return po.GetFloat(keys.WatermarkXOffset, 0.0)
}

func (po ProcessingOptions) WatermarkYOffset() float64 {
	return po.GetFloat(keys.WatermarkYOffset, 0.0)
}

func (po ProcessingOptions) WatermarkScale() float64 {
	return po.GetFloat(keys.WatermarkScale, 0.0)
}

// Quality retrieves the quality setting for a given image format.
// It first checks for a general quality setting, then for a format-specific setting,
// and finally falls back to the configured default quality.
func (po ProcessingOptions) Quality(format imagetype.Type) int {
	// First, check if quality is explicitly set in options.
	if q := po.Main().GetInt(keys.Quality, 0); q > 0 {
		return q
	}

	// Then, check if format-specific quality is set in options.
	if q := po.Main().GetInt(keys.FormatQuality(format), 0); q > 0 {
		return q
	}

	// Then, check if format-specific quality is set in config.
	if q := po.config.FormatQuality[format]; q > 0 {
		return q
	}

	// Finally, return the general quality setting from config.
	return po.config.Quality
}

func (po ProcessingOptions) MaxBytes() int {
	return po.Main().GetInt(keys.MaxBytes, 0)
}

func (po ProcessingOptions) StripMetadata() bool {
	return po.Main().GetBool(keys.StripMetadata, po.config.StripMetadata)
}

func (po ProcessingOptions) KeepCopyright() bool {
	return po.Main().GetBool(keys.KeepCopyright, po.config.KeepCopyright)
}

func (po ProcessingOptions) StripColorProfile() bool {
	return po.Main().GetBool(keys.StripColorProfile, po.config.StripColorProfile)
}
