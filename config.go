package main

import (
	"encoding/hex"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type config struct {
	Bind         string
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`

	Key     string
	Salt    string
	KeyBin  []byte
	SaltBin []byte

	MaxSrcDimension int `yaml:"max_src_dimension"`

	Quality     int
	Compression int
}

var conf = config{
	Bind:            ":8080",
	MaxSrcDimension: 4096,
}

func absPathToFile(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	appPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return filepath.Join(appPath, path)
}

func init() {
	cpath := flag.String(
		"config", "./config.yml", "path to configuration file",
	)
	flag.Parse()

	file, err := os.Open(absPathToFile(*cpath))
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	cdata, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(cdata, &conf)
	if err != nil {
		log.Fatalln(err)
	}

	if len(conf.Bind) == 0 {
		conf.Bind = ":8080"
	}

	if conf.MaxSrcDimension == 0 {
		conf.MaxSrcDimension = 4096
	}

	if conf.KeyBin, err = hex.DecodeString(conf.Key); err != nil {
		log.Fatalln("Invalid key. Key should be encoded to hex")
	}

	if conf.SaltBin, err = hex.DecodeString(conf.Salt); err != nil {
		log.Fatalln("Invalid salt. Salt should be encoded to hex")
	}

	if conf.MaxSrcDimension == 0 {
		conf.MaxSrcDimension = 4096
	}

	if conf.Quality == 0 {
		conf.Quality = 80
	}

	if conf.Compression == 0 {
		conf.Compression = 6
	}
}
