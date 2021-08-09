package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func intEnvConfig(i *int, name string) {
	if env, err := strconv.Atoi(os.Getenv(name)); err == nil {
		*i = env
	}
}

func floatEnvConfig(i *float64, name string) {
	if env, err := strconv.ParseFloat(os.Getenv(name), 64); err == nil {
		*i = env
	}
}

func megaIntEnvConfig(f *int, name string) {
	if env, err := strconv.ParseFloat(os.Getenv(name), 64); err == nil {
		*f = int(env * 1000000)
	}
}

func strEnvConfig(s *string, name string) {
	if env := os.Getenv(name); len(env) > 0 {
		*s = env
	}
}

func boolEnvConfig(b *bool, name string) {
	if env, err := strconv.ParseBool(os.Getenv(name)); err == nil {
		*b = env
	}
}

func imageTypesEnvConfig(it *[]imageType, name string) {
	*it = []imageType{}

	if env := os.Getenv(name); len(env) > 0 {
		parts := strings.Split(env, ",")

		for _, p := range parts {
			pt := strings.TrimSpace(p)
			if t, ok := imageTypes[pt]; ok {
				*it = append(*it, t)
			} else {
				logWarning("Unknown image format to skip: %s", pt)
			}
		}
	}
}

func formatQualityEnvConfig(m map[imageType]int, name string) {
	if env := os.Getenv(name); len(env) > 0 {
		parts := strings.Split(env, ",")

		for _, p := range parts {
			i := strings.Index(p, "=")
			if i < 0 {
				logWarning("Invalid format quality string: %s", p)
				continue
			}

			imgtypeStr, qStr := strings.TrimSpace(p[:i]), strings.TrimSpace(p[i+1:])

			imgtype, ok := imageTypes[imgtypeStr]
			if !ok {
				logWarning("Invalid format: %s", p)
			}

			q, err := strconv.Atoi(qStr)
			if err != nil || q <= 0 || q > 100 {
				logWarning("Invalid quality: %s", p)
			}

			m[imgtype] = q
		}
	}
}

func hexEnvConfig(b *[]securityKey, name string) error {
	var err error

	if env := os.Getenv(name); len(env) > 0 {
		parts := strings.Split(env, ",")

		keys := make([]securityKey, len(parts))

		for i, part := range parts {
			if keys[i], err = hex.DecodeString(part); err != nil {
				return fmt.Errorf("%s expected to be hex-encoded strings. Invalid: %s\n", name, part)
			}
		}

		*b = keys
	}

	return nil
}

func hexFileConfig(b *[]securityKey, filepath string) error {
	if len(filepath) == 0 {
		return nil
	}

	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("Can't open file %s\n", filepath)
	}

	keys := []securityKey{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		part := scanner.Text()

		if len(part) == 0 {
			continue
		}

		if key, err := hex.DecodeString(part); err == nil {
			keys = append(keys, key)
		} else {
			return fmt.Errorf("%s expected to contain hex-encoded strings. Invalid: %s\n", filepath, part)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Failed to read file %s: %s", filepath, err)
	}

	*b = keys

	return nil
}

func presetEnvConfig(p presets, name string) error {
	if env := os.Getenv(name); len(env) > 0 {
		presetStrings := strings.Split(env, ",")

		for _, presetStr := range presetStrings {
			if err := parsePreset(p, presetStr); err != nil {
				return fmt.Errorf(err.Error())
			}
		}
	}

	return nil
}

