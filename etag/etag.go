package etag

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"net/textproto"
	"strings"
	"sync"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
)

type eTagCalc struct {
	hash hash.Hash
	enc  *json.Encoder
}

var eTagCalcPool = sync.Pool{
	New: func() interface{} {
		h := sha256.New()

		enc := json.NewEncoder(h)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "")

		return &eTagCalc{h, enc}
	},
}

type Handler struct {
	poHashActual, poHashExpected string

	imgEtagActual, imgEtagExpected string
	imgHashActual, imgHashExpected string
}

func (h *Handler) ParseExpectedETag(etag string) {
	// We suuport only a single ETag value
	if i := strings.IndexByte(etag, ','); i >= 0 {
		etag = textproto.TrimString(etag[:i])
	}

	etagLen := len(etag)

	// ETag is empty or invalid
	if etagLen < 2 {
		return
	}

	// We support strong ETags only
	if etag[0] != '"' || etag[etagLen-1] != '"' {
		return
	}

	// Remove quotes
	etag = etag[1 : etagLen-1]

	i := strings.Index(etag, "/")
	if i < 0 || i > etagLen-3 {
		// Doesn't look like imgproxy ETag
		return
	}

	poPart, imgPartMark, imgPart := etag[:i], etag[i+1], etag[i+2:]

	switch imgPartMark {
	case 'R':
		imgPartDec, err := base64.RawStdEncoding.DecodeString(imgPart)
		if err == nil {
			h.imgEtagExpected = string(imgPartDec)
		}
	case 'D':
		h.imgHashExpected = imgPart
	default:
		// Unknown image part mark
		return
	}

	h.poHashExpected = poPart
}

func (h *Handler) ProcessingOptionsMatch() bool {
	return h.poHashActual == h.poHashExpected
}

func (h *Handler) SetActualProcessingOptions(po *options.ProcessingOptions) bool {
	c := eTagCalcPool.Get().(*eTagCalc)
	defer eTagCalcPool.Put(c)

	c.hash.Reset()
	c.hash.Write([]byte(config.ETagBuster))
	c.enc.Encode(po)

	h.poHashActual = base64.RawURLEncoding.EncodeToString(c.hash.Sum(nil))

	return h.ProcessingOptionsMatch()
}

func (h *Handler) ImageEtagExpected() string {
	return h.imgEtagExpected
}

func (h *Handler) SetActualImageData(imgdata *imagedata.ImageData) bool {
	var haveActualImgETag bool
	h.imgEtagActual, haveActualImgETag = imgdata.Headers["ETag"]
	haveActualImgETag = haveActualImgETag && len(h.imgEtagActual) > 0

	// Just in case server didn't check ETag properly and returned the same one
	// as we expected
	if haveActualImgETag && h.imgEtagExpected == h.imgEtagActual {
		return true
	}

	haveExpectedImgHash := len(h.imgHashExpected) != 0

	if !haveActualImgETag || haveExpectedImgHash {
		c := eTagCalcPool.Get().(*eTagCalc)
		defer eTagCalcPool.Put(c)

		c.hash.Reset()
		c.hash.Write(imgdata.Data)

		h.imgHashActual = base64.RawURLEncoding.EncodeToString(c.hash.Sum(nil))

		return haveExpectedImgHash && h.imgHashActual == h.imgHashExpected
	}

	return false
}

func (h *Handler) GenerateActualETag() string {
	return h.generate(h.poHashActual, h.imgEtagActual, h.imgHashActual)
}

func (h *Handler) GenerateExpectedETag() string {
	return h.generate(h.poHashExpected, h.imgEtagExpected, h.imgHashExpected)
}

func (h *Handler) generate(poHash, imgEtag, imgHash string) string {
	imgPartMark := 'D'
	imgPart := imgHash
	if len(imgEtag) != 0 {
		imgPartMark = 'R'
		imgPart = base64.RawURLEncoding.EncodeToString([]byte(imgEtag))
	}

	return fmt.Sprintf(`"%s/%c%s"`, poHash, imgPartMark, imgPart)
}
