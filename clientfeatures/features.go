package clientfeatures

// Features holds information about features extracted from HTTP request
type Features struct {
	PreferWebP  bool // Whether to prefer WebP format when resulting image format is unknown
	EnforceWebP bool // Whether to enforce WebP format even if resulting image format is set

	PreferAvif  bool // Whether to prefer AVIF format when resulting image format is unknown
	EnforceAvif bool // Whether to enforce AVIF format even if resulting image format is set

	PreferJxl  bool // Whether to prefer JXL format when resulting image format is unknown
	EnforceJxl bool // Whether to enforce JXL format even if resulting image format is set

	ClientHintsWidth int     // Client hint width
	ClientHintsDPR   float64 // Client hint device pixel ratio
}
