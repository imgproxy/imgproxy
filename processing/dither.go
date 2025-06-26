package processing

import (
	"fmt"
	"github.com/imgproxy/imgproxy/v3/security"
	"math"
	"os"
	"os/exec"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
	log "github.com/sirupsen/logrus"
)

func dither(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	if po.Dither.Type == options.DitherNone {
		return nil
	}

	// Resize image to desired dimensions, retaining aspect, before dithering
	// Usually smaller images are returned to the frame for upscaling, but in the dithering case we want to
	// dither the image only after it's been resized to be bounded within the [po.Width X po.Height] box
	// we trust that anything rendering to the frame itself will have the proper aspect defined

	widthScale := float64(po.Width) / float64(img.Width())
	heightScale := float64(po.Height) / float64(img.Height())
	minScale := math.Min(widthScale, heightScale)
	if err := img.Resize(minScale, minScale); err != nil {
		return err
	}

	// Get a snapshot of current image
	if err := img.CopyMemory(); err != nil {
		return err
	}

	// create empty temp file
	f, err := os.CreateTemp("", "dither*.png")
	if err != nil {
		return err
	}

	// clean up temp file with error logging
	defer func(name string) {
		if err = os.Remove(name); err != nil {
			log.Errorf("failed to remove temp file: %s", err)
		}
	}(f.Name())

	// close immediately, we will replace its contents below
	if err = f.Close(); err != nil {
		return err
	}

	// standard img.Save(imagetype.PNG ...) is lossy and colors are not preserved
	pngData, err := img.SaveHighQualityPNG()
	if err != nil {
		return err
	}
	defer pngData.Close()

	if err = os.WriteFile(f.Name(), pngData.Data, 0644); err != nil {
		return err
	}

	// the dirty business - will clobber the file
	if po.Dither.OptionsSetVendor {
		if err = shellOutVendor(f.Name(), po); err != nil {
			return err
		}
	} else {
		if err = shellOutDither(f.Name(), po); err != nil {
			return err
		}
	}

	// read from dithered file
	ditheredData, err := imagedata.FromFile(f.Name(), "dithered image", security.DefaultOptions())
	if err != nil {
		return err
	}
	defer ditheredData.Close()

	ditheredImg := new(vips.Image)
	if err = ditheredImg.Load(ditheredData, 1, 1.0, 1); err != nil {
		return err
	}
	defer ditheredImg.Clear()

	// replace original image
	// FIXME: use copy? embed image is a bit of a hack on a hack to not have to manage the png data lifecycle
	if err = img.EmbedImage(0, 0, ditheredImg); err != nil {
		return err
	}

	// force lossless output
	po.Format = imagetype.PNG

	// the resulting images are occasionally corrupted if we don't invoke CopyMemory once we're done
	return img.CopyMemory()
}

func shellOutVendor(inFile string, po *options.ProcessingOptions) error {
	// create new temp dir so we don't clobber other running instances
	tmpDir, err := os.MkdirTemp("", "dither-vendor")
	if err != nil {
		return err
	}

	// clean up temp dir when done
	defer func() {
		if err = os.RemoveAll(tmpDir); err != nil {
			log.Errorf("failed to remove vendor temp directory: %s", err)
		}
	}()

	// vendor tool requires file to reside in input directory
	err = os.Mkdir(tmpDir+"/input", 0755)
	if err != nil {
		return fmt.Errorf("failed to create vendor input directory: %s", err)
	}

	// move the infile to the input directory
	tmpInputFile := tmpDir + "/input/vendor_file.png"
	err = os.Rename(inFile, tmpInputFile)
	if err != nil {
		return err
	}

	var cmdArgs []string
	if po.Width < po.Height {
		cmdArgs = append(cmdArgs, "1200x1600")
	} else {
		cmdArgs = append(cmdArgs, "1600x1200")
	}

	// installed via Dockerfile
	cmd := exec.Command("/opt/pushd-dither/vendor/run-vendored.sh", cmdArgs...)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("vendor dither failed: %s: %s", err, output)
	}
	// we don't get an exit code from the vendor tool, so we have to check for the output file

	// move the desired outfile back to the original location

	if po.Dither.SoftProof {
		if err := os.Rename(tmpDir+"/output/vendor_file_proof.png", inFile); err != nil {
			return fmt.Errorf("vendor dither failed to produce proof file, tool output: %s", output)
		}
		return nil
	}

	// FIXME here we're renaming a bmp to a png and just letting the caller load it (without issue but messy)
	if err := os.Rename(tmpDir+"/output/vendor_file_hulk.bmp", inFile); err != nil {
		return fmt.Errorf("vendor dither failed to produce output file, tool output: %s", output)
	}

	return nil
}

