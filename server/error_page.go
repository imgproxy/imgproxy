package server

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/version"
)

var (
	// errorPageTemplateStr is the HTML template for the error page
	//go:embed error_page.html
	errorPageTemplateStr string

	// errorPageTemplate is the parsed HTML template for the error page
	errorPageTemplate = template.Must(
		template.New("error_page").Funcs(template.FuncMap{
			"add":     func(a, b int) int { return a + b },
			"replace": strings.ReplaceAll,
		}).Parse(errorPageTemplateStr),
	)

	// errorTypeReplacer is used to clean up error type strings
	errorTypeReplacer = strings.NewReplacer("*", "", "{", "", "}", "")

	// availableProcessingOptions lists all available image processing options
	// Please note, that this list might not be comprehensive and does not bear such
	// responsibility.
	availableProcessingOptions = []optionCategory{
		{
			Category: "Resizing & Size",
			Options: []optionItem{
				{Name: "resize", Aliases: "rs", Arguments: "type:width:height:enlarge:extend"},
				{Name: "size", Aliases: "s", Arguments: "width:height:enlarge:extend"},
				{Name: "resizing_type", Aliases: "rt", Arguments: "type (fit, fill, fill-down, force, auto)"},
				{Name: "width", Aliases: "w", Arguments: "pixels"},
				{Name: "height", Aliases: "h", Arguments: "pixels"},
				{Name: "min-width", Aliases: "mw", Arguments: "pixels"},
				{Name: "min-height", Aliases: "mh", Arguments: "pixels"},
				{Name: "zoom", Aliases: "z", Arguments: "x_factor[:y_factor]"},
				{Name: "dpr", Aliases: "", Arguments: "factor"},
				{Name: "enlarge", Aliases: "el", Arguments: "0|1"},
				{Name: "extend", Aliases: "ex", Arguments: "0|1[:gravity[:offset_x:offset_y]]"},
				{Name: "extend_aspect_ratio", Aliases: "extend_ar, exar", Arguments: "0|1[:gravity[:offset_x:offset_y]]"},
			},
		},
		{
			Category: "Positioning & Cropping",
			Options: []optionItem{
				{Name: "gravity", Aliases: "g", Arguments: "type[:x:y]"},
				{Name: "crop", Aliases: "c", Arguments: "width[:height[:gravity[:x:y]]]"},
				{Name: "trim", Aliases: "t", Arguments: "threshold[:color[:equal_hor[:equal_ver]]]"},
				{Name: "padding", Aliases: "pd", Arguments: "top:right:bottom:left"},
			},
		},
		{
			Category: "Rotation & Orientation",
			Options: []optionItem{
				{Name: "auto_rotate", Aliases: "ar", Arguments: "0|1"},
				{Name: "rotate", Aliases: "rot", Arguments: "angle"},
				{Name: "flip", Aliases: "fl", Arguments: "0|1 (horizontal)"},
			},
		},
		{
			Category: "Adjustments & Effects",
			Options: []optionItem{
				{Name: "background", Aliases: "bg", Arguments: "r:g:b or hex_color"},
				{Name: "blur", Aliases: "bl", Arguments: "sigma"},
				{Name: "sharpen", Aliases: "sh", Arguments: "sigma"},
				{Name: "pixelate", Aliases: "pix", Arguments: "size"},
			},
		},
		{
			Category: "Watermark & Metadata",
			Options: []optionItem{
				{Name: "watermark", Aliases: "wm", Arguments: "opacity:position[:x:y[:scale]]"},
				{Name: "strip_metadata", Aliases: "sm", Arguments: "0|1"},
				{Name: "keep_copyright", Aliases: "kcr", Arguments: "0|1"},
				{Name: "strip_color_profile", Aliases: "scp", Arguments: "0|1"},
				{Name: "enforce_thumbnail", Aliases: "eth", Arguments: "0|1"},
			},
		},
		{
			Category: "Format & Quality",
			Options: []optionItem{
				{Name: "quality", Aliases: "q", Arguments: "quality"},
				{Name: "format_quality", Aliases: "fq", Arguments: "format:quality"},
				{Name: "max_bytes", Aliases: "mb", Arguments: "bytes"},
				{Name: "format", Aliases: "f, ext", Arguments: "format (jpg, png, webp, avif, gif, ico, svg, heic, bmp, tiff)"},
			},
		},
		{
			Category: "Handling Options",
			Options: []optionItem{
				{Name: "skip_processing", Aliases: "skp", Arguments: "format1[:format2:...]"},
				{Name: "raw", Aliases: "", Arguments: "0|1"},
				{Name: "cachebuster", Aliases: "cb", Arguments: "string"},
				{Name: "expires", Aliases: "exp", Arguments: "timestamp"},
				{Name: "filename", Aliases: "fn", Arguments: "filename"},
				{Name: "return_attachment", Aliases: "att", Arguments: "0|1"},
			},
		},
		{
			Category: "Security & Limits",
			Options: []optionItem{
				{Name: "max_src_resolution", Aliases: "msr", Arguments: "megapixels"},
				{Name: "max_src_file_size", Aliases: "msfs", Arguments: "bytes"},
				{Name: "max_animation_frames", Aliases: "maf", Arguments: "frames"},
				{Name: "max_animation_frame_resolution", Aliases: "mafr", Arguments: "megapixels"},
				{Name: "max_result_dimension", Aliases: "mrd", Arguments: "pixels"},
			},
		},
		{
			Category: "Presets",
			Options: []optionItem{
				{Name: "preset", Aliases: "pr", Arguments: "preset_name1[:preset_name2:...]"},
			},
		},
	}
)

