package config

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config/configurators"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/logger"
	"github.com/imgproxy/imgproxy/v3/version"
)

type URLReplacement = configurators.URLReplacement

var (
	Network                string
	Bind                   string
	Timeout                int
	GracefulStopTimeout    int
	ReadRequestTimeout     int
	WriteResponseTimeout   int
	KeepAliveTimeout       int
	ClientKeepAliveTimeout int
	DownloadTimeout        int
	Workers                int
	RequestsQueueSize      int
	MaxClients             int

	TTL                     int
	CacheControlPassthrough bool
	SetCanonicalHeader      bool

	SoReuseport bool

	PathPrefix string

	MaxSrcResolution            int
	MaxSrcFileSize              int
	MaxAnimationFrames          int
	MaxAnimationFrameResolution int
	MaxSvgCheckBytes            int
	MaxRedirects                int
	PngUnlimited                bool
	SvgUnlimited                bool
	MaxResultDimension          int
	AllowedProcessiongOptions   []string
	AllowSecurityOptions        bool

	JpegProgressive       bool
	PngInterlaced         bool
	PngQuantize           bool
	PngQuantizationColors int
	AvifSpeed             int
	JxlEffort             int
	WebpEffort            int
	WebpPreset            WebpPresetKind
	Quality               int
	FormatQuality         map[imagetype.Type]int
	StripMetadata         bool
	KeepCopyright         bool
	StripColorProfile     bool
	AutoRotate            bool
	EnforceThumbnail      bool
	ReturnAttachment      bool

	AutoWebp          bool
	EnforceWebp       bool
	AutoAvif          bool
	EnforceAvif       bool
	AutoJxl           bool
	EnforceJxl        bool
	EnableClientHints bool

	PreferredFormats []imagetype.Type

	SkipProcessingFormats []imagetype.Type

	UseLinearColorspace bool
	DisableShrinkOnLoad bool

	Keys              [][]byte
	Salts             [][]byte
	SignatureSize     int
	TrustedSignatures []string

	Secret string

	AllowOrigin string

	UserAgent string

	IgnoreSslVerification bool
	DevelopmentErrorsMode bool

	AllowedSources                []*regexp.Regexp
	AllowLoopbackSourceAddresses  bool
	AllowLinkLocalSourceAddresses bool
	AllowPrivateSourceAddresses   bool

	SanitizeSvg        bool
	AlwaysRasterizeSvg bool

	CookiePassthrough    bool
	CookieBaseURL        string
	CookiePassthroughAll bool

	SourceURLQuerySeparator string

	LocalFileSystemRoot string

	S3Enabled                 bool
	S3Region                  string
	S3Endpoint                string
	S3EndpointUsePathStyle    bool
	S3AssumeRoleArn           string
	S3AssumeRoleExternalID    string
	S3DecryptionClientEnabled bool

	GCSEnabled  bool
	GCSKey      string
	GCSEndpoint string

	ABSEnabled  bool
	ABSName     string
	ABSKey      string
	ABSEndpoint string

	SwiftEnabled               bool
	SwiftUsername              string
	SwiftAPIKey                string
	SwiftAuthURL               string
	SwiftDomain                string
	SwiftTenant                string
	SwiftAuthVersion           int
	SwiftConnectTimeoutSeconds int
	SwiftTimeoutSeconds        int

	ETagEnabled bool
	ETagBuster  string

	LastModifiedEnabled bool

	BaseURL                   string
	URLReplacements           []URLReplacement
	Base64URLIncludesFilename bool

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
	FallbackImageTTL      int

	DataDogEnable        bool
	DataDogEnableMetrics bool

	NewRelicAppName string
	NewRelicKey     string
	NewRelicLabels  map[string]string

	PrometheusBind      string
	PrometheusNamespace string

	OpenTelemetryEnable           bool
	OpenTelemetryEnableMetrics    bool
	OpenTelemetryServerCert       string
	OpenTelemetryClientCert       string
	OpenTelemetryClientKey        string
	OpenTelemetryTraceIDGenerator string

	CloudWatchServiceName string
	CloudWatchNamespace   string
	CloudWatchRegion      string

	BugsnagKey   string
	BugsnagStage string

	HoneybadgerKey string
	HoneybadgerEnv string

	SentryDSN         string
	SentryEnvironment string
	SentryRelease     string

	AirbrakeProjectID  int
	AirbrakeProjectKey string
	AirbrakeEnv        string

	ReportDownloadingErrors bool
	ReportIOErrors          bool

	EnableDebugHeaders bool

	FreeMemoryInterval             int
	DownloadBufferSize             int
	BufferPoolCalibrationThreshold int

	HealthCheckPath string

	ArgumentsSeparator string
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
	Timeout = 10
	GracefulStopTimeout = 20
	ReadRequestTimeout = 10
	WriteResponseTimeout = 10
	KeepAliveTimeout = 10
	ClientKeepAliveTimeout = 90
	DownloadTimeout = 5
	Workers = runtime.GOMAXPROCS(0) * 2
	RequestsQueueSize = 0
	MaxClients = 2048

	TTL = 31536000
	CacheControlPassthrough = false
	SetCanonicalHeader = false

	SoReuseport = false

	PathPrefix = ""

	MaxSrcResolution = 50000000
	MaxSrcFileSize = 0
	MaxAnimationFrames = 1
	MaxAnimationFrameResolution = 0
	MaxSvgCheckBytes = 32 * 1024
	MaxRedirects = 10
	PngUnlimited = false
	SvgUnlimited = false
	MaxResultDimension = 0
	AllowedProcessiongOptions = make([]string, 0)
	AllowSecurityOptions = false

	JpegProgressive = false
	PngInterlaced = false
	PngQuantize = false
	PngQuantizationColors = 256
	AvifSpeed = 8
	JxlEffort = 4
	WebpEffort = 4
	WebpPreset = WebpPresetDefault
	Quality = 80
	FormatQuality = map[imagetype.Type]int{
		imagetype.WEBP: 79,
		imagetype.AVIF: 63,
		imagetype.JXL:  77,
	}
	StripMetadata = true
	KeepCopyright = true
	StripColorProfile = true
	AutoRotate = true
	EnforceThumbnail = false
	ReturnAttachment = false

	AutoWebp = false
	EnforceWebp = false
	AutoAvif = false
	EnforceAvif = false
	AutoJxl = false
	EnforceJxl = false
	EnableClientHints = false

	PreferredFormats = []imagetype.Type{
		imagetype.JPEG,
		imagetype.PNG,
		imagetype.GIF,
	}

	SkipProcessingFormats = make([]imagetype.Type, 0)

	UseLinearColorspace = false
	DisableShrinkOnLoad = false

	Keys = make([][]byte, 0)
	Salts = make([][]byte, 0)
	SignatureSize = 32
	TrustedSignatures = make([]string, 0)

	Secret = ""

	AllowOrigin = ""

	UserAgent = "imgproxy/%current_version"

	IgnoreSslVerification = false
	DevelopmentErrorsMode = false

	AllowedSources = make([]*regexp.Regexp, 0)
	AllowLoopbackSourceAddresses = false
	AllowLinkLocalSourceAddresses = false
	AllowPrivateSourceAddresses = true

	SanitizeSvg = true
	AlwaysRasterizeSvg = false

	CookiePassthrough = false
	CookieBaseURL = ""
	CookiePassthroughAll = false

	SourceURLQuerySeparator = "?"
	LocalFileSystemRoot = ""
	S3Enabled = false
	S3Region = ""
	S3Endpoint = ""
	S3EndpointUsePathStyle = true
	S3AssumeRoleArn = ""
	S3AssumeRoleExternalID = ""
	S3DecryptionClientEnabled = false
	GCSEnabled = false
	GCSKey = ""
	ABSEnabled = false
	ABSName = ""
	ABSKey = ""
	ABSEndpoint = ""
	SwiftEnabled = false
	SwiftUsername = ""
	SwiftAPIKey = ""
	SwiftAuthURL = ""
	SwiftAuthVersion = 0
	SwiftTenant = ""
	SwiftDomain = ""
	SwiftConnectTimeoutSeconds = 10
	SwiftTimeoutSeconds = 60

	ETagEnabled = false
	ETagBuster = ""

	LastModifiedEnabled = false

	BaseURL = ""
	URLReplacements = make([]URLReplacement, 0)
	Base64URLIncludesFilename = false

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
	FallbackImageTTL = 0

	DataDogEnable = false

	NewRelicAppName = ""
	NewRelicKey = ""
	NewRelicLabels = make(map[string]string)

	PrometheusBind = ""
	PrometheusNamespace = ""

	OpenTelemetryEnable = false
	OpenTelemetryEnableMetrics = false
	OpenTelemetryServerCert = ""
	OpenTelemetryClientCert = ""
	OpenTelemetryClientKey = ""
	OpenTelemetryTraceIDGenerator = "xray"

	CloudWatchServiceName = ""
	CloudWatchNamespace = "imgproxy"
	CloudWatchRegion = ""

	BugsnagKey = ""
	BugsnagStage = "production"

	HoneybadgerKey = ""
	HoneybadgerEnv = "production"

	SentryDSN = ""
	SentryEnvironment = "production"
	SentryRelease = fmt.Sprintf("imgproxy@%s", version.Version)

	AirbrakeProjectID = 0
	AirbrakeProjectKey = ""
	AirbrakeEnv = "production"

	ReportDownloadingErrors = true
	ReportIOErrors = false

	EnableDebugHeaders = false

	FreeMemoryInterval = 10
	DownloadBufferSize = 0
	BufferPoolCalibrationThreshold = 1024

	HealthCheckPath = ""

	ArgumentsSeparator = ":"
}

