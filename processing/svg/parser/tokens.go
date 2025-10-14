package svgparser

import (
	"encoding/xml"
	"io"
)

type Directive = xml.Directive
type Comment = xml.Comment
type ProcInst = xml.ProcInst

type CData []byte

func readRawCData(r io.ReadSeeker, from, size int64) (CData, error) {
	// Get the current position of the reader so we can return there
	// after reading raw CData.
	curPos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	// Seek to the CData start
	if _, err := r.Seek(from, io.SeekStart); err != nil {
		return nil, err
	}

	// Read the raw CData.
	cdata := make(CData, size)
	if _, err := io.ReadFull(r, cdata); err != nil {
		return nil, err
	}

	// Restore the reader position
	if _, err := r.Seek(curPos, io.SeekStart); err != nil {
		return nil, err
	}

	return cdata, nil
}
