package vips_test

import (
	"os"
	"testing"

	"github.com/imgproxy/imgproxy/v4/vips"
)

func TestMain(m *testing.M) {
	cfg := vips.NewDefaultConfig()
	err := vips.Init(&cfg)
	if err != nil {
		panic(err)
	}

	r := m.Run()

	vips.Shutdown()

	os.Exit(r)
}
