package cmd

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
)

var cmd = &cli.Command{
	Name:     "yt-transcribe",
	Usage:    "Transcribe YouTube videos using AI speech recognition",
	Commands: []*cli.Command{versionCmd, transcribeCmd, runserverCmd},
}

func Run() error {
	return cmd.Run(context.Background(), os.Args)
}
