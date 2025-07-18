package cmd

import (
	"context"
	"fmt"

	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/ffmpeg"
	"github.com/urfave/cli/v3"
)

var (
	Version = "development"

	versionCmd = &cli.Command{
		Name:  "version",
		Usage: "Show current version",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show versions of ffmpeg and yt-dlp dependencies",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			fmt.Printf("yt-transcribe %s\n", Version)

			if cmd.Bool("verbose") {
				ffmpegProcessor, err := ffmpeg.NewFFMPEG()
				if err != nil {
					fmt.Printf("ffmpeg initialization error: %v\n", err)
				} else if ffmpegVersion, err := ffmpegProcessor.GetFFMPEGVersion(); err == nil {
					fmt.Printf("ffmpeg %s\n", ffmpegVersion)
				} else {
					fmt.Printf("ffmpeg not found or error: %v\n", err)
				}

				downloader, err := fetch.NewYouTubeDownloader("")
				if err != nil {
					fmt.Printf("yt-dlp initialization error: %v\n", err)
				} else if ytDlpVersion, err := downloader.GetYTDLPVersion(); err == nil {
					fmt.Printf("yt-dlp %s\n", ytDlpVersion)
				} else {
					fmt.Printf("yt-dlp not found or error: %v\n", err)
				}
			}

			return nil
		},
	}
)
