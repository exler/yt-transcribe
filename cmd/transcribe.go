package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/ffmpeg"
	"github.com/urfave/cli/v3"
)

var (
	transcribeCmd = &cli.Command{
		Name:  "transcribe",
		Usage: "Transcribe a YouTube video",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "whisper-model-path",
				Usage: "Path to ggml whisper.cpp model file (e.g. /models/ggml-small.bin)",
				Value: "models/ggml-small.bin",
			},
			&cli.StringFlag{
				Name:  "language",
				Usage: "Language code to use or 'auto' to autodetect",
				Value: "auto",
			},
			&cli.IntFlag{
				Name:  "queue",
				Usage: "FFmpeg whisper filter queue size in seconds",
				Value: 15,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			videoURL := cmd.Args().First()
			if videoURL == "" {
				return cli.Exit("Please provide a YouTube video URL to transcribe", 1)
			}

			modelPath := cmd.String("whisper-model-path")
			if modelPath == "" {
				return cli.Exit("Path to whisper.cpp ggml model is required", 1)
			}
			language := cmd.String("language")
			queue := cmd.Int("queue")

			tempDir, err := os.MkdirTemp("", "yt-transcribe-*")
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to create temporary directory: %v", err), 1)
			}

			// Ensure cleanup of temp directory
			defer func() {
				if err := os.RemoveAll(tempDir); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup temporary files: %v\n", err)
				}
			}()

			fmt.Printf("Using temporary directory: %s\n", tempDir)

			downloader, err := fetch.NewYouTubeDownloader(tempDir)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to initialize YouTube downloader: %v", err), 1)
			}

			if err := downloader.CheckYTDLP(); err != nil {
				return cli.Exit(fmt.Sprintf("yt-dlp check failed: %v", err), 1)
			}

			fmt.Println("Downloading video...")
			downloadedMetadata, err := downloader.DownloadAudio(videoURL) // Renamed variable
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to download audio: %v", err), 1)
			}

			ff, err := ffmpeg.NewFFMPEG()
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to initialize ffmpeg: %v", err), 1)
			}

			// No speed-up: use original downloaded file for whisper filter
			fmt.Println("Transcribing audio with FFmpeg whisper filter...")
			transcriptionText, err := ff.TranscribeWithWhisperFilter(downloadedMetadata.AudioFilePath, modelPath, language, queue)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to transcribe audio with whisper filter: %v", err), 1)
			}

			fmt.Println(transcriptionText)

			return nil
		},
	}
)
