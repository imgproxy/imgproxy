// Inspired by https://github.com/mattrobenolt/go-httpheaders
// Thanks, Matt Robenolt!
package httpheaders

import (
	"fmt"
	"mime"
	"path/filepath"
	"strings"
)

const (
	Accept                          = "Accept"
	AcceptCharset                   = "Accept-Charset"
	AcceptEncoding                  = "Accept-Encoding"
	AcceptLanguage                  = "Accept-Language"
	AcceptRanges                    = "Accept-Ranges"
	AccessControlAllowCredentials   = "Access-Control-Allow-Credentials"
	AccessControlAllowHeaders       = "Access-Control-Allow-Headers"
	AccessControlAllowMethods       = "Access-Control-Allow-Methods"
	AccessControlAllowOrigin        = "Access-Control-Allow-Origin"
	AccessControlMaxAge             = "Access-Control-Max-Age"
	Age                             = "Age"
	AltSvc                          = "Alt-Svc"
	Authorization                   = "Authorization"
	CacheControl                    = "Cache-Control"
	Connection                      = "Connection"
	ContentDisposition              = "Content-Disposition"
	ContentEncoding                 = "Content-Encoding"
	ContentLanguage                 = "Content-Language"
	ContentLength                   = "Content-Length"
	ContentRange                    = "Content-Range"
	ContentSecurityPolicy           = "Content-Security-Policy"
	ContentSecurityPolicyReportOnly = "Content-Security-Policy-Report-Only"
	ContentType                     = "Content-Type"
	Cookie                          = "Cookie"
	Date                            = "Date"
	Dnt                             = "Dnt"
	Etag                            = "Etag"
	Expect                          = "Expect"
	ExpectCt                        = "Expect-Ct"
	Expires                         = "Expires"
	Forwarded                       = "Forwarded"
	Host                            = "Host"
	IfMatch                         = "If-Match"
	IfModifiedSince                 = "If-Modified-Since"
	IfNoneMatch                     = "If-None-Match"
	IfUnmodifiedSince               = "If-Unmodified-Since"
	KeepAlive                       = "Keep-Alive"
	LastModified                    = "Last-Modified"
	Link                            = "Link"
	Location                        = "Location"
	Origin                          = "Origin"
	Pragma                          = "Pragma"
	Referer                         = "Referer"
	RequestId                       = "Request-Id"
	RetryAfter                      = "Retry-After"
	Server                          = "Server"
	SetCookie                       = "Set-Cookie"
	StrictTransportSecurity         = "Strict-Transport-Security"
	Upgrade                         = "Upgrade"
	UserAgent                       = "User-Agent"
	Vary                            = "Vary"
	Via                             = "Via"
	WwwAuthenticate                 = "Www-Authenticate"
	XContentTypeOptions             = "X-Content-Type-Options"
	XForwardedFor                   = "X-Forwarded-For"
	XForwardedHost                  = "X-Forwarded-Host"
	XForwardedProto                 = "X-Forwarded-Proto"
	XFrameOptions                   = "X-Frame-Options"
	XOriginWidth                    = "X-Origin-Width"
	XOriginHeight                   = "X-Origin-Height"
	XResultWidth                    = "X-Result-Width"
	XResultHeight                   = "X-Result-Height"
	XOriginContentLength            = "X-Origin-Content-Length"
)

const (
	// fallbackStem is used when the stem cannot be determined from the URL.
	fallbackStem = "image"

	// Content-Disposition header format
	contentDispositionsHeader = "%s; filename=\"%s%s\""

	// "inline" disposition types
	inlineDisposition = "inline"

	// "attachment" disposition type
	attachmentDisposition = "attachment"
)

// Value generates the content-disposition header value.
//
// It uses the following priorities:
// 1. By default, it uses the filename and extension from the URL.
// 2. If `filename` is provided, it overrides the URL filename.
// 3. If `contentType` is provided, it tries to determine the extension from the content type.
// 4. If `ext` is provided, it overrides any extension determined from the URL or header.
// 5. If `fallback` is true and the filename is still empty, it uses fallback stem.
func ContentDispositionValue(url, filename, ext, contentType string, returnAttachment, fallback bool) string {
	// By default, let's use the URL filename and extension
	_, urlFilename := filepath.Split(url)
	urlExt := filepath.Ext(urlFilename)

	var rStem string

	// Avoid strings.TrimSuffix allocation by using slice operation
	if urlExt != "" {
		rStem = urlFilename[:len(urlFilename)-len(urlExt)]
	} else {
		rStem = urlFilename
	}

	var rExt = urlExt

	// If filename is provided explicitly, use it
	if len(filename) > 0 {
		rStem = filename
	}

	// If ext is provided explicitly, use it
	if len(ext) > 0 {
		rExt = ext
	} else if len(contentType) > 0 {
		exts, err := mime.ExtensionsByType(contentType)
		if err == nil && len(exts) != 0 {
			rExt = exts[0]
		}
	}

	// If fallback is requested, and filename is still empty, override it with fallbackStem
	if fallback && len(rStem) == 0 {
		rStem = fallbackStem
	}

	disposition := inlineDisposition

	// Create the content-disposition header value
	if returnAttachment {
		disposition = attachmentDisposition
	}

	return fmt.Sprintf(contentDispositionsHeader, disposition, strings.ReplaceAll(rStem, `"`, "%22"), rExt)
}