// errorPageData holds the data passed to the error page template.
type errorPageData struct {
	StatusCode       int
	ErrorType        string
	RequestID        string
	PublicMessage    string
	Version          string
	ErrorChain       []errorChainItem
	StackTrace       []stackFrame
	AvailableOptions []optionCategory
}

// optionItem represents a single image processing option.
type optionItem struct {
	Name      string
	Aliases   string
	Arguments string
}

// optionCategory represents a category of image processing options.
type optionCategory struct {
	Category string
	Options  []optionItem
}

// errorChainItem represents a single item in the error chain.
type errorChainItem struct {
	Type       string
	Message    string
	StackFrame *stackFrame // First stack frame where this error occurred
}

// stackFrame represents a single frame in the stack trace.
type stackFrame struct {
	File       string
	Line       int
	Function   string
	ErrorIndex int  // Index in the error chain (0 = outermost, increases to root cause)
	IsLast     bool // True if this frame belongs to the last error in the chain
}

// unwrapErrorChain traverses the error chain using errors.Unwrap and returns
// all error types and messages from the outermost wrapper to the root cause.
func unwrapErrorChain(err error) []errorChainItem {
	var chain []errorChainItem

	for err != nil {
		// Get the error type and extract the first stack frame if available
		errorType := errctx.ErrorType(err)

		var firstFrame *stackFrame
		var errWithCtx errctx.Error

		if errors.As(err, &errWithCtx) {
			stackStr := errWithCtx.FormatStack()

			if stackStr != "" {
				frames := parseStackString(stackStr)

				if len(frames) > 0 {
					firstFrame = &frames[0]
				}
			}
		}

		chain = append(chain, errorChainItem{
			Type:       cleanErrorType(errorType),
			Message:    err.Error(),
			StackFrame: firstFrame,
		})

		err = errors.Unwrap(err)
	}

	return chain
}

// cleanErrorType removes Go syntax artifacts from error type string
func cleanErrorType(t string) string {
	return errorTypeReplacer.Replace(t)
}

// buildStackTrace gets the stack trace from the outermost error and marks
// which frames correspond to which errors in the chain.
func buildStackTrace(err error, chain []errorChainItem) []stackFrame {
	// Get stack trace from the outermost error (it has the full trace)
	var errWithCtx errctx.Error

	if !errors.As(err, &errWithCtx) {
		return nil
	}

	stackStr := errWithCtx.FormatStack()
	if stackStr == "" {
		return nil
	}

	frames := parseStackString(stackStr)

	// Match stack frames with errors in the chain
	for i := range frames {
		frames[i].ErrorIndex = findErrorIndexForFrame(&frames[i], chain)
		frames[i].IsLast = frames[i].ErrorIndex == len(chain)-1
	}

	return frames
}

// findErrorIndexForFrame determines which error in the chain this frame belongs to
func findErrorIndexForFrame(frame *stackFrame, chain []errorChainItem) int {
	// Match the frame with errors in the chain by comparing file and line
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].StackFrame == nil {
			continue
		}

		if chain[i].StackFrame.File == frame.File && chain[i].StackFrame.Line == frame.Line {
			return i
		}
	}

	// If no match found, assign to outermost error
	return 0
}

// parseStackString parses a formatted stack trace string into stack frames
func parseStackString(stackStr string) []stackFrame {
	if stackStr == "" {
		return nil
	}

	lines := bytes.Split([]byte(stackStr), []byte("\n"))
	frames := make([]stackFrame, 0, len(lines))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		// Parse line format: "/path/to/file.go:123 package.FunctionName"
		parts := bytes.SplitN(line, []byte(" "), 2)
		if len(parts) != 2 {
			continue
		}

		// Split file path and line number
		fileLine := bytes.Split(parts[0], []byte(":"))
		if len(fileLine) != 2 {
			continue
		}

		var lineNum int
		fmt.Sscanf(string(fileLine[1]), "%d", &lineNum)

		frames = append(frames, stackFrame{
			File:     string(fileLine[0]),
			Line:     lineNum,
			Function: string(parts[1]),
		})
	}

	return frames
}

// generateErrorHTML renders the error page template with the given error information.
func generateErrorHTML(err errctx.Error, reqID string) (string, error) {
	errorChain := unwrapErrorChain(err)

	data := errorPageData{
		StatusCode:       err.StatusCode(),
		ErrorType:        cleanErrorType(errctx.ErrorType(err)),
		RequestID:        reqID,
		PublicMessage:    err.PublicMessage(),
		Version:          version.Version,
		ErrorChain:       errorChain,
		StackTrace:       buildStackTrace(err, errorChain),
		AvailableOptions: availableProcessingOptions,
	}

	var buf bytes.Buffer
	if err := errorPageTemplate.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
