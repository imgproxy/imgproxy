package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
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

func hexEnvConfig(b *[]securityKey, name string) {
	var err error

	if env := os.Getenv(name); len(env) > 0 {
		parts := strings.Split(env, ",")

		keys := make([]securityKey, len(parts))

		for i, part := range parts {
			if keys[i], err = hex.DecodeString(part); err != nil {
				logFatal("%s expected to be hex-encoded strings. Invalid: %s\n", name, part)
			}
		}

		*b = keys
	}
}

func hexFileConfig(b *[]securityKey, filepath string) {
	if len(filepath) == 0 {
		return
	}

	f, err := os.Open(filepath)
	if err != nil {
		logFatal("Can't open file %s\n", filepath)
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
			logFatal("%s expected to contain hex-encoded strings. Invalid: %s\n", filepath, part)
		}
	}

	if err := scanner.Err(); err != nil {
		logFatal("Failed to read file %s: %s", filepath, err)
	}

	*b = keys
}

func presetEnvConfig(p presets, name string) {
	if env := os.Getenv(name); len(env) > 0 {
		presetStrings := strings.Split(env, ",")

		for _, presetStr := range presetStrings {
			if err := parsePreset(p, presetStr); err != nil {
				logFatal(err.Error())
			}
		}
	}
}