func presetFileConfig(p presets, filepath string) error {
	if len(filepath) == 0 {
		return nil
	}

	f, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("Can't open file %s\n", filepath)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := parsePreset(p, scanner.Text()); err != nil {
			return fmt.Errorf(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Failed to read presets file: %s", err)
	}

	return nil
}

func patternsEnvConfig(s *[]*regexp.Regexp, name string) {
	if env := os.Getenv(name); len(env) > 0 {
		parts := strings.Split(env, ",")
		result := make([]*regexp.Regexp, len(parts))

		for i, p := range parts {
			result[i] = regexpFromPattern(strings.TrimSpace(p))
		}

		*s = result
	} else {
		*s = []*regexp.Regexp{}
	}
}

func regexpFromPattern(pattern string) *regexp.Regexp {
	var result strings.Builder
	// Perform prefix matching
	result.WriteString("^")
	for i, part := range strings.Split(pattern, "*") {
		// Add a regexp match all without slashes for each wildcard character
		if i > 0 {
			result.WriteString("[^/]*")
		}

		// Quote other parts of the pattern
		result.WriteString(regexp.QuoteMeta(part))
	}
	// It is safe to use regexp.MustCompile since the expression is always valid
	return regexp.MustCompile(result.String())
}

type config struct {
	Network          string
	Bind             string
	ReadTimeout      int
	WriteTimeout     int
	KeepAliveTimeout int
	DownloadTimeout  int
	Concurrency      int
	MaxClients       int

	TTL                     int
	CacheControlPassthrough bool
	SetCanonicalHeader      bool

	SoReuseport bool

	PathPrefix string

	MaxSrcDimension    int
	MaxSrcResolution   int
	MaxSrcFileSize     int
	MaxAnimationFrames int
	MaxSvgCheckBytes   int

	JpegProgressive       bool
	PngInterlaced         bool
	PngQuantize           bool
	PngQuantizationColors int
	AvifSpeed             int
	Quality               int
	FormatQuality         map[imageType]int
	GZipCompression       int
	StripMetadata         bool
	StripColorProfile     bool
	AutoRotate            bool

	EnableWebpDetection bool
	EnforceWebp         bool
	EnableAvifDetection bool
	EnforceAvif         bool
	EnableClientHints   bool

	SkipProcessingFormats []imageType

	UseLinearColorspace bool
	DisableShrinkOnLoad bool

	Keys          []securityKey
	Salts         []securityKey
	AllowInsecure bool
	SignatureSize int

	Secret string

	AllowOrigin string

	UserAgent string

	IgnoreSslVerification bool
	DevelopmentErrorsMode bool

	AllowedSources      []*regexp.Regexp
	LocalFileSystemRoot string
	S3Enabled           bool
	S3Region            string
	S3Endpoint          string
	GCSEnabled          bool
	GCSKey              string
	ABSEnabled          bool
	ABSName             string
	ABSKey              string
	ABSEndpoint         string

	ETagEnabled bool

	BaseURL string

	Presets     presets
	OnlyPresets bool

	WatermarkData    string
	WatermarkPath    string
	WatermarkURL     string
	WatermarkOpacity float64

	FallbackImageData string
	FallbackImagePath string
	FallbackImageURL  string

	NewRelicAppName string
	NewRelicKey     string

	PrometheusBind      string
	PrometheusNamespace string

	BugsnagKey        string
	BugsnagStage      string
	HoneybadgerKey    string
	HoneybadgerEnv    string
	SentryDSN         string
	SentryEnvironment string
	SentryRelease     string
	AirbrakeProjecID  int
	AirbrakeProjecKey string
	AirbrakeEnv       string

	ReportDownloadingErrors bool

	EnableDebugHeaders bool

	FreeMemoryInterval             int
	DownloadBufferSize             int
	GZipBufferSize                 int
	BufferPoolCalibrationThreshold int
}

var conf = config{
	Network:                        "tcp",
	Bind:                           ":8080",
	ReadTimeout:                    10,
	WriteTimeout:                   10,
	KeepAliveTimeout:               10,
	DownloadTimeout:                5,
	Concurrency:                    runtime.NumCPU() * 2,
	TTL:                            3600,
	MaxSrcResolution:               16800000,
	MaxAnimationFrames:             1,
	MaxSvgCheckBytes:               32 * 1024,
	SignatureSize:                  32,
	PngQuantizationColors:          256,
	Quality:                        80,
	AvifSpeed:                      5,
	FormatQuality:                  map[imageType]int{imageTypeAVIF: 50},
	StripMetadata:                  true,
	StripColorProfile:              true,
	AutoRotate:                     true,
	UserAgent:                      fmt.Sprintf("imgproxy/%s", version),
	Presets:                        make(presets),
	WatermarkOpacity:               1,
	BugsnagStage:                   "production",
	HoneybadgerEnv:                 "production",
	SentryEnvironment:              "production",
	SentryRelease:                  fmt.Sprintf("imgproxy/%s", version),
	AirbrakeEnv:                    "production",
	ReportDownloadingErrors:        true,
	FreeMemoryInterval:             10,
	BufferPoolCalibrationThreshold: 1024,
}

func configure() error {
	keyPath := flag.String("keypath", "", "path of the file with hex-encoded key")
	saltPath := flag.String("saltpath", "", "path of the file with hex-encoded salt")
	presetsPath := flag.String("presets", "", "path of the file with presets")
	flag.Parse()

	if port := os.Getenv("PORT"); len(port) > 0 {
		conf.Bind = fmt.Sprintf(":%s", port)
	}

	strEnvConfig(&conf.Network, "IMGPROXY_NETWORK")
	strEnvConfig(&conf.Bind, "IMGPROXY_BIND")
	intEnvConfig(&conf.ReadTimeout, "IMGPROXY_READ_TIMEOUT")
	intEnvConfig(&conf.WriteTimeout, "IMGPROXY_WRITE_TIMEOUT")
	intEnvConfig(&conf.KeepAliveTimeout, "IMGPROXY_KEEP_ALIVE_TIMEOUT")
	intEnvConfig(&conf.DownloadTimeout, "IMGPROXY_DOWNLOAD_TIMEOUT")
	intEnvConfig(&conf.Concurrency, "IMGPROXY_CONCURRENCY")
	intEnvConfig(&conf.MaxClients, "IMGPROXY_MAX_CLIENTS")

	intEnvConfig(&conf.TTL, "IMGPROXY_TTL")
	boolEnvConfig(&conf.CacheControlPassthrough, "IMGPROXY_CACHE_CONTROL_PASSTHROUGH")
	boolEnvConfig(&conf.SetCanonicalHeader, "IMGPROXY_SET_CANONICAL_HEADER")

	boolEnvConfig(&conf.SoReuseport, "IMGPROXY_SO_REUSEPORT")

	strEnvConfig(&conf.PathPrefix, "IMGPROXY_PATH_PREFIX")

	intEnvConfig(&conf.MaxSrcDimension, "IMGPROXY_MAX_SRC_DIMENSION")
	megaIntEnvConfig(&conf.MaxSrcResolution, "IMGPROXY_MAX_SRC_RESOLUTION")
	intEnvConfig(&conf.MaxSrcFileSize, "IMGPROXY_MAX_SRC_FILE_SIZE")
	intEnvConfig(&conf.MaxSvgCheckBytes, "IMGPROXY_MAX_SVG_CHECK_BYTES")

	if _, ok := os.LookupEnv("IMGPROXY_MAX_GIF_FRAMES"); ok {
		logWarning("`IMGPROXY_MAX_GIF_FRAMES` is deprecated and will be removed in future versions. Use `IMGPROXY_MAX_ANIMATION_FRAMES` instead")
		intEnvConfig(&conf.MaxAnimationFrames, "IMGPROXY_MAX_GIF_FRAMES")
	}
	intEnvConfig(&conf.MaxAnimationFrames, "IMGPROXY_MAX_ANIMATION_FRAMES")

	patternsEnvConfig(&conf.AllowedSources, "IMGPROXY_ALLOWED_SOURCES")

	intEnvConfig(&conf.AvifSpeed, "IMGPROXY_AVIF_SPEED")
	boolEnvConfig(&conf.JpegProgressive, "IMGPROXY_JPEG_PROGRESSIVE")
	boolEnvConfig(&conf.PngInterlaced, "IMGPROXY_PNG_INTERLACED")
	boolEnvConfig(&conf.PngQuantize, "IMGPROXY_PNG_QUANTIZE")
	intEnvConfig(&conf.PngQuantizationColors, "IMGPROXY_PNG_QUANTIZATION_COLORS")
	intEnvConfig(&conf.Quality, "IMGPROXY_QUALITY")
	formatQualityEnvConfig(conf.FormatQuality, "IMGPROXY_FORMAT_QUALITY")
	intEnvConfig(&conf.GZipCompression, "IMGPROXY_GZIP_COMPRESSION")
	boolEnvConfig(&conf.StripMetadata, "IMGPROXY_STRIP_METADATA")
	boolEnvConfig(&conf.StripColorProfile, "IMGPROXY_STRIP_COLOR_PROFILE")
	boolEnvConfig(&conf.AutoRotate, "IMGPROXY_AUTO_ROTATE")

	boolEnvConfig(&conf.EnableWebpDetection, "IMGPROXY_ENABLE_WEBP_DETECTION")
	boolEnvConfig(&conf.EnforceWebp, "IMGPROXY_ENFORCE_WEBP")
	boolEnvConfig(&conf.EnableAvifDetection, "IMGPROXY_ENABLE_AVIF_DETECTION")
	boolEnvConfig(&conf.EnforceAvif, "IMGPROXY_ENFORCE_AVIF")
	boolEnvConfig(&conf.EnableClientHints, "IMGPROXY_ENABLE_CLIENT_HINTS")

	imageTypesEnvConfig(&conf.SkipProcessingFormats, "IMGPROXY_SKIP_PROCESSING_FORMATS")

	boolEnvConfig(&conf.UseLinearColorspace, "IMGPROXY_USE_LINEAR_COLORSPACE")
	boolEnvConfig(&conf.DisableShrinkOnLoad, "IMGPROXY_DISABLE_SHRINK_ON_LOAD")

	if err := hexEnvConfig(&conf.Keys, "IMGPROXY_KEY"); err != nil {
		return err
	}
	if err := hexEnvConfig(&conf.Salts, "IMGPROXY_SALT"); err != nil {
		return err
	}
	intEnvConfig(&conf.SignatureSize, "IMGPROXY_SIGNATURE_SIZE")

	if err := hexFileConfig(&conf.Keys, *keyPath); err != nil {
		return err
	}
	if err := hexFileConfig(&conf.Salts, *saltPath); err != nil {
		return err
	}

	strEnvConfig(&conf.Secret, "IMGPROXY_SECRET")

	strEnvConfig(&conf.AllowOrigin, "IMGPROXY_ALLOW_ORIGIN")

	strEnvConfig(&conf.UserAgent, "IMGPROXY_USER_AGENT")

	boolEnvConfig(&conf.IgnoreSslVerification, "IMGPROXY_IGNORE_SSL_VERIFICATION")
	boolEnvConfig(&conf.DevelopmentErrorsMode, "IMGPROXY_DEVELOPMENT_ERRORS_MODE")

	strEnvConfig(&conf.LocalFileSystemRoot, "IMGPROXY_LOCAL_FILESYSTEM_ROOT")

	boolEnvConfig(&conf.S3Enabled, "IMGPROXY_USE_S3")
	strEnvConfig(&conf.S3Region, "IMGPROXY_S3_REGION")
	strEnvConfig(&conf.S3Endpoint, "IMGPROXY_S3_ENDPOINT")

	boolEnvConfig(&conf.GCSEnabled, "IMGPROXY_USE_GCS")
	strEnvConfig(&conf.GCSKey, "IMGPROXY_GCS_KEY")

	boolEnvConfig(&conf.ABSEnabled, "IMGPROXY_USE_ABS")
	strEnvConfig(&conf.ABSName, "IMGPROXY_ABS_NAME")
	strEnvConfig(&conf.ABSKey, "IMGPROXY_ABS_KEY")
	strEnvConfig(&conf.ABSEndpoint, "IMGPROXY_ABS_ENDPOINT")

	boolEnvConfig(&conf.ETagEnabled, "IMGPROXY_USE_ETAG")

	strEnvConfig(&conf.BaseURL, "IMGPROXY_BASE_URL")

	if err := presetEnvConfig(conf.Presets, "IMGPROXY_PRESETS"); err != nil {
		return err
	}
	if err := presetFileConfig(conf.Presets, *presetsPath); err != nil {
		return err
	}
	boolEnvConfig(&conf.OnlyPresets, "IMGPROXY_ONLY_PRESETS")

	strEnvConfig(&conf.WatermarkData, "IMGPROXY_WATERMARK_DATA")
	strEnvConfig(&conf.WatermarkPath, "IMGPROXY_WATERMARK_PATH")
	strEnvConfig(&conf.WatermarkURL, "IMGPROXY_WATERMARK_URL")
	floatEnvConfig(&conf.WatermarkOpacity, "IMGPROXY_WATERMARK_OPACITY")

	strEnvConfig(&conf.FallbackImageData, "IMGPROXY_FALLBACK_IMAGE_DATA")
	strEnvConfig(&conf.FallbackImagePath, "IMGPROXY_FALLBACK_IMAGE_PATH")
	strEnvConfig(&conf.FallbackImageURL, "IMGPROXY_FALLBACK_IMAGE_URL")

	strEnvConfig(&conf.NewRelicAppName, "IMGPROXY_NEW_RELIC_APP_NAME")
	strEnvConfig(&conf.NewRelicKey, "IMGPROXY_NEW_RELIC_KEY")

	strEnvConfig(&conf.PrometheusBind, "IMGPROXY_PROMETHEUS_BIND")
	strEnvConfig(&conf.PrometheusNamespace, "IMGPROXY_PROMETHEUS_NAMESPACE")

	strEnvConfig(&conf.BugsnagKey, "IMGPROXY_BUGSNAG_KEY")
	strEnvConfig(&conf.BugsnagStage, "IMGPROXY_BUGSNAG_STAGE")
	strEnvConfig(&conf.HoneybadgerKey, "IMGPROXY_HONEYBADGER_KEY")
	strEnvConfig(&conf.HoneybadgerEnv, "IMGPROXY_HONEYBADGER_ENV")
	strEnvConfig(&conf.SentryDSN, "IMGPROXY_SENTRY_DSN")
	strEnvConfig(&conf.SentryEnvironment, "IMGPROXY_SENTRY_ENVIRONMENT")
	strEnvConfig(&conf.SentryRelease, "IMGPROXY_SENTRY_RELEASE")
	intEnvConfig(&conf.AirbrakeProjecID, "IMGPROXY_AIRBRAKE_PROJECT_ID")
	strEnvConfig(&conf.AirbrakeProjecKey, "IMGPROXY_AIRBRAKE_PROJECT_KEY")
	strEnvConfig(&conf.AirbrakeEnv, "IMGPROXY_AIRBRAKE_ENVIRONMENT")
	boolEnvConfig(&conf.ReportDownloadingErrors, "IMGPROXY_REPORT_DOWNLOADING_ERRORS")
	boolEnvConfig(&conf.EnableDebugHeaders, "IMGPROXY_ENABLE_DEBUG_HEADERS")

	intEnvConfig(&conf.FreeMemoryInterval, "IMGPROXY_FREE_MEMORY_INTERVAL")
	intEnvConfig(&conf.DownloadBufferSize, "IMGPROXY_DOWNLOAD_BUFFER_SIZE")
	intEnvConfig(&conf.GZipBufferSize, "IMGPROXY_GZIP_BUFFER_SIZE")
	intEnvConfig(&conf.BufferPoolCalibrationThreshold, "IMGPROXY_BUFFER_POOL_CALIBRATION_THRESHOLD")

	if len(conf.Keys) != len(conf.Salts) {
		return fmt.Errorf("Number of keys and number of salts should be equal. Keys: %d, salts: %d", len(conf.Keys), len(conf.Salts))
	}
	if len(conf.Keys) == 0 {
		logWarning("No keys defined, so signature checking is disabled")
		conf.AllowInsecure = true
	}
	if len(conf.Salts) == 0 {
		logWarning("No salts defined, so signature checking is disabled")
		conf.AllowInsecure = true
	}

	if conf.SignatureSize < 1 || conf.SignatureSize > 32 {
		return fmt.Errorf("Signature size should be within 1 and 32, now - %d\n", conf.SignatureSize)
	}

	if len(conf.Bind) == 0 {
		return fmt.Errorf("Bind address is not defined")
	}

	if conf.ReadTimeout <= 0 {
		return fmt.Errorf("Read timeout should be greater than 0, now - %d\n", conf.ReadTimeout)
	}

	if conf.WriteTimeout <= 0 {
		return fmt.Errorf("Write timeout should be greater than 0, now - %d\n", conf.WriteTimeout)
	}
	if conf.KeepAliveTimeout < 0 {
		return fmt.Errorf("KeepAlive timeout should be greater than or equal to 0, now - %d\n", conf.KeepAliveTimeout)
	}

	if conf.DownloadTimeout <= 0 {
		return fmt.Errorf("Download timeout should be greater than 0, now - %d\n", conf.DownloadTimeout)
	}

	if conf.Concurrency <= 0 {
		return fmt.Errorf("Concurrency should be greater than 0, now - %d\n", conf.Concurrency)
	}

	if conf.MaxClients <= 0 {
		conf.MaxClients = conf.Concurrency * 10
	}

	if conf.TTL <= 0 {
		return fmt.Errorf("TTL should be greater than 0, now - %d\n", conf.TTL)
	}

	if conf.MaxSrcDimension < 0 {
		return fmt.Errorf("Max src dimension should be greater than or equal to 0, now - %d\n", conf.MaxSrcDimension)
	} else if conf.MaxSrcDimension > 0 {
		logWarning("IMGPROXY_MAX_SRC_DIMENSION is deprecated and can be removed in future versions. Use IMGPROXY_MAX_SRC_RESOLUTION")
	}

	if conf.MaxSrcResolution <= 0 {
		return fmt.Errorf("Max src resolution should be greater than 0, now - %d\n", conf.MaxSrcResolution)
	}

	if conf.MaxSrcFileSize < 0 {
		return fmt.Errorf("Max src file size should be greater than or equal to 0, now - %d\n", conf.MaxSrcFileSize)
	}

	if conf.MaxAnimationFrames <= 0 {
		return fmt.Errorf("Max animation frames should be greater than 0, now - %d\n", conf.MaxAnimationFrames)
	}

	if conf.PngQuantizationColors < 2 {
		return fmt.Errorf("Png quantization colors should be greater than 1, now - %d\n", conf.PngQuantizationColors)
	} else if conf.PngQuantizationColors > 256 {
		return fmt.Errorf("Png quantization colors can't be greater than 256, now - %d\n", conf.PngQuantizationColors)
	}

	if conf.Quality <= 0 {
		return fmt.Errorf("Quality should be greater than 0, now - %d\n", conf.Quality)
	} else if conf.Quality > 100 {
		return fmt.Errorf("Quality can't be greater than 100, now - %d\n", conf.Quality)
	}

	if conf.AvifSpeed <= 0 {
		return fmt.Errorf("Avif speed should be greater than 0, now - %d\n", conf.AvifSpeed)
	} else if conf.AvifSpeed > 8 {
		return fmt.Errorf("Avif speed can't be greater than 8, now - %d\n", conf.AvifSpeed)
	}

	if conf.GZipCompression < 0 {
		return fmt.Errorf("GZip compression should be greater than or equal to 0, now - %d\n", conf.GZipCompression)
	} else if conf.GZipCompression > 9 {
		return fmt.Errorf("GZip compression can't be greater than 9, now - %d\n", conf.GZipCompression)
	}

	if conf.GZipCompression > 0 {
		logWarning("GZip compression is deprecated and can be removed in future versions")
	}

	if conf.IgnoreSslVerification {
		logWarning("Ignoring SSL verification is very unsafe")
	}

	if conf.LocalFileSystemRoot != "" {
		stat, err := os.Stat(conf.LocalFileSystemRoot)

		if err != nil {
			return fmt.Errorf("Cannot use local directory: %s", err)
		}

		if !stat.IsDir() {
			return fmt.Errorf("Cannot use local directory: not a directory")
		}

		if conf.LocalFileSystemRoot == "/" {
			logWarning("Exposing root via IMGPROXY_LOCAL_FILESYSTEM_ROOT is unsafe")
		}
	}

	if _, ok := os.LookupEnv("IMGPROXY_USE_GCS"); !ok && len(conf.GCSKey) > 0 {
		logWarning("Set IMGPROXY_USE_GCS to true since it may be required by future versions to enable GCS support")
		conf.GCSEnabled = true
	}

	if conf.WatermarkOpacity <= 0 {
		return fmt.Errorf("Watermark opacity should be greater than 0")
	} else if conf.WatermarkOpacity > 1 {
		return fmt.Errorf("Watermark opacity should be less than or equal to 1")
	}

	if len(conf.PrometheusBind) > 0 && conf.PrometheusBind == conf.Bind {
		return fmt.Errorf("Can't use the same binding for the main server and Prometheus")
	}

	if conf.FreeMemoryInterval <= 0 {
		return fmt.Errorf("Free memory interval should be greater than zero")
	}

	if conf.DownloadBufferSize < 0 {
		return fmt.Errorf("Download buffer size should be greater than or equal to 0")
	} else if conf.DownloadBufferSize > math.MaxInt32 {
		return fmt.Errorf("Download buffer size can't be greater than %d", math.MaxInt32)
	}

	if conf.GZipBufferSize < 0 {
		return fmt.Errorf("GZip buffer size should be greater than or equal to 0")
	} else if conf.GZipBufferSize > math.MaxInt32 {
		return fmt.Errorf("GZip buffer size can't be greater than %d", math.MaxInt32)
	}

	if conf.BufferPoolCalibrationThreshold < 64 {
		return fmt.Errorf("Buffer pool calibration threshold should be greater than or equal to 64")
	}

	return nil
}
