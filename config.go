package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func absPathToFile(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	appPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}

	return filepath.Join(appPath, path)
}

func intEnvConfig(i *int, name string) {
	if env, err := strconv.Atoi(os.Getenv(name)); err == nil {
		*i = env
	}
}

func strEnvConfig(s *string, name string) {
	if env := os.Getenv(name); len(env) > 0 {
		*s = env
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

	fullfp := absPathToFile(filepath)
	f, err := os.Open(fullfp)
	if err != nil {
		log.Fatalf("Can't open file %s\n", fullfp)
	}

	src, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}

	src = bytes.TrimSpace(src)

	dst := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(dst, src)
	if err != nil {
		log.Fatalf("%s expected to contain hex-encoded string\n", fullfp)
	}

	*b = dst[:n]
}

type config struct {
	Bind         string
	ReadTimeout  int
	WriteTimeout int

	MaxSrcDimension int

	Quality     int
	Compression int

	Key  []byte
	Salt []byte
}

var conf = config{
	Bind:            ":8080",
	ReadTimeout:     10,
	WriteTimeout:    10,
	MaxSrcDimension: 4096,
	Quality:         80,
	Compression:     6,
}

func init() {
	keypath := flag.String("keypath", "", "path of the file with hex-encoded key")
	saltpath := flag.String("saltpath", "", "path of the file with hex-encoded salt")
	flag.Parse()

	strEnvConfig(&conf.Bind, "IMGPROXY_BIND")
	intEnvConfig(&conf.ReadTimeout, "IMGPROXY_READ_TIMEOUT")
	intEnvConfig(&conf.WriteTimeout, "IMGPROXY_WRITE_TIMEOUT")

	intEnvConfig(&conf.MaxSrcDimension, "IMGPROXY_MAX_SRC_DIMENSION")

	intEnvConfig(&conf.Quality, "IMGPROXY_QUALITY")
	intEnvConfig(&conf.Compression, "IMGPROXY_COMPRESSION")

	hexEnvConfig(&conf.Key, "IMGPROXY_KEY")
	hexEnvConfig(&conf.Salt, "IMGPROXY_SALT")

	hexFileConfig(&conf.Key, *keypath)
	hexFileConfig(&conf.Salt, *saltpath)

	if len(conf.Key) == 0 {
		log.Fatalln("Key is not defined")
	}
	if len(conf.Salt) == 0 {
		log.Fatalln("Salt is not defined")
	}
}
