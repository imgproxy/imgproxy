package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/imgproxy/imgproxy/v3"
	"github.com/imgproxy/imgproxy/v3/version"
	"github.com/urfave/cli/v3"
)

// ver prints the imgproxy version and runs the main application
func ver(ctx context.Context, c *cli.Command) error {
	fmt.Println(version.Version)
	return nil
}

// run starts the imgproxy server
func run(ctx context.Context, cmd *cli.Command) error {
	// NOTE: for now, these flags are loaded in config.go package

	// keypath := cmd.String("keypath")
	// saltpath := cmd.String("saltpath")
	// presets := cmd.String("presets")

	if err := imgproxy.Init(); err != nil {
		return err
	}
	defer imgproxy.Shutdown()

	cfg, err := imgproxy.LoadConfigFromEnv(nil)
	if err != nil {
		return err
	}

	instance, err := imgproxy.New(ctx, cfg)
	if err != nil {
		return err
	}

	if err := instance.StartServer(ctx); err != nil {
		return err
	}

	return nil
}

func main() {
	cmd := &cli.Command{
		Name:  "imgproxy",
		Usage: "Fast and secure standalone server for resizing and converting remote images",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "keypath",
				Usage: "path of the file with hex-encoded key",
			},
			&cli.StringFlag{
				Name:  "saltpath",
				Usage: "path of the file with hex-encoded salt",
			},
			&cli.StringFlag{
				Name:  "presets",
				Usage: "path of the file with presets",
			},
		},
		Action: run,
		Commands: []*cli.Command{
			{
				Name:   "version",
				Usage:  "print the version",
				Action: ver,
			},
			{
				Name:   "health",
				Usage:  "perform a healthcheck on a running imgproxy instance",
				Action: healthcheck,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
