package vips_test

import (
	"testing"

	"github.com/imgproxy/imgproxy/v4/imagedata"
	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/testutil"
	"github.com/imgproxy/imgproxy/v4/vips"
	"github.com/stretchr/testify/require"
)

func TestIcoLoadInvalid(t *testing.T) {
	testData := testutil.NewTestDataProvider(func() *testing.T { return t })

	testCases := []struct {
		filename string
		errStr   string
	}{
		{"ico-bad1.ico", "ICO image data is too large"},
		{"ico-bad2.ico", "ICO image data is too small"},
		{"ico-bad3.ico", "ICO image data is too large"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			data := testData.Read(tc.filename)
			id := imagedata.NewFromBytesWithFormat(imagetype.ICO, data)
			defer id.Close()

			img := new(vips.Image)
			defer img.Clear()

			err := img.Load(id, 1.0, 0, 1)
			require.ErrorContains(t, err, tc.errStr)
		})
	}
}
