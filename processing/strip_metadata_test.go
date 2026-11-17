package processing_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/imagedata"
	"github.com/imgproxy/imgproxy/v4/imagemeta/photoshop"
	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/processing"
	"github.com/imgproxy/imgproxy/v4/testutil"
	"github.com/imgproxy/imgproxy/v4/vips"
)

func TestStripPS3DropsPartiallyInvalidIPTC(t *testing.T) {
	testData := testutil.NewTestDataProvider(func() *testing.T { return t })
	id := imagedata.NewFromBytesWithFormat(imagetype.PNG, testData.Read("test-images/png/png.png"))
	defer id.Close()

	img := new(vips.Image)
	defer img.Clear()
	require.NoError(t, img.Load(id, 1.0, 0, 1))

	// A complete copyright notice followed by a caption whose declared
	// length exceeds the remaining bytes.
	iptcData := "\x1c\x02\x74\x00\x03abc\x1c\x02\x78\x00\x0adef"
	img.SetBlob("iptc-data", photoshop.PhotoshopMap{photoshop.IptcKey: []byte(iptcData)}.Dump())

	require.Nil(t, processing.StripPS3(img))
}
