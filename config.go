package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
)

func intEnvConfig(i *int, name string) {
	if env, err := strconv.Atoi(os.Getenv(name)); err == nil {
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
	*b = false
	if env, err := strconv.ParseBool(os.Getenv(name)); err == nil {
		*b = env
	}
}

func hexEnvConfig(b *[]byte, name string) {
	var err error

	if env := os.Getenv(name); len(env) > 0 {
		if *b, err = hex.DecodeString(env); err != nil {
			log.Fatalf("%s expected to be hex-encoded string\n", name)
		}
	}
}

func hexFileConfig(b *[]byte, filepath string) {
	if len(filepath) == 0 {
		return
	}

	f, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Can't open file %s\n", filepath)
	}

	src, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}

	src = bytes.TrimSpace(src)

	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		log.Fatalf("%s expected to contain hex-encoded string\n", filepath)
	}

	*b = dst[:n]
}

type config struct {
	Bind            string
	ReadTimeout     int
	WaitTimeout     int
	WriteTimeout    int
	DownloadTimeout int
	Concurrency     int
	MaxClients      int
	TTL             int

	MaxSrcDimension  int
	MaxSrcResolution int

	Quality         int
	GZipCompression int

	Key  []byte
	Salt []byte

	Secret string

	AllowOrigin string

	LocalFileSystemRoot string

	ETagEnabled   bool
	ETagSignature []byte

	BaseURL string
}

var conf = config{
	Bind:             ":8080",
	ReadTimeout:      10,
	WriteTimeout:     10,
	DownloadTimeout:  5,
	Concurrency:      runtime.NumCPU() * 2,
	TTL:              3600,
	MaxSrcDimension:  8192,
	MaxSrcResolution: 16800000,
	Quality:          80,
	GZipCompression:  5,
	ETagEnabled:      false,
}

func init() {
	keypath := flag.String("keypath", "", "path of the file with hex-encoded key")
	saltpath := flag.String("saltpath", "", "path of the file with hex-encoded salt")
	flag.Parse()

	if port := os.Getenv("PORT"); len(port) > 0 {
		conf.Bind = fmt.Sprintf(":%s", port)
	}

	strEnvConfig(&conf.Bind, "IMGPROXY_BIND")
	intEnvConfig(&conf.ReadTimeout, "IMGPROXY_READ_TIMEOUT")
	intEnvConfig(&conf.WriteTimeout, "IMGPROXY_WRITE_TIMEOUT")
	intEnvConfig(&conf.DownloadTimeout, "IMGPROXY_DOWNLOAD_TIMEOUT")
	intEnvConfig(&conf.Concurrency, "IMGPROXY_CONCURRENCY")
	intEnvConfig(&conf.MaxClients, "IMGPROXY_MAX_CLIENTS")

	intEnvConfig(&conf.TTL, "IMGPROXY_TTL")

	intEnvConfig(&conf.MaxSrcDimension, "IMGPROXY_MAX_SRC_DIMENSION")
	megaIntEnvConfig(&conf.MaxSrcResolution, "IMGPROXY_MAX_SRC_RESOLUTION")

	intEnvConfig(&conf.Quality, "IMGPROXY_QUALITY")
	intEnvConfig(&conf.GZipCompression, "IMGPROXY_GZIP_COMPRESSION")

	hexEnvConfig(&conf.Key, "IMGPROXY_KEY")
	hexEnvConfig(&conf.Salt, "IMGPROXY_SALT")

	hexFileConfig(&conf.Key, *keypath)
	hexFileConfig(&conf.Salt, *saltpath)

	strEnvConfig(&conf.Secret, "IMGPROXY_SECRET")

	strEnvConfig(&conf.AllowOrigin, "IMGPROXY_ALLOW_ORIGIN")

	strEnvConfig(&conf.LocalFileSystemRoot, "IMGPROXY_LOCAL_FILESYSTEM_ROOT")

	boolEnvConfig(&conf.ETagEnabled, "IMGPROXY_USE_ETAG")

	strEnvConfig(&conf.BaseURL, "IMGPROXY_BASE_URL")

	if len(conf.Key) == 0 {
		log.Fatalln("Key is not defined")
	}
	if len(conf.Salt) == 0 {
		log.Fatalln("Salt is not defined")
	}

	if len(conf.Bind) == 0 {
		log.Fatalln("Bind address is not defined")
	}

	if conf.ReadTimeout <= 0 {
		log.Fatalf("Read timeout should be greater than 0, now - %d\n", conf.ReadTimeout)
	}

	if conf.WriteTimeout <= 0 {
		log.Fatalf("Write timeout should be greater than 0, now - %d\n", conf.WriteTimeout)
	}

	if conf.DownloadTimeout <= 0 {
		log.Fatalf("Download timeout should be greater than 0, now - %d\n", conf.DownloadTimeout)
	}

	if conf.Concurrency <= 0 {
		log.Fatalf("Concurrency should be greater than 0, now - %d\n", conf.Concurrency)
	}

	if conf.MaxClients <= 0 {
		conf.MaxClients = conf.Concurrency * 10
	}

	if conf.TTL <= 0 {
		log.Fatalf("TTL should be greater than 0, now - %d\n", conf.TTL)
	}

	if conf.MaxSrcDimension <= 0 {
		log.Fatalf("Max src dimension should be greater than 0, now - %d\n", conf.MaxSrcDimension)
	}

	if conf.MaxSrcResolution <= 0 {
		log.Fatalf("Max src resolution should be greater than 0, now - %d\n", conf.MaxSrcResolution)
	}

	if conf.Quality <= 0 {
		log.Fatalf("Quality should be greater than 0, now - %d\n", conf.Quality)
	} else if conf.Quality > 100 {
		log.Fatalf("Quality can't be greater than 100, now - %d\n", conf.Quality)
	}

	if conf.GZipCompression < 0 {
		log.Fatalf("GZip compression should be greater than or quual to 0, now - %d\n", conf.GZipCompression)
	} else if conf.GZipCompression > 9 {
		log.Fatalf("GZip compression can't be greater than 9, now - %d\n", conf.GZipCompression)
	}

	if conf.LocalFileSystemRoot != "" {
		stat, err := os.Stat(conf.LocalFileSystemRoot)
		if err != nil {
			log.Fatalf("Cannot use local directory: %s", err)
		} else {
			if !stat.IsDir() {
				log.Fatalf("Cannot use local directory: not a directory")
			}
		}
		if conf.LocalFileSystemRoot == "/" {
			log.Print("Exposing root via IMGPROXY_LOCAL_FILESYSTEM_ROOT is unsafe")
		}
	}

	if conf.ETagEnabled {
		conf.ETagSignature = make([]byte, 16)
		rand.Read(conf.ETagSignature)
		log.Printf("ETag support is activated. The random value was generated to be used for ETag calculation: %s\n",
			fmt.Sprintf("%x", conf.ETagSignature))
	}

	initVips()
	initDownloading()
}
