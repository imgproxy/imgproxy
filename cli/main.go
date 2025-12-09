package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/imgproxy/imgproxy/v3"
	"github.com/imgproxy/imgproxy/v3/logger"
	optionsparser "github.com/imgproxy/imgproxy/v3/options/parser"
	"github.com/imgproxy/imgproxy/v3/version"
	"github.com/urfave/cli/v3"
)

// ver prints the imgproxy version and runs the main application
func ver(ctx context.Context, c *cli.Command) error {
	//nolint:forbidigo
	fmt.Println(version.Version)
	return nil
}

// run starts the imgproxy server
func run(ctx context.Context, cmd *cli.Command) error {
	if err := imgproxy.Init(ctx); err != nil {
		return err
	}
	defer imgproxy.Shutdown()

	cfg, err := imgproxy.LoadConfigFromEnv(nil)
	if err != nil {
		return err
	}

	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	instance, err := imgproxy.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer instance.Close(ctx)

	if err := instance.StartServer(ctx, nil); err != nil {
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
				Name:  optionsparser.PresetsFlagName,
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
		logger.Fatal(err.Error())
	}
}