func Configure() error {
	Reset()

	if port := os.Getenv("PORT"); len(port) > 0 {
		Bind = fmt.Sprintf(":%s", port)
	}

	configurators.String(&Network, "IMGPROXY_NETWORK")
	configurators.String(&Bind, "IMGPROXY_BIND")

	if _, ok := os.LookupEnv("IMGPROXY_WRITE_TIMEOUT"); ok {
		logger.Deprecated("IMGPROXY_WRITE_TIMEOUT", "IMGPROXY_TIMEOUT")
		configurators.Int(&Timeout, "IMGPROXY_WRITE_TIMEOUT")
	}
	configurators.Int(&Timeout, "IMGPROXY_TIMEOUT")

	GracefulStopTimeout = Timeout * 2
	configurators.Int(&GracefulStopTimeout, "IMGPROXY_GRACEFUL_STOP_TIMEOUT")

	if _, ok := os.LookupEnv("IMGPROXY_READ_TIMEOUT"); ok {
		logger.Deprecated("IMGPROXY_READ_TIMEOUT", "IMGPROXY_READ_REQUEST_TIMEOUT")
		configurators.Int(&ReadRequestTimeout, "IMGPROXY_READ_TIMEOUT")
	}
	configurators.Int(&ReadRequestTimeout, "IMGPROXY_READ_REQUEST_TIMEOUT")

	configurators.Int(&WriteResponseTimeout, "IMGPROXY_WRITE_RESPONSE_TIMEOUT")

	configurators.Int(&KeepAliveTimeout, "IMGPROXY_KEEP_ALIVE_TIMEOUT")
	configurators.Int(&ClientKeepAliveTimeout, "IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT")

	configurators.Int(&DownloadTimeout, "IMGPROXY_DOWNLOAD_TIMEOUT")

	if lambdaFn := os.Getenv("AWS_LAMBDA_FUNCTION_NAME"); len(lambdaFn) > 0 {
		Workers = 1
		log.Info("AWS Lambda environment detected, setting workers to 1")
	} else {
		configurators.Int(&Workers, "IMGPROXY_CONCURRENCY")
		configurators.Int(&Workers, "IMGPROXY_WORKERS")
	}

	configurators.Int(&RequestsQueueSize, "IMGPROXY_REQUESTS_QUEUE_SIZE")
	configurators.Int(&MaxClients, "IMGPROXY_MAX_CLIENTS")

	configurators.Int(&TTL, "IMGPROXY_TTL")
	configurators.Bool(&CacheControlPassthrough, "IMGPROXY_CACHE_CONTROL_PASSTHROUGH")
	configurators.Bool(&SetCanonicalHeader, "IMGPROXY_SET_CANONICAL_HEADER")

	configurators.Bool(&SoReuseport, "IMGPROXY_SO_REUSEPORT")

	configurators.URLPath(&PathPrefix, "IMGPROXY_PATH_PREFIX")

	configurators.MegaInt(&MaxSrcResolution, "IMGPROXY_MAX_SRC_RESOLUTION")
	configurators.Int(&MaxSrcFileSize, "IMGPROXY_MAX_SRC_FILE_SIZE")
	configurators.Int(&MaxSvgCheckBytes, "IMGPROXY_MAX_SVG_CHECK_BYTES")

	configurators.Int(&MaxAnimationFrames, "IMGPROXY_MAX_ANIMATION_FRAMES")
	configurators.MegaInt(&MaxAnimationFrameResolution, "IMGPROXY_MAX_ANIMATION_FRAME_RESOLUTION")

	configurators.Int(&MaxRedirects, "IMGPROXY_MAX_REDIRECTS")

	configurators.Patterns(&AllowedSources, "IMGPROXY_ALLOWED_SOURCES")
	configurators.Bool(&AllowLoopbackSourceAddresses, "IMGPROXY_ALLOW_LOOPBACK_SOURCE_ADDRESSES")
	configurators.Bool(&AllowLinkLocalSourceAddresses, "IMGPROXY_ALLOW_LINK_LOCAL_SOURCE_ADDRESSES")
	configurators.Bool(&AllowPrivateSourceAddresses, "IMGPROXY_ALLOW_PRIVATE_SOURCE_ADDRESSES")

	configurators.Bool(&SanitizeSvg, "IMGPROXY_SANITIZE_SVG")
	configurators.Bool(&AlwaysRasterizeSvg, "IMGPROXY_ALWAYS_RASTERIZE_SVG")

	configurators.Bool(&PngUnlimited, "IMGPROXY_PNG_UNLIMITED")
	configurators.Bool(&SvgUnlimited, "IMGPROXY_SVG_UNLIMITED")

	configurators.Int(&MaxResultDimension, "IMGPROXY_MAX_RESULT_DIMENSION")
	configurators.StringSlice(&AllowedProcessiongOptions, "IMGPROXY_ALLOWED_PROCESSING_OPTIONS")

	configurators.Bool(&AllowSecurityOptions, "IMGPROXY_ALLOW_SECURITY_OPTIONS")

	configurators.Bool(&JpegProgressive, "IMGPROXY_JPEG_PROGRESSIVE")
	configurators.Bool(&PngInterlaced, "IMGPROXY_PNG_INTERLACED")
	configurators.Bool(&PngQuantize, "IMGPROXY_PNG_QUANTIZE")
	configurators.Int(&PngQuantizationColors, "IMGPROXY_PNG_QUANTIZATION_COLORS")
	configurators.Int(&AvifSpeed, "IMGPROXY_AVIF_SPEED")
	configurators.Int(&JxlEffort, "IMGPROXY_JXL_EFFORT")
	configurators.Int(&WebpEffort, "IMGPROXY_WEBP_EFFORT")
	if err := configurators.FromMap(&WebpPreset, "IMGPROXY_WEBP_PRESET", WebpPresets); err != nil {
		return err
	}
	configurators.Int(&Quality, "IMGPROXY_QUALITY")
	if err := configurators.ImageTypesQuality(FormatQuality, "IMGPROXY_FORMAT_QUALITY"); err != nil {
		return err
	}
	configurators.Bool(&StripMetadata, "IMGPROXY_STRIP_METADATA")
	configurators.Bool(&KeepCopyright, "IMGPROXY_KEEP_COPYRIGHT")
	configurators.Bool(&StripColorProfile, "IMGPROXY_STRIP_COLOR_PROFILE")
	configurators.Bool(&AutoRotate, "IMGPROXY_AUTO_ROTATE")
	configurators.Bool(&EnforceThumbnail, "IMGPROXY_ENFORCE_THUMBNAIL")
	configurators.Bool(&ReturnAttachment, "IMGPROXY_RETURN_ATTACHMENT")

	if _, ok := os.LookupEnv("IMGPROXY_ENABLE_WEBP_DETECTION"); ok {
		logger.Deprecated("IMGPROXY_ENABLE_WEBP_DETECTION", "IMGPROXY_AUTO_WEBP")
		configurators.Bool(&AutoWebp, "IMGPROXY_ENABLE_WEBP_DETECTION")
	}
	if _, ok := os.LookupEnv("IMGPROXY_ENABLE_AVIF_DETECTION"); ok {
		logger.Deprecated("IMGPROXY_ENABLE_AVIF_DETECTION", "IMGPROXY_AUTO_AVIF")
		configurators.Bool(&AutoAvif, "IMGPROXY_ENABLE_AVIF_DETECTION")
	}

	configurators.Bool(&AutoWebp, "IMGPROXY_AUTO_WEBP")
	configurators.Bool(&EnforceWebp, "IMGPROXY_ENFORCE_WEBP")
	configurators.Bool(&AutoAvif, "IMGPROXY_AUTO_AVIF")
	configurators.Bool(&EnforceAvif, "IMGPROXY_ENFORCE_AVIF")
	configurators.Bool(&AutoJxl, "IMGPROXY_AUTO_JXL")
	configurators.Bool(&EnforceJxl, "IMGPROXY_ENFORCE_JXL")
	configurators.Bool(&EnableClientHints, "IMGPROXY_ENABLE_CLIENT_HINTS")

	configurators.URLPath(&HealthCheckPath, "IMGPROXY_HEALTH_CHECK_PATH")

	configurators.String(&ArgumentsSeparator, "IMGPROXY_ARGUMENTS_SEPARATOR")

	if err := configurators.ImageTypes(&PreferredFormats, "IMGPROXY_PREFERRED_FORMATS"); err != nil {
		return err
	}

	if err := configurators.ImageTypes(&SkipProcessingFormats, "IMGPROXY_SKIP_PROCESSING_FORMATS"); err != nil {
		return err
	}

	configurators.Bool(&UseLinearColorspace, "IMGPROXY_USE_LINEAR_COLORSPACE")
	configurators.Bool(&DisableShrinkOnLoad, "IMGPROXY_DISABLE_SHRINK_ON_LOAD")

	if err := configurators.HexSlice(&Keys, "IMGPROXY_KEY"); err != nil {
		return err
	}
	if err := configurators.HexSlice(&Salts, "IMGPROXY_SALT"); err != nil {
		return err
	}
	configurators.Int(&SignatureSize, "IMGPROXY_SIGNATURE_SIZE")
	configurators.StringSlice(&TrustedSignatures, "IMGPROXY_TRUSTED_SIGNATURES")

	if err := configurators.HexSliceFile(&Keys, keyPath); err != nil {
		return err
	}
	if err := configurators.HexSliceFile(&Salts, saltPath); err != nil {
		return err
	}

	configurators.String(&Secret, "IMGPROXY_SECRET")

	configurators.String(&AllowOrigin, "IMGPROXY_ALLOW_ORIGIN")

	configurators.String(&UserAgent, "IMGPROXY_USER_AGENT")
	UserAgent = strings.ReplaceAll(UserAgent, "%current_version", version.Version)

	configurators.Bool(&IgnoreSslVerification, "IMGPROXY_IGNORE_SSL_VERIFICATION")
	configurators.Bool(&DevelopmentErrorsMode, "IMGPROXY_DEVELOPMENT_ERRORS_MODE")

	configurators.Bool(&CookiePassthrough, "IMGPROXY_COOKIE_PASSTHROUGH")
	configurators.String(&CookieBaseURL, "IMGPROXY_COOKIE_BASE_URL")
	configurators.Bool(&CookiePassthroughAll, "IMGPROXY_COOKIE_PASSTHROUGH_ALL")

	// Can't rely on configurators.String here because it ignores empty values
	if s, ok := os.LookupEnv("IMGPROXY_SOURCE_URL_QUERY_SEPARATOR"); ok {
		SourceURLQuerySeparator = s
	}

	configurators.String(&LocalFileSystemRoot, "IMGPROXY_LOCAL_FILESYSTEM_ROOT")

	configurators.Bool(&S3Enabled, "IMGPROXY_USE_S3")
	configurators.String(&S3Region, "IMGPROXY_S3_REGION")
	configurators.String(&S3Endpoint, "IMGPROXY_S3_ENDPOINT")
	configurators.Bool(&S3EndpointUsePathStyle, "IMGPROXY_S3_ENDPOINT_USE_PATH_STYLE")
	configurators.String(&S3AssumeRoleArn, "IMGPROXY_S3_ASSUME_ROLE_ARN")
	configurators.String(&S3AssumeRoleExternalID, "IMGPROXY_S3_ASSUME_ROLE_EXTERNAL_ID")
	configurators.Bool(&S3DecryptionClientEnabled, "IMGPROXY_S3_USE_DECRYPTION_CLIENT")

	configurators.Bool(&GCSEnabled, "IMGPROXY_USE_GCS")
	configurators.String(&GCSKey, "IMGPROXY_GCS_KEY")
	configurators.String(&GCSEndpoint, "IMGPROXY_GCS_ENDPOINT")

	configurators.Bool(&ABSEnabled, "IMGPROXY_USE_ABS")
	configurators.String(&ABSName, "IMGPROXY_ABS_NAME")
	configurators.String(&ABSKey, "IMGPROXY_ABS_KEY")
	configurators.String(&ABSEndpoint, "IMGPROXY_ABS_ENDPOINT")

	configurators.Bool(&SwiftEnabled, "IMGPROXY_USE_SWIFT")
	configurators.String(&SwiftUsername, "IMGPROXY_SWIFT_USERNAME")
	configurators.String(&SwiftAPIKey, "IMGPROXY_SWIFT_API_KEY")
	configurators.String(&SwiftAuthURL, "IMGPROXY_SWIFT_AUTH_URL")
	configurators.String(&SwiftDomain, "IMGPROXY_SWIFT_DOMAIN")
	configurators.String(&SwiftTenant, "IMGPROXY_SWIFT_TENANT")
	configurators.Int(&SwiftConnectTimeoutSeconds, "IMGPROXY_SWIFT_CONNECT_TIMEOUT_SECONDS")
	configurators.Int(&SwiftTimeoutSeconds, "IMGPROXY_SWIFT_TIMEOUT_SECONDS")

	configurators.Bool(&ETagEnabled, "IMGPROXY_USE_ETAG")
	configurators.String(&ETagBuster, "IMGPROXY_ETAG_BUSTER")

	configurators.Bool(&LastModifiedEnabled, "IMGPROXY_USE_LAST_MODIFIED")

	configurators.String(&BaseURL, "IMGPROXY_BASE_URL")
	if err := configurators.Replacements(&URLReplacements, "IMGPROXY_URL_REPLACEMENTS"); err != nil {
		return err
	}
	configurators.Bool(&Base64URLIncludesFilename, "IMGPROXY_BASE64_URL_INCLUDES_FILENAME")

	presetsSep := ","
	configurators.String(&presetsSep, "IMGPROXY_PRESETS_SEPARATOR")
	configurators.StringSliceSep(&Presets, "IMGPROXY_PRESETS", presetsSep)
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
	configurators.Int(&FallbackImageTTL, "IMGPROXY_FALLBACK_IMAGE_TTL")

	configurators.Bool(&DataDogEnable, "IMGPROXY_DATADOG_ENABLE")
	configurators.Bool(&DataDogEnableMetrics, "IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS")

	configurators.String(&NewRelicAppName, "IMGPROXY_NEW_RELIC_APP_NAME")
	configurators.String(&NewRelicKey, "IMGPROXY_NEW_RELIC_KEY")
	configurators.StringMap(&NewRelicLabels, "IMGPROXY_NEW_RELIC_LABELS")

	configurators.String(&PrometheusBind, "IMGPROXY_PROMETHEUS_BIND")
	configurators.String(&PrometheusNamespace, "IMGPROXY_PROMETHEUS_NAMESPACE")

	configurators.Bool(&OpenTelemetryEnable, "IMGPROXY_OPEN_TELEMETRY_ENABLE")
	configurators.Bool(&OpenTelemetryEnableMetrics, "IMGPROXY_OPEN_TELEMETRY_ENABLE_METRICS")
	configurators.String(&OpenTelemetryServerCert, "IMGPROXY_OPEN_TELEMETRY_SERVER_CERT")
	configurators.String(&OpenTelemetryClientCert, "IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT")
	configurators.String(&OpenTelemetryClientKey, "IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY")
	configurators.String(&OpenTelemetryTraceIDGenerator, "IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR")

	configurators.String(&CloudWatchServiceName, "IMGPROXY_CLOUD_WATCH_SERVICE_NAME")
	configurators.String(&CloudWatchNamespace, "IMGPROXY_CLOUD_WATCH_NAMESPACE")
	configurators.String(&CloudWatchRegion, "IMGPROXY_CLOUD_WATCH_REGION")

	configurators.String(&BugsnagKey, "IMGPROXY_BUGSNAG_KEY")
	configurators.String(&BugsnagStage, "IMGPROXY_BUGSNAG_STAGE")
	configurators.String(&HoneybadgerKey, "IMGPROXY_HONEYBADGER_KEY")
	configurators.String(&HoneybadgerEnv, "IMGPROXY_HONEYBADGER_ENV")
	configurators.String(&SentryDSN, "IMGPROXY_SENTRY_DSN")
	configurators.String(&SentryEnvironment, "IMGPROXY_SENTRY_ENVIRONMENT")
	configurators.String(&SentryRelease, "IMGPROXY_SENTRY_RELEASE")
	configurators.Int(&AirbrakeProjectID, "IMGPROXY_AIRBRAKE_PROJECT_ID")
	configurators.String(&AirbrakeProjectKey, "IMGPROXY_AIRBRAKE_PROJECT_KEY")
	configurators.String(&AirbrakeEnv, "IMGPROXY_AIRBRAKE_ENVIRONMENT")
	configurators.Bool(&ReportDownloadingErrors, "IMGPROXY_REPORT_DOWNLOADING_ERRORS")
	configurators.Bool(&ReportIOErrors, "IMGPROXY_REPORT_IO_ERRORS")
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
		return errors.New("Bind address is not defined")
	}

	if Timeout <= 0 {
		return fmt.Errorf("Timeout should be greater than 0, now - %d\n", Timeout)
	}
	if ReadRequestTimeout <= 0 {
		return fmt.Errorf("Read request timeout should be greater than 0, now - %d\n", ReadRequestTimeout)
	}
	if WriteResponseTimeout <= 0 {
		return fmt.Errorf("Write response timeout should be greater than 0, now - %d\n", WriteResponseTimeout)
	}
	if KeepAliveTimeout < 0 {
		return fmt.Errorf("KeepAlive timeout should be greater than or equal to 0, now - %d\n", KeepAliveTimeout)
	}
	if ClientKeepAliveTimeout < 0 {
		return fmt.Errorf("Client KeepAlive timeout should be greater than or equal to 0, now - %d\n", ClientKeepAliveTimeout)
	}

	if DownloadTimeout <= 0 {
		return fmt.Errorf("Download timeout should be greater than 0, now - %d\n", DownloadTimeout)
	}

	if Workers <= 0 {
		return fmt.Errorf("Workers number should be greater than 0, now - %d\n", Workers)
	}

	if RequestsQueueSize < 0 {
		return fmt.Errorf("Requests queue size should be greater than or equal 0, now - %d\n", RequestsQueueSize)
	}

	if MaxClients < 0 {
		return fmt.Errorf("Max clients number should be greater than or equal 0, now - %d\n", MaxClients)
	}

	if TTL < 0 {
		return fmt.Errorf("TTL should be greater than or equal to 0, now - %d\n", TTL)
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

	if AvifSpeed < 0 {
		return fmt.Errorf("Avif speed should be greater than or equal to 0, now - %d\n", AvifSpeed)
	} else if AvifSpeed > 9 {
		return fmt.Errorf("Avif speed can't be greater than 9, now - %d\n", AvifSpeed)
	}

	if JxlEffort < 1 {
		return fmt.Errorf("JXL effort should be greater than 0, now - %d\n", JxlEffort)
	} else if JxlEffort > 9 {
		return fmt.Errorf("JXL effort can't be greater than 9, now - %d\n", JxlEffort)
	}

	if WebpEffort < 1 {
		return fmt.Errorf("Webp effort should be greater than 0, now - %d\n", WebpEffort)
	} else if WebpEffort > 6 {
		return fmt.Errorf("Webp effort can't be greater than 6, now - %d\n", WebpEffort)
	}

	if Quality <= 0 {
		return fmt.Errorf("Quality should be greater than 0, now - %d\n", Quality)
	} else if Quality > 100 {
		return fmt.Errorf("Quality can't be greater than 100, now - %d\n", Quality)
	}

	if len(PreferredFormats) == 0 {
		return errors.New("At least one preferred format should be specified")
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
			return errors.New("Cannot use local directory: not a directory")
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
		return errors.New("Watermark opacity should be greater than 0")
	} else if WatermarkOpacity > 1 {
		return errors.New("Watermark opacity should be less than or equal to 1")
	}

	if FallbackImageTTL < 0 {
		return fmt.Errorf("Fallback image TTL should be greater than or equal to 0, now - %d\n", TTL)
	}

	if FallbackImageHTTPCode != 0 && (FallbackImageHTTPCode < 100 || FallbackImageHTTPCode > 599) {
		return errors.New("Fallback image HTTP code should be between 100 and 599")
	}

	if len(PrometheusBind) > 0 && PrometheusBind == Bind {
		return errors.New("Can't use the same binding for the main server and Prometheus")
	}

	if FreeMemoryInterval <= 0 {
		return errors.New("Free memory interval should be greater than zero")
	}

	if DownloadBufferSize < 0 {
		return errors.New("Download buffer size should be greater than or equal to 0")
	} else if DownloadBufferSize > math.MaxInt32 {
		return fmt.Errorf("Download buffer size can't be greater than %d", math.MaxInt32)
	}

	if BufferPoolCalibrationThreshold < 64 {
		return errors.New("Buffer pool calibration threshold should be greater than or equal to 64")
	}

	return nil
}
