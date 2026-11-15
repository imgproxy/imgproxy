package processing_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/imagedata"
	"github.com/imgproxy/imgproxy/v4/imagemeta/iptc"
	"github.com/imgproxy/imgproxy/v4/imagemeta/photoshop"
	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/processing"
	"github.com/imgproxy/imgproxy/v4/testutil"
	"github.com/imgproxy/imgproxy/v4/vips"
)

func TestStripPS3KeepsEarlierTagsOnParseError(t *testing.T) {
	testData := testutil.NewTestDataProvider(func() *testing.T { return t })
	id := imagedata.NewFromBytesWithFormat(imagetype.BMP, testData.Read("bmp-rle8-overlong-run.bmp"))
	defer id.Close()

	img := new(vips.Image)
	defer img.Clear()
	require.NoError(t, img.Load(id, 1.0, 0, 1))

	// A complete copyright notice followed by a caption whose declared
	// length exceeds the remaining bytes.
	iptcData := "\x1c\x02\x74\x00\x03abc\x1c\x02\x78\x00\x0adef"
	img.SetBlob("iptc-data", photoshop.PhotoshopMap{photoshop.IptcKey: []byte(iptcData)}.Dump())

	ps3Map := make(photoshop.PhotoshopMap)
	photoshop.Parse(processing.StripPS3(img), ps3Map)

	iptcMap := make(iptc.IptcMap)
	require.NoError(t, iptc.Parse(ps3Map[photoshop.IptcKey], iptcMap))
	require.Len(t, iptcMap, 1)
	require.Equal(t, []byte("abc"), iptcMap[iptc.TagKey{RecordID: 2, TagID: 116}][0].Raw)
}
