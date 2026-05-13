package imagetype_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/bufreader"
	"github.com/imgproxy/imgproxy/v4/imagetype"
)

type readerError struct{ error }

func (r readerError) Read(p []byte) (n int, err error) { return 0, r.error }

func TestSVGDetectSuccess(t *testing.T) {
	r := bufreader.New(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`))
	typ, err := imagetype.IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, imagetype.SVG, typ)

	r = bufreader.New(strings.NewReader(`<svg:svg xmlns:svg="http://www.w3.org/2000/svg"></svg:svg>`))
	typ, err = imagetype.IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, imagetype.SVG, typ)

	// Partial content; Simulate limit reader
	r = bufreader.New(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg">SomethingSomething...`))
	typ, err = imagetype.IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, imagetype.SVG, typ)
}

func TestSVGDetectNotSvg(t *testing.T) {
	r := bufreader.New(strings.NewReader(`<html><body>Not an SVG</body></html>`))
	typ, err := imagetype.IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, imagetype.Unknown, typ)

	r = bufreader.New(strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?><not-svg></not-svg>`))
	typ, err = imagetype.IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, imagetype.Unknown, typ)

	r = bufreader.New(strings.NewReader(`<!-- Only comments -->`))
	typ, err = imagetype.IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, imagetype.Unknown, typ)

	// Random byte data that does not match SVG
	r = bufreader.New(bytes.NewReader([]byte{0x42, 0x4D, 0x3C, 0x3F, 0x78, 0x6D, 0x6C, 0x20}))
	typ, err = imagetype.IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, imagetype.Unknown, typ)
}

func TestSVGDetectError(t *testing.T) {
	// Should not return an error for io.EOF
	r := bufreader.New(readerError{error: io.EOF})
	typ, err := imagetype.IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, imagetype.Unknown, typ)

	// Should return an error for other read errors
	r = bufreader.New(readerError{error: io.ErrClosedPipe})
	typ, err = imagetype.IsSVG(r, "", "")
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Equal(t, imagetype.Unknown, typ)
}
