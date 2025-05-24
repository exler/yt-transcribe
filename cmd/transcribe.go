package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/transcribe"
	"github.com/urfave/cli/v3"
)

var (
	transcribeCmd = &cli.Command{
		Name:  "transcribe",
		Usage: "Transcribe a YouTube video",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api-key",
				Aliases: []string{"k"},
				Usage:   "OpenAI API key (can also be set via OPENAI_API_KEY env variable)",
			},
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Usage:   "Whisper model to use for transcription",
				Value:   "whisper-1",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			videoURL := cmd.Args().First()
			if videoURL == "" {
				return fmt.Errorf("video URL is required")
			}

			// Check for API key in environment variable if not provided via flag
			apiKey := cmd.String("api-key")
			if apiKey == "" {
				apiKey = os.Getenv("OPENAI_API_KEY")
			}

			model := cmd.String("model")

			fmt.Printf("Downloading video: %s\n", videoURL)

			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "yt-transcribe-*")
			if err != nil {
				return fmt.Errorf("failed to create temporary directory: %w", err)
			}

			// Ensure cleanup of temp directory when we're done
			defer func() {
				if err := os.RemoveAll(tempDir); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to cleanup temporary files: %v\n", err)
				}
			}()

			fmt.Printf("Using temporary directory: %s\n", tempDir)

			// Initialize downloader with the temporary directory
			downloader, err := fetch.NewYouTubeDownloader(tempDir)
			if err != nil {
				return fmt.Errorf("failed to initialize downloader: %w", err)
			}

			// Check if yt-dlp is installed
			if err := downloader.CheckYTDLP(); err != nil {
				fmt.Println("yt-dlp is required but not found or not working correctly.")
				fmt.Println("Please install yt-dlp: https://github.com/yt-dlp/yt-dlp#installation")
				os.Exit(1)
			}

			fmt.Println("Downloading video...")
			downloadedMetadata, err := downloader.DownloadAudio(videoURL) // Renamed variable
			if err != nil {
				return fmt.Errorf("failed to download video: %w", err)
			}

			// Initialize the transcriber
			whisperTranscriber, err := transcribe.NewWhisperTranscriber(apiKey, model)
			if err != nil {
				return fmt.Errorf("failed to initialize transcriber: %w", err)
			}

			fmt.Println("Transcribing audio with OpenAI Whisper API...")
			// Use the AudioFilePath field from the downloadedMetadata struct
			transcriptionText, err := whisperTranscriber.TranscribeFile(ctx, downloadedMetadata.AudioFilePath)
			if err != nil {
				return fmt.Errorf("failed to transcribe audio: %w", err)
			}

			fmt.Println(transcriptionText)

			return nil
		},
	}
)