func shellOutDither(inFile string, po *options.ProcessingOptions) error {
	outFile := fmt.Sprintf("%s-dithered-tmp.png", inFile)
	proofFile := fmt.Sprintf("%s-dithered-proof-tmp.png", inFile)

	// installed via Dockerfile in /opt/pushd-dither
	cmdArgs := []string{"test.py",
		"--image-in", inFile,
		"--image-out", outFile}

	if po.Dither.Type == options.DitherBNSF {
		cmdArgs = append(cmdArgs, "--shiau-fan")
	}
	if po.Dither.Contrast {
		cmdArgs = append(cmdArgs, "--contrast")
	}
	if po.Dither.Native {
		cmdArgs = append(cmdArgs, "--native")
	}
	if po.Dither.Desaturate {
		cmdArgs = append(cmdArgs, "--desaturate")
	}
	if po.Dither.Meter13 {
		cmdArgs = append(cmdArgs, "--pal-meter-13-hack")
	}
	if po.Dither.SoftProof {
		cmdArgs = append(cmdArgs, "--image-raw", proofFile)
	}
	if po.Dither.Clamp {
		cmdArgs = append(cmdArgs, "--clamp")
	}
	if po.Dither.CLAHESize > 0 {
		cmdArgs = append(cmdArgs, "--clahe-size", fmt.Sprintf("%d", po.Dither.CLAHESize))
	}
	if po.Dither.SaturationScale > 0 {
		cmdArgs = append(cmdArgs, "--saturation-scale", fmt.Sprintf("%f", po.Dither.SaturationScale))
	}
	if po.Dither.HullProject {
		cmdArgs = append(cmdArgs, "--hull-project")
	}
	if po.Dither.AutoEnhance {
		cmdArgs = append(cmdArgs, "--auto-enhance")
	}
	if len(po.Dither.LUTFile) > 0 {
		cmdArgs = append(cmdArgs, "--lut", fmt.Sprintf("lut_dither/%s.npy", po.Dither.LUTFile))
		// specifying the precomputed hue-sat file is a speed optimization
		cmdArgs = append(cmdArgs, "--lut-hue-sat", fmt.Sprintf("lut_dither/%s.hue_sat", po.Dither.LUTFile))
	}
	if po.Dither.LUTBlue {
		cmdArgs = append(cmdArgs, "--lut-blue")
	}
	if po.Dither.NormalizeContrast {
		cmdArgs = append(cmdArgs, "--normalize-contrast")
	}

	switch {
	case po.Dither.OptionsSet01:
		cmdArgs = append(cmdArgs, "--cam16")
		cmdArgs = append(cmdArgs, "--chroma-lightness")
		cmdArgs = append(cmdArgs, "--saturation-scale", "1.0")
		cmdArgs = append(cmdArgs, "--hull-project")
		cmdArgs = append(cmdArgs, "--pal-meter-13-hack")
		cmdArgs = append(cmdArgs, "--contrast")
		cmdArgs = append(cmdArgs, "--shrink-gamut", "1.5")
	case po.Dither.OptionsSet02:
		cmdArgs = append(cmdArgs, "--cam16")
		cmdArgs = append(cmdArgs, "--chroma-lightness")
		cmdArgs = append(cmdArgs, "--saturation-scale", "1.0")
		cmdArgs = append(cmdArgs, "--hull-project")
		cmdArgs = append(cmdArgs, "--pal-meter-13-hack")
		cmdArgs = append(cmdArgs, "--contrast")
		cmdArgs = append(cmdArgs, "--shrink-gamut", "1.5")
		cmdArgs = append(cmdArgs, "--clip-error")
	case po.Dither.OptionsSet03:
		cmdArgs = append(cmdArgs, "--jzazbz")
		cmdArgs = append(cmdArgs, "--map-palette", "pal_inflate_extra")
		cmdArgs = append(cmdArgs, "--pal-inflate")
		cmdArgs = append(cmdArgs, "--chroma-lightness")
		cmdArgs = append(cmdArgs, "--shrink-gamut", "1.1")
		cmdArgs = append(cmdArgs, "--saturation-scale", "1.0")
		cmdArgs = append(cmdArgs, "--clip-error")
	case po.Dither.OptionsSet04:
		cmdArgs = append(cmdArgs, "--jzazbz")
		cmdArgs = append(cmdArgs, "--map-palette", "pal_inflate_extra")
		cmdArgs = append(cmdArgs, "--pal-inflate")
		cmdArgs = append(cmdArgs, "--chroma-lightness")
		cmdArgs = append(cmdArgs, "--shrink-gamut", "1.1")
		cmdArgs = append(cmdArgs, "--saturation-scale", "1.0")
		cmdArgs = append(cmdArgs, "--clip-error")
		cmdArgs = append(cmdArgs, "--auto-enhance")
	case po.Dither.OptionsSet05:
		cmdArgs = append(cmdArgs, "--jzazbz")
		cmdArgs = append(cmdArgs, "--hull-project")
		cmdArgs = append(cmdArgs, "--chroma-lightness")
		cmdArgs = append(cmdArgs, "--shrink-gamut", "1.1")
		cmdArgs = append(cmdArgs, "--saturation-scale", "1.0")
		cmdArgs = append(cmdArgs, "--clip-error")
		cmdArgs = append(cmdArgs, "--auto-enhance")
		cmdArgs = append(cmdArgs, "--dea-weight", "0.95")
		cmdArgs = append(cmdArgs, "--pal-auto-expand", "2.0")
		cmdArgs = append(cmdArgs, "--inflate-color-space", "jzazbz")
		cmdArgs = append(cmdArgs, "--pal-str", po.Dither.MeasuredPalette)
	case po.Dither.OptionsSetCam16:
		cmdArgs = append(cmdArgs, "--cam16")
		cmdArgs = append(cmdArgs, "--chroma-lightness")
		cmdArgs = append(cmdArgs, "--saturation-scale", "1.0")
		cmdArgs = append(cmdArgs, "--hull-project")
		cmdArgs = append(cmdArgs, "--pal-meter-13")
		cmdArgs = append(cmdArgs, "--contrast")
		cmdArgs = append(cmdArgs, "--shrink-gamut", "1.5")
		cmdArgs = append(cmdArgs, "--clip-error")
	case po.Dither.OptionsSetHpminde:
		cmdArgs = append(cmdArgs, "--contrast")
		cmdArgs = append(cmdArgs, "--hull-project")
		cmdArgs = append(cmdArgs, "--lut-blue")
		cmdArgs = append(cmdArgs, "--lut", "lut_dither/hpminde_rgb.npy")
		cmdArgs = append(cmdArgs, "--lut-hue-sat", "lut_dither/hpminde_rgb.hue_sat")
		cmdArgs = append(cmdArgs, "--saturation-scale", "1.0")
		cmdArgs = append(cmdArgs, "--pal-meter-13")
	case po.Dither.OptionsSetScam:
		cmdArgs = append(cmdArgs, "--scam")
		cmdArgs = append(cmdArgs, "--chroma-lightness")
		cmdArgs = append(cmdArgs, "--saturation-scale", "1.0")
		cmdArgs = append(cmdArgs, "--hull-project")
		cmdArgs = append(cmdArgs, "--pal-meter-13")
		cmdArgs = append(cmdArgs, "--contrast")
		cmdArgs = append(cmdArgs, "--shrink-gamut", "1.5")
		cmdArgs = append(cmdArgs, "--clip-error")
		cmdArgs = append(cmdArgs, "--project-3d")
	}

	cmd := exec.Command("python3", cmdArgs...)
	cmd.Dir = "/opt/pushd-dither"
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dither failed: %s: %s", err, output)
	}

	if po.Dither.SoftProof {
		// cleanup unused outfile, replace with proof file
		err := os.Remove(outFile)
		if err != nil {
			log.Errorf("failed to remove unused out file: %s", err)
		}
		outFile = proofFile
	}
	// clobber the original file
	return os.Rename(outFile, inFile)
}
