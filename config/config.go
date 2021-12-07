package config

import (
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config/configurators"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/version"
)

var (
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
	FormatQuality         map[imagetype.Type]int
	StripMetadata         bool
	StripColorProfile     bool
	AutoRotate            bool

	EnableWebpDetection bool
	EnforceWebp         bool
	EnableAvifDetection bool
	EnforceAvif         bool
	EnableClientHints   bool

	SkipProcessingFormats []imagetype.Type

	UseLinearColorspace bool
	DisableShrinkOnLoad bool

	Keys          [][]byte
	Salts         [][]byte
	SignatureSize int

	Secret string

	AllowOrigin string

	UserAgent string

	IgnoreSslVerification bool
	DevelopmentErrorsMode bool

	AllowedSources []*regexp.Regexp

	CookiePassthrough bool
	CookieBaseURL     string

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
	ETagBuster  string

	BaseURL string

	Presets     []string
	OnlyPresets bool

	WatermarkData    string
	WatermarkPath    string
	WatermarkURL     string
	WatermarkOpacity float64

	FallbackImageData     string
	FallbackImagePath     string
	FallbackImageURL      string
	FallbackImageHTTPCode int

	DataDogEnable bool

	NewRelicAppName string
	NewRelicKey     string

	PrometheusBind      string
	PrometheusNamespace string

	BugsnagKey   string
	BugsnagStage string

	HoneybadgerKey string
	HoneybadgerEnv string

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
	BufferPoolCalibrationThreshold int
)

var (
	keyPath     string
	saltPath    string
	presetsPath string
)

func init() {
	Reset()

	flag.StringVar(&keyPath, "keypath", "", "path of the file with hex-encoded key")
	flag.StringVar(&saltPath, "saltpath", "", "path of the file with hex-encoded salt")
	flag.StringVar(&presetsPath, "presets", "", "path of the file with presets")
}

func Reset() {
	Network = "tcp"
	Bind = ":8080"
	ReadTimeout = 10
	WriteTimeout = 10
	KeepAliveTimeout = 10
	DownloadTimeout = 5
	Concurrency = runtime.NumCPU() * 2
	MaxClients = 0

	TTL = 3600
	CacheControlPassthrough = false
	SetCanonicalHeader = false

	SoReuseport = false

	PathPrefix = ""

	MaxSrcResolution = 16800000
	MaxSrcFileSize = 0
	MaxAnimationFrames = 1
	MaxSvgCheckBytes = 32 * 1024

	JpegProgressive = false
	PngInterlaced = false
	PngQuantize = false
	PngQuantizationColors = 256
	AvifSpeed = 5
	Quality = 80
	FormatQuality = map[imagetype.Type]int{imagetype.AVIF: 50}
	StripMetadata = true
	StripColorProfile = true
	AutoRotate = true

	EnableWebpDetection = false
	EnforceWebp = false
	EnableAvifDetection = false
	EnforceAvif = false
	EnableClientHints = false

	SkipProcessingFormats = make([]imagetype.Type, 0)

	UseLinearColorspace = false
	DisableShrinkOnLoad = false

	Keys = make([][]byte, 0)
	Salts = make([][]byte, 0)
	SignatureSize = 32

	Secret = ""

	AllowOrigin = ""

	UserAgent = fmt.Sprintf("imgproxy/%s", version.Version())

	IgnoreSslVerification = false
	DevelopmentErrorsMode = false

	AllowedSources = make([]*regexp.Regexp, 0)

	CookiePassthrough = false
	CookieBaseURL = ""

	LocalFileSystemRoot = ""
	S3Enabled = false
	S3Region = ""
	S3Endpoint = ""
	GCSEnabled = false
	GCSKey = ""
	ABSEnabled = false
	ABSName = ""
	ABSKey = ""
	ABSEndpoint = ""

	ETagEnabled = false
	ETagBuster = ""

	BaseURL = ""

	Presets = make([]string, 0)
	OnlyPresets = false

	WatermarkData = ""
	WatermarkPath = ""
	WatermarkURL = ""
	WatermarkOpacity = 1

	FallbackImageData = ""
	FallbackImagePath = ""
	FallbackImageURL = ""
	FallbackImageHTTPCode = 200

	DataDogEnable = false

	NewRelicAppName = ""
	NewRelicKey = ""

	PrometheusBind = ""
	PrometheusNamespace = ""

	BugsnagKey = ""
	BugsnagStage = "production"

	HoneybadgerKey = ""
	HoneybadgerEnv = "production"

	SentryDSN = ""
	SentryEnvironment = "production"
	SentryRelease = fmt.Sprintf("imgproxy@%s", version.Version())

	AirbrakeProjecID = 0
	AirbrakeProjecKey = ""
	AirbrakeEnv = "production"

	ReportDownloadingErrors = true

	EnableDebugHeaders = false

	FreeMemoryInterval = 10
	DownloadBufferSize = 0
	BufferPoolCalibrationThreshold = 1024
}

