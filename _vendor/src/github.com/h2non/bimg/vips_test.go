package bimg

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestVipsRead(t *testing.T) {
	files := []struct {
		name     string
		expected ImageType
	}{
		{"test.jpg", JPEG},
		{"test.png", PNG},
		{"test.webp", WEBP},
	}

	for _, file := range files {
		image, imageType, _ := vipsRead(readImage(file.name))
		if image == nil {
			t.Fatal("Empty image")
		}
		if imageType != file.expected {
			t.Fatal("Invalid image type")
		}
	}
}

func TestVipsSave(t *testing.T) {
	types := [...]ImageType{JPEG, PNG, WEBP}

	for _, typ := range types {
		image, _, _ := vipsRead(readImage("test.jpg"))
		options := vipsSaveOptions{Quality: 95, Type: typ}

		buf, err := vipsSave(image, options)
		if err != nil {
			t.Fatalf("Cannot save the image as '%v'", ImageTypes[typ])
		}
		if len(buf) == 0 {
			t.Fatalf("Empty saved '%v' image", ImageTypes[typ])
		}
	}
}

func TestVipsSaveTiff(t *testing.T) {
	if !IsTypeSupportedSave(TIFF) {
		t.Skipf("Format %#v is not supported", ImageTypes[TIFF])
	}
	image, _, _ := vipsRead(readImage("test.jpg"))
	options := vipsSaveOptions{Quality: 95, Type: TIFF}
	buf, _ := vipsSave(image, options)

	if len(buf) == 0 {
		t.Fatalf("Empty saved '%v' image", ImageTypes[TIFF])
	}
}

func TestVipsRotate(t *testing.T) {
	files := []struct {
		name   string
		rotate Angle
	}{
		{"test.jpg", D90},
		{"test_square.jpg", D45},
	}

	for _, file := range files {
		image, _, _ := vipsRead(readImage(file.name))

		newImg, err := vipsRotate(image, file.rotate)
		if err != nil {
			t.Fatal("Cannot rotate the image")
		}

		buf, _ := vipsSave(newImg, vipsSaveOptions{Quality: 95})
		if len(buf) == 0 {
			t.Fatal("Empty image")
		}
	}
}

func TestVipsZoom(t *testing.T) {
	image, _, _ := vipsRead(readImage("test.jpg"))

	newImg, err := vipsZoom(image, 1)
	if err != nil {
		t.Fatal("Cannot save the image")
	}

	buf, _ := vipsSave(newImg, vipsSaveOptions{Quality: 95})
	if len(buf) == 0 {
		t.Fatal("Empty image")
	}
}

func TestVipsWatermark(t *testing.T) {
	image, _, _ := vipsRead(readImage("test.jpg"))

	watermark := Watermark{
		Text:       "Copy me if you can",
		Font:       "sans bold 12",
		Opacity:    0.5,
		Width:      200,
		DPI:        100,
		Margin:     100,
		Background: Color{255, 255, 255},
	}

	newImg, err := vipsWatermark(image, watermark)
	if err != nil {
		t.Errorf("Cannot add watermark: %s", err)
	}

	buf, _ := vipsSave(newImg, vipsSaveOptions{Quality: 95})
	if len(buf) == 0 {
		t.Fatal("Empty image")
	}
}

func TestVipsWatermarkWithImage(t *testing.T) {
	image, _, _ := vipsRead(readImage("test.jpg"))

	watermark := readImage("transparent.png")

	options := WatermarkImage{Left: 100, Top: 100, Opacity: 1.0, Buf: watermark}
	newImg, err := vipsDrawWatermark(image, options)
	if err != nil {
		t.Errorf("Cannot add watermark: %s", err)
	}

	buf, _ := vipsSave(newImg, vipsSaveOptions{Quality: 95})
	if len(buf) == 0 {
		t.Fatal("Empty image")
	}
}

func TestVipsImageType(t *testing.T) {
	imgType := vipsImageType(readImage("test.jpg"))
	if imgType != JPEG {
		t.Fatal("Invalid image type")
	}
}

func TestVipsMemory(t *testing.T) {
	mem := VipsMemory()

	if mem.Memory < 1024 {
		t.Fatal("Invalid memory")
	}
	if mem.Allocations == 0 {
		t.Fatal("Invalid memory allocations")
	}
}

func readImage(file string) []byte {
	img, _ := os.Open(path.Join("fixtures", file))
	buf, _ := ioutil.ReadAll(img)
	defer img.Close()
	return buf
}
