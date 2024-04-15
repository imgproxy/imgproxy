package processing

import (
	"fmt"
	"math"
	"os"
	"os/exec"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/security"
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

	pngData, err := img.Save(imagetype.PNG, 0)
	if err != nil {
		return err
	}
	defer pngData.Close()

	if err = os.WriteFile(f.Name(), pngData.Data, 0644); err != nil {
		return err
	}

	// the dirty business - will clobber the file
	if err = shellOutDither(f.Name(), po); err != nil {
		return err
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
	if po.Dither.LUTBlue {
		cmdArgs = append(cmdArgs, "--lut-blue")
	}
	if len(po.Dither.LUTFile) > 0 {
		cmdArgs = append(cmdArgs, "--lut-file", fmt.Sprintf("lut_dither/%s", po.Dither.LUTFile))
	}
	if po.Dither.NormalizeContrast {
		cmdArgs = append(cmdArgs, "--normalize-contrast")
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
