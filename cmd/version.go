package cmd

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

var (
	Version = "development"

	versionCmd = &cli.Command{
		Name:  "version",
		Usage: "Show current version",
		Action: func(context.Context, *cli.Command) error {
			fmt.Printf("yt-transcribe %s\n", Version)
			return nil
		},
	}
)