func presetFileConfig(p presets, filepath string) {
	if len(filepath) == 0 {
		return
	}

	f, err := os.Open(filepath)
	if err != nil {
		logFatal("Can't open file %s\n", filepath)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := parsePreset(p, scanner.Text()); err != nil {
			logFatal(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		logFatal("Failed to read presets file: %s", err)
	}
}

type config struct {
	Bind             string
	ReadTimeout      int
	WriteTimeout     int
	KeepAliveTimeout int
	DownloadTimeout  int
	Concurrency      int
	MaxClients       int
	TTL              int
	SoReuseport      bool

	MaxSrcDimension    int
	MaxSrcResolution   int
	MaxSrcFileSize     int
	MaxAnimationFrames int

	JpegProgressive       bool
	PngInterlaced         bool
	PngQuantize           bool
	PngQuantizationColors int
	Quality               int
	GZipCompression       int

	EnableWebpDetection bool
	EnforceWebp         bool
	EnableClientHints   bool

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

	LocalFileSystemRoot string
	S3Enabled           bool
	S3Region            string
	S3Endpoint          string
	GCSKey              string

	ETagEnabled bool

	BaseURL string

	Presets     presets
	OnlyPresets bool

	WatermarkData    string
	WatermarkPath    string
	WatermarkURL     string
	WatermarkOpacity float64

	NewRelicAppName string
	NewRelicKey     string

	PrometheusBind string

	BugsnagKey        string
	BugsnagStage      string
	HoneybadgerKey    string
	HoneybadgerEnv    string
	SentryDSN         string
	SentryEnvironment string
	SentryRelease     string

	FreeMemoryInterval             int
	DownloadBufferSize             int
	GZipBufferSize                 int
	BufferPoolCalibrationThreshold int
}

var conf = config{
	Bind:                           ":8080",
	ReadTimeout:                    10,
	WriteTimeout:                   10,
	KeepAliveTimeout:               10,
	DownloadTimeout:                5,
	Concurrency:                    runtime.NumCPU() * 2,
	TTL:                            3600,
	MaxSrcResolution:               16800000,
	MaxAnimationFrames:             1,
	SignatureSize:                  32,
	PngQuantizationColors:          256,
	Quality:                        80,
	GZipCompression:                5,
	UserAgent:                      fmt.Sprintf("imgproxy/%s", version),
	Presets:                        make(presets),
	WatermarkOpacity:               1,
	BugsnagStage:                   "production",
	HoneybadgerEnv:                 "production",
	SentryEnvironment:              "production",
	SentryRelease:                  fmt.Sprintf("imgproxy/%s", version),
	FreeMemoryInterval:             10,
	BufferPoolCalibrationThreshold: 1024,
}

func configure() {
	keyPath := flag.String("keypath", "", "path of the file with hex-encoded key")
	saltPath := flag.String("saltpath", "", "path of the file with hex-encoded salt")
	presetsPath := flag.String("presets", "", "path of the file with presets")
	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if port := os.Getenv("PORT"); len(port) > 0 {
		conf.Bind = fmt.Sprintf(":%s", port)
	}

	strEnvConfig(&conf.Bind, "IMGPROXY_BIND")
	intEnvConfig(&conf.ReadTimeout, "IMGPROXY_READ_TIMEOUT")
	intEnvConfig(&conf.WriteTimeout, "IMGPROXY_WRITE_TIMEOUT")
	intEnvConfig(&conf.KeepAliveTimeout, "IMGPROXY_KEEP_ALIVE_TIMEOUT")
	intEnvConfig(&conf.DownloadTimeout, "IMGPROXY_DOWNLOAD_TIMEOUT")
	intEnvConfig(&conf.Concurrency, "IMGPROXY_CONCURRENCY")
	intEnvConfig(&conf.MaxClients, "IMGPROXY_MAX_CLIENTS")

	intEnvConfig(&conf.TTL, "IMGPROXY_TTL")

	boolEnvConfig(&conf.SoReuseport, "IMGPROXY_SO_REUSEPORT")

	intEnvConfig(&conf.MaxSrcDimension, "IMGPROXY_MAX_SRC_DIMENSION")
	megaIntEnvConfig(&conf.MaxSrcResolution, "IMGPROXY_MAX_SRC_RESOLUTION")
	intEnvConfig(&conf.MaxSrcFileSize, "IMGPROXY_MAX_SRC_FILE_SIZE")

	if _, ok := os.LookupEnv("IMGPROXY_MAX_GIF_FRAMES"); ok {
		logWarning("`IMGPROXY_MAX_GIF_FRAMES` is deprecated and will be removed in future versions. Use `IMGPROXY_MAX_ANIMATION_FRAMES` instead")
		intEnvConfig(&conf.MaxAnimationFrames, "IMGPROXY_MAX_GIF_FRAMES")
	}
	intEnvConfig(&conf.MaxAnimationFrames, "IMGPROXY_MAX_ANIMATION_FRAMES")

	boolEnvConfig(&conf.JpegProgressive, "IMGPROXY_JPEG_PROGRESSIVE")
	boolEnvConfig(&conf.PngInterlaced, "IMGPROXY_PNG_INTERLACED")
	boolEnvConfig(&conf.PngQuantize, "IMGPROXY_PNG_QUANTIZE")
	intEnvConfig(&conf.PngQuantizationColors, "IMGPROXY_PNG_QUANTIZATION_COLORS")
	intEnvConfig(&conf.Quality, "IMGPROXY_QUALITY")
	intEnvConfig(&conf.GZipCompression, "IMGPROXY_GZIP_COMPRESSION")

	boolEnvConfig(&conf.EnableWebpDetection, "IMGPROXY_ENABLE_WEBP_DETECTION")
	boolEnvConfig(&conf.EnforceWebp, "IMGPROXY_ENFORCE_WEBP")
	boolEnvConfig(&conf.EnableClientHints, "IMGPROXY_ENABLE_CLIENT_HINTS")

	boolEnvConfig(&conf.UseLinearColorspace, "IMGPROXY_USE_LINEAR_COLORSPACE")
	boolEnvConfig(&conf.DisableShrinkOnLoad, "IMGPROXY_DISABLE_SHRINK_ON_LOAD")

	hexEnvConfig(&conf.Keys, "IMGPROXY_KEY")
	hexEnvConfig(&conf.Salts, "IMGPROXY_SALT")
	intEnvConfig(&conf.SignatureSize, "IMGPROXY_SIGNATURE_SIZE")

	hexFileConfig(&conf.Keys, *keyPath)
	hexFileConfig(&conf.Salts, *saltPath)

	strEnvConfig(&conf.Secret, "IMGPROXY_SECRET")

	strEnvConfig(&conf.AllowOrigin, "IMGPROXY_ALLOW_ORIGIN")

	strEnvConfig(&conf.UserAgent, "IMGPROXY_USER_AGENT")

	boolEnvConfig(&conf.IgnoreSslVerification, "IMGPROXY_IGNORE_SSL_VERIFICATION")
	boolEnvConfig(&conf.DevelopmentErrorsMode, "IMGPROXY_DEVELOPMENT_ERRORS_MODE")

	strEnvConfig(&conf.LocalFileSystemRoot, "IMGPROXY_LOCAL_FILESYSTEM_ROOT")

	boolEnvConfig(&conf.S3Enabled, "IMGPROXY_USE_S3")
	strEnvConfig(&conf.S3Region, "IMGPROXY_S3_REGION")
	strEnvConfig(&conf.S3Endpoint, "IMGPROXY_S3_ENDPOINT")

	strEnvConfig(&conf.GCSKey, "IMGPROXY_GCS_KEY")

	boolEnvConfig(&conf.ETagEnabled, "IMGPROXY_USE_ETAG")

	strEnvConfig(&conf.BaseURL, "IMGPROXY_BASE_URL")

	presetEnvConfig(conf.Presets, "IMGPROXY_PRESETS")
	presetFileConfig(conf.Presets, *presetsPath)
	boolEnvConfig(&conf.OnlyPresets, "IMGPROXY_ONLY_PRESETS")

	strEnvConfig(&conf.WatermarkData, "IMGPROXY_WATERMARK_DATA")
	strEnvConfig(&conf.WatermarkPath, "IMGPROXY_WATERMARK_PATH")
	strEnvConfig(&conf.WatermarkURL, "IMGPROXY_WATERMARK_URL")
	floatEnvConfig(&conf.WatermarkOpacity, "IMGPROXY_WATERMARK_OPACITY")

	strEnvConfig(&conf.NewRelicAppName, "IMGPROXY_NEW_RELIC_APP_NAME")
	strEnvConfig(&conf.NewRelicKey, "IMGPROXY_NEW_RELIC_KEY")

	strEnvConfig(&conf.PrometheusBind, "IMGPROXY_PROMETHEUS_BIND")

	strEnvConfig(&conf.BugsnagKey, "IMGPROXY_BUGSNAG_KEY")
	strEnvConfig(&conf.BugsnagStage, "IMGPROXY_BUGSNAG_STAGE")
	strEnvConfig(&conf.HoneybadgerKey, "IMGPROXY_HONEYBADGER_KEY")
	strEnvConfig(&conf.HoneybadgerEnv, "IMGPROXY_HONEYBADGER_ENV")
	strEnvConfig(&conf.SentryDSN, "IMGPROXY_SENTRY_DSN")
	strEnvConfig(&conf.SentryEnvironment, "IMGPROXY_SENTRY_ENVIRONMENT")
	strEnvConfig(&conf.SentryRelease, "IMGPROXY_SENTRY_RELEASE")

	intEnvConfig(&conf.FreeMemoryInterval, "IMGPROXY_FREE_MEMORY_INTERVAL")
	intEnvConfig(&conf.DownloadBufferSize, "IMGPROXY_DOWNLOAD_BUFFER_SIZE")
	intEnvConfig(&conf.GZipBufferSize, "IMGPROXY_GZIP_BUFFER_SIZE")
	intEnvConfig(&conf.BufferPoolCalibrationThreshold, "IMGPROXY_BUFFER_POOL_CALIBRATION_THRESHOLD")

	if len(conf.Keys) != len(conf.Salts) {
		logFatal("Number of keys and number of salts should be equal. Keys: %d, salts: %d", len(conf.Keys), len(conf.Salts))
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
		logFatal("Signature size should be within 1 and 32, now - %d\n", conf.SignatureSize)
	}

	if len(conf.Bind) == 0 {
		logFatal("Bind address is not defined")
	}

	if conf.ReadTimeout <= 0 {
		logFatal("Read timeout should be greater than 0, now - %d\n", conf.ReadTimeout)
	}

	if conf.WriteTimeout <= 0 {
		logFatal("Write timeout should be greater than 0, now - %d\n", conf.WriteTimeout)
	}
	if conf.KeepAliveTimeout < 0 {
		logFatal("KeepAlive timeout should be greater than or equal to 0, now - %d\n", conf.KeepAliveTimeout)
	}

	if conf.DownloadTimeout <= 0 {
		logFatal("Download timeout should be greater than 0, now - %d\n", conf.DownloadTimeout)
	}

	if conf.Concurrency <= 0 {
		logFatal("Concurrency should be greater than 0, now - %d\n", conf.Concurrency)
	}

	if conf.MaxClients <= 0 {
		conf.MaxClients = conf.Concurrency * 10
	}

	if conf.TTL <= 0 {
		logFatal("TTL should be greater than 0, now - %d\n", conf.TTL)
	}

	if conf.MaxSrcDimension < 0 {
		logFatal("Max src dimension should be greater than or equal to 0, now - %d\n", conf.MaxSrcDimension)
	} else if conf.MaxSrcDimension > 0 {
		logWarning("IMGPROXY_MAX_SRC_DIMENSION is deprecated and can be removed in future versions. Use IMGPROXY_MAX_SRC_RESOLUTION")
	}

	if conf.MaxSrcResolution <= 0 {
		logFatal("Max src resolution should be greater than 0, now - %d\n", conf.MaxSrcResolution)
	}

	if conf.MaxSrcFileSize < 0 {
		logFatal("Max src file size should be greater than or equal to 0, now - %d\n", conf.MaxSrcFileSize)
	}

	if conf.MaxAnimationFrames <= 0 {
		logFatal("Max animation frames should be greater than 0, now - %d\n", conf.MaxAnimationFrames)
	}

	if conf.PngQuantizationColors < 2 {
		logFatal("Png quantization colors should be greater than 1, now - %d\n", conf.PngQuantizationColors)
	} else if conf.PngQuantizationColors > 256 {
		logFatal("Png quantization colors can't be greater than 256, now - %d\n", conf.PngQuantizationColors)
	}

	if conf.Quality <= 0 {
		logFatal("Quality should be greater than 0, now - %d\n", conf.Quality)
	} else if conf.Quality > 100 {
		logFatal("Quality can't be greater than 100, now - %d\n", conf.Quality)
	}

	if conf.GZipCompression < 0 {
		logFatal("GZip compression should be greater than or equal to 0, now - %d\n", conf.GZipCompression)
	} else if conf.GZipCompression > 9 {
		logFatal("GZip compression can't be greater than 9, now - %d\n", conf.GZipCompression)
	}

	if conf.IgnoreSslVerification {
		logWarning("Ignoring SSL verification is very unsafe")
	}

	if conf.LocalFileSystemRoot != "" {
		stat, err := os.Stat(conf.LocalFileSystemRoot)
		if err != nil {
			logFatal("Cannot use local directory: %s", err)
		} else {
			if !stat.IsDir() {
				logFatal("Cannot use local directory: not a directory")
			}
		}
		if conf.LocalFileSystemRoot == "/" {
			logNotice("Exposing root via IMGPROXY_LOCAL_FILESYSTEM_ROOT is unsafe")
		}
	}

	if err := checkPresets(conf.Presets); err != nil {
		logFatal(err.Error())
	}

	if conf.WatermarkOpacity <= 0 {
		logFatal("Watermark opacity should be greater than 0")
	} else if conf.WatermarkOpacity > 1 {
		logFatal("Watermark opacity should be less than or equal to 1")
	}

	if len(conf.PrometheusBind) > 0 && conf.PrometheusBind == conf.Bind {
		logFatal("Can't use the same binding for the main server and Prometheus")
	}

	if conf.FreeMemoryInterval <= 0 {
		logFatal("Free memory interval should be greater than zero")
	}

	if conf.DownloadBufferSize < 0 {
		logFatal("Download buffer size should be greater than or equal to 0")
	} else if conf.DownloadBufferSize > int(^uint32(0)) {
		logFatal("Download buffer size can't be greater than %d", ^uint32(0))
	}

	if conf.GZipBufferSize < 0 {
		logFatal("GZip buffer size should be greater than or equal to 0")
	} else if conf.GZipBufferSize > int(^uint32(0)) {
		logFatal("GZip buffer size can't be greater than %d", ^uint32(0))
	}

	if conf.BufferPoolCalibrationThreshold < 64 {
		logFatal("Buffer pool calibration threshold should be greater than or equal to 64")
	}
}