func Configure() error {
	if port := os.Getenv("PORT"); len(port) > 0 {
		Bind = fmt.Sprintf(":%s", port)
	}

	configurators.String(&Network, "IMGPROXY_NETWORK")
	configurators.String(&Bind, "IMGPROXY_BIND")
	configurators.Int(&ReadTimeout, "IMGPROXY_READ_TIMEOUT")
	configurators.Int(&WriteTimeout, "IMGPROXY_WRITE_TIMEOUT")
	configurators.Int(&KeepAliveTimeout, "IMGPROXY_KEEP_ALIVE_TIMEOUT")
	configurators.Int(&DownloadTimeout, "IMGPROXY_DOWNLOAD_TIMEOUT")
	configurators.Int(&Concurrency, "IMGPROXY_CONCURRENCY")
	configurators.Int(&MaxClients, "IMGPROXY_MAX_CLIENTS")

	configurators.Int(&TTL, "IMGPROXY_TTL")
	configurators.Bool(&CacheControlPassthrough, "IMGPROXY_CACHE_CONTROL_PASSTHROUGH")
	configurators.Bool(&SetCanonicalHeader, "IMGPROXY_SET_CANONICAL_HEADER")

	configurators.Bool(&SoReuseport, "IMGPROXY_SO_REUSEPORT")

	configurators.String(&PathPrefix, "IMGPROXY_PATH_PREFIX")

	configurators.MegaInt(&MaxSrcResolution, "IMGPROXY_MAX_SRC_RESOLUTION")
	configurators.Int(&MaxSrcFileSize, "IMGPROXY_MAX_SRC_FILE_SIZE")
	configurators.Int(&MaxSvgCheckBytes, "IMGPROXY_MAX_SVG_CHECK_BYTES")

	configurators.Int(&MaxAnimationFrames, "IMGPROXY_MAX_ANIMATION_FRAMES")

	configurators.Patterns(&AllowedSources, "IMGPROXY_ALLOWED_SOURCES")

	configurators.Bool(&JpegProgressive, "IMGPROXY_JPEG_PROGRESSIVE")
	configurators.Bool(&PngInterlaced, "IMGPROXY_PNG_INTERLACED")
	configurators.Bool(&PngQuantize, "IMGPROXY_PNG_QUANTIZE")
	configurators.Int(&PngQuantizationColors, "IMGPROXY_PNG_QUANTIZATION_COLORS")
	configurators.Int(&AvifSpeed, "IMGPROXY_AVIF_SPEED")
	configurators.Int(&Quality, "IMGPROXY_QUALITY")
	if err := configurators.ImageTypesQuality(FormatQuality, "IMGPROXY_FORMAT_QUALITY"); err != nil {
		return err
	}
	configurators.Bool(&StripMetadata, "IMGPROXY_STRIP_METADATA")
	configurators.Bool(&StripColorProfile, "IMGPROXY_STRIP_COLOR_PROFILE")
	configurators.Bool(&AutoRotate, "IMGPROXY_AUTO_ROTATE")

	configurators.Bool(&EnableWebpDetection, "IMGPROXY_ENABLE_WEBP_DETECTION")
	configurators.Bool(&EnforceWebp, "IMGPROXY_ENFORCE_WEBP")
	configurators.Bool(&EnableAvifDetection, "IMGPROXY_ENABLE_AVIF_DETECTION")
	configurators.Bool(&EnforceAvif, "IMGPROXY_ENFORCE_AVIF")
	configurators.Bool(&EnableClientHints, "IMGPROXY_ENABLE_CLIENT_HINTS")

	if err := configurators.ImageTypes(&SkipProcessingFormats, "IMGPROXY_SKIP_PROCESSING_FORMATS"); err != nil {
		return err
	}

	configurators.Bool(&UseLinearColorspace, "IMGPROXY_USE_LINEAR_COLORSPACE")
	configurators.Bool(&DisableShrinkOnLoad, "IMGPROXY_DISABLE_SHRINK_ON_LOAD")

	if err := configurators.Hex(&Keys, "IMGPROXY_KEY"); err != nil {
		return err
	}
	if err := configurators.Hex(&Salts, "IMGPROXY_SALT"); err != nil {
		return err
	}
	configurators.Int(&SignatureSize, "IMGPROXY_SIGNATURE_SIZE")

	if err := configurators.HexFile(&Keys, keyPath); err != nil {
		return err
	}
	if err := configurators.HexFile(&Salts, saltPath); err != nil {
		return err
	}

	configurators.String(&Secret, "IMGPROXY_SECRET")

	configurators.String(&AllowOrigin, "IMGPROXY_ALLOW_ORIGIN")

	configurators.String(&UserAgent, "IMGPROXY_USER_AGENT")

	configurators.Bool(&IgnoreSslVerification, "IMGPROXY_IGNORE_SSL_VERIFICATION")
	configurators.Bool(&DevelopmentErrorsMode, "IMGPROXY_DEVELOPMENT_ERRORS_MODE")

	configurators.Bool(&CookiePassthrough, "IMGPROXY_COOKIE_PASSTHROUGH")
	configurators.String(&CookieBaseURL, "IMGPROXY_COOKIE_BASE_URL")

	configurators.String(&LocalFileSystemRoot, "IMGPROXY_LOCAL_FILESYSTEM_ROOT")

	configurators.Bool(&S3Enabled, "IMGPROXY_USE_S3")
	configurators.String(&S3Region, "IMGPROXY_S3_REGION")
	configurators.String(&S3Endpoint, "IMGPROXY_S3_ENDPOINT")

	configurators.Bool(&GCSEnabled, "IMGPROXY_USE_GCS")
	configurators.String(&GCSKey, "IMGPROXY_GCS_KEY")

	configurators.Bool(&ABSEnabled, "IMGPROXY_USE_ABS")
	configurators.String(&ABSName, "IMGPROXY_ABS_NAME")
	configurators.String(&ABSKey, "IMGPROXY_ABS_KEY")
	configurators.String(&ABSEndpoint, "IMGPROXY_ABS_ENDPOINT")

	configurators.Bool(&ETagEnabled, "IMGPROXY_USE_ETAG")
	configurators.String(&ETagBuster, "IMGPROXY_ETAG_BUSTER")

	configurators.String(&BaseURL, "IMGPROXY_BASE_URL")

	configurators.StringSlice(&Presets, "IMGPROXY_PRESETS")
	if err := configurators.StringSliceFile(&Presets, presetsPath); err != nil {
		return err
	}
	configurators.Bool(&OnlyPresets, "IMGPROXY_ONLY_PRESETS")

	configurators.String(&WatermarkData, "IMGPROXY_WATERMARK_DATA")
	configurators.String(&WatermarkPath, "IMGPROXY_WATERMARK_PATH")
	configurators.String(&WatermarkURL, "IMGPROXY_WATERMARK_URL")
	configurators.Float(&WatermarkOpacity, "IMGPROXY_WATERMARK_OPACITY")

	configurators.String(&FallbackImageData, "IMGPROXY_FALLBACK_IMAGE_DATA")
	configurators.String(&FallbackImagePath, "IMGPROXY_FALLBACK_IMAGE_PATH")
	configurators.String(&FallbackImageURL, "IMGPROXY_FALLBACK_IMAGE_URL")
	configurators.Int(&FallbackImageHTTPCode, "IMGPROXY_FALLBACK_IMAGE_HTTP_CODE")

	configurators.Bool(&DataDogEnable, "IMGPROXY_DATADOG_ENABLE")

	configurators.String(&NewRelicAppName, "IMGPROXY_NEW_RELIC_APP_NAME")
	configurators.String(&NewRelicKey, "IMGPROXY_NEW_RELIC_KEY")

	configurators.String(&PrometheusBind, "IMGPROXY_PROMETHEUS_BIND")
	configurators.String(&PrometheusNamespace, "IMGPROXY_PROMETHEUS_NAMESPACE")

	configurators.String(&BugsnagKey, "IMGPROXY_BUGSNAG_KEY")
	configurators.String(&BugsnagStage, "IMGPROXY_BUGSNAG_STAGE")
	configurators.String(&HoneybadgerKey, "IMGPROXY_HONEYBADGER_KEY")
	configurators.String(&HoneybadgerEnv, "IMGPROXY_HONEYBADGER_ENV")
	configurators.String(&SentryDSN, "IMGPROXY_SENTRY_DSN")
	configurators.String(&SentryEnvironment, "IMGPROXY_SENTRY_ENVIRONMENT")
	configurators.String(&SentryRelease, "IMGPROXY_SENTRY_RELEASE")
	configurators.Int(&AirbrakeProjecID, "IMGPROXY_AIRBRAKE_PROJECT_ID")
	configurators.String(&AirbrakeProjecKey, "IMGPROXY_AIRBRAKE_PROJECT_KEY")
	configurators.String(&AirbrakeEnv, "IMGPROXY_AIRBRAKE_ENVIRONMENT")
	configurators.Bool(&ReportDownloadingErrors, "IMGPROXY_REPORT_DOWNLOADING_ERRORS")
	configurators.Bool(&EnableDebugHeaders, "IMGPROXY_ENABLE_DEBUG_HEADERS")

	configurators.Int(&FreeMemoryInterval, "IMGPROXY_FREE_MEMORY_INTERVAL")
	configurators.Int(&DownloadBufferSize, "IMGPROXY_DOWNLOAD_BUFFER_SIZE")
	configurators.Int(&BufferPoolCalibrationThreshold, "IMGPROXY_BUFFER_POOL_CALIBRATION_THRESHOLD")

	if len(Keys) != len(Salts) {
		return fmt.Errorf("Number of keys and number of salts should be equal. Keys: %d, salts: %d", len(Keys), len(Salts))
	}
	if len(Keys) == 0 {
		log.Warning("No keys defined, so signature checking is disabled")
	}
	if len(Salts) == 0 {
		log.Warning("No salts defined, so signature checking is disabled")
	}

	if SignatureSize < 1 || SignatureSize > 32 {
		return fmt.Errorf("Signature size should be within 1 and 32, now - %d\n", SignatureSize)
	}

	if len(Bind) == 0 {
		return fmt.Errorf("Bind address is not defined")
	}

	if ReadTimeout <= 0 {
		return fmt.Errorf("Read timeout should be greater than 0, now - %d\n", ReadTimeout)
	}

	if WriteTimeout <= 0 {
		return fmt.Errorf("Write timeout should be greater than 0, now - %d\n", WriteTimeout)
	}
	if KeepAliveTimeout < 0 {
		return fmt.Errorf("KeepAlive timeout should be greater than or equal to 0, now - %d\n", KeepAliveTimeout)
	}

	if DownloadTimeout <= 0 {
		return fmt.Errorf("Download timeout should be greater than 0, now - %d\n", DownloadTimeout)
	}

	if Concurrency <= 0 {
		return fmt.Errorf("Concurrency should be greater than 0, now - %d\n", Concurrency)
	}

	if MaxClients <= 0 {
		MaxClients = Concurrency * 10
	}

	if TTL <= 0 {
		return fmt.Errorf("TTL should be greater than 0, now - %d\n", TTL)
	}

	if MaxSrcResolution <= 0 {
		return fmt.Errorf("Max src resolution should be greater than 0, now - %d\n", MaxSrcResolution)
	}

	if MaxSrcFileSize < 0 {
		return fmt.Errorf("Max src file size should be greater than or equal to 0, now - %d\n", MaxSrcFileSize)
	}

	if MaxAnimationFrames <= 0 {
		return fmt.Errorf("Max animation frames should be greater than 0, now - %d\n", MaxAnimationFrames)
	}

	if PngQuantizationColors < 2 {
		return fmt.Errorf("Png quantization colors should be greater than 1, now - %d\n", PngQuantizationColors)
	} else if PngQuantizationColors > 256 {
		return fmt.Errorf("Png quantization colors can't be greater than 256, now - %d\n", PngQuantizationColors)
	}

	if AvifSpeed <= 0 {
		return fmt.Errorf("Avif speed should be greater than 0, now - %d\n", AvifSpeed)
	} else if AvifSpeed > 8 {
		return fmt.Errorf("Avif speed can't be greater than 8, now - %d\n", AvifSpeed)
	}

	if Quality <= 0 {
		return fmt.Errorf("Quality should be greater than 0, now - %d\n", Quality)
	} else if Quality > 100 {
		return fmt.Errorf("Quality can't be greater than 100, now - %d\n", Quality)
	}

	if IgnoreSslVerification {
		log.Warning("Ignoring SSL verification is very unsafe")
	}

	if LocalFileSystemRoot != "" {
		stat, err := os.Stat(LocalFileSystemRoot)

		if err != nil {
			return fmt.Errorf("Cannot use local directory: %s", err)
		}

		if !stat.IsDir() {
			return fmt.Errorf("Cannot use local directory: not a directory")
		}

		if LocalFileSystemRoot == "/" {
			log.Warning("Exposing root via IMGPROXY_LOCAL_FILESYSTEM_ROOT is unsafe")
		}
	}

	if _, ok := os.LookupEnv("IMGPROXY_USE_GCS"); !ok && len(GCSKey) > 0 {
		log.Warning("Set IMGPROXY_USE_GCS to true since it may be required by future versions to enable GCS support")
		GCSEnabled = true
	}

	if WatermarkOpacity <= 0 {
		return fmt.Errorf("Watermark opacity should be greater than 0")
	} else if WatermarkOpacity > 1 {
		return fmt.Errorf("Watermark opacity should be less than or equal to 1")
	}

	if FallbackImageHTTPCode < 100 || FallbackImageHTTPCode > 599 {
		return fmt.Errorf("Fallback image HTTP code should be between 100 and 599")
	}

	if len(PrometheusBind) > 0 && PrometheusBind == Bind {
		return fmt.Errorf("Can't use the same binding for the main server and Prometheus")
	}

	if FreeMemoryInterval <= 0 {
		return fmt.Errorf("Free memory interval should be greater than zero")
	}

	if DownloadBufferSize < 0 {
		return fmt.Errorf("Download buffer size should be greater than or equal to 0")
	} else if DownloadBufferSize > math.MaxInt32 {
		return fmt.Errorf("Download buffer size can't be greater than %d", math.MaxInt32)
	}

	if BufferPoolCalibrationThreshold < 64 {
		return fmt.Errorf("Buffer pool calibration threshold should be greater than or equal to 64")
	}

	return nil
}
