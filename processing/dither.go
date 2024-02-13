package processing

import (
	"fmt"
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
	// installed via Dockerfile in /opt/pushd-dither
	outFile := fmt.Sprintf("%s-dithered-tmp.png", inFile)
	cmdArgs := []string{"test.py",
		"--pal-meter-13",
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

	cmd := exec.Command("python3", cmdArgs...)
	cmd.Dir = "/opt/pushd-dither"
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("dither failed: %s: %s", err, output)
	}

	// clobber the original file
	return os.Rename(outFile, inFile)
}
