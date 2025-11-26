package imagetype

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v3/bufreader"
)

type errReader struct{ error }

func (r errReader) Read(p []byte) (n int, err error) { return 0, r.error }

func TestSVGDetectSuccess(t *testing.T) {
	r := bufreader.New(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`))
	typ, err := IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, SVG, typ)

	r = bufreader.New(strings.NewReader(`<svg:svg xmlns:svg="http://www.w3.org/2000/svg"></svg:svg>`))
	typ, err = IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, SVG, typ)

	// Partial content; Simulate limit reader
	r = bufreader.New(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg">SomethingSomething...`))
	typ, err = IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, SVG, typ)
}

func TestSVGDetectNotSvg(t *testing.T) {
	r := bufreader.New(strings.NewReader(`<html><body>Not an SVG</body></html>`))
	typ, err := IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, Unknown, typ)

	r = bufreader.New(strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?><not-svg></not-svg>`))
	typ, err = IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, Unknown, typ)

	r = bufreader.New(strings.NewReader(`<!-- Only comments -->`))
	typ, err = IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, Unknown, typ)

	// Random byte data that does not match SVG
	r = bufreader.New(bytes.NewReader([]byte{0x42, 0x4D, 0x3C, 0x3F, 0x78, 0x6D, 0x6C, 0x20}))
	typ, err = IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, Unknown, typ)
}

func TestSVGDetectError(t *testing.T) {
	// Should not return an error for io.EOF
	r := bufreader.New(errReader{error: io.EOF})
	typ, err := IsSVG(r, "", "")
	require.NoError(t, err)
	require.Equal(t, Unknown, typ)

	// Should return an error for other read errors
	r = bufreader.New(errReader{error: io.ErrClosedPipe})
	typ, err = IsSVG(r, "", "")
	require.ErrorIs(t, err, io.ErrClosedPipe)
	require.Equal(t, Unknown, typ)
}
