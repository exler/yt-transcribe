package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/exler/yt-transcribe/internal/ai"
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
				Name:    "openai-api-key",
				Usage:   "OpenAI API key for Whisper transcription",
				Sources: cli.EnvVars("OPENAI_API_KEY"),
			},
			&cli.StringFlag{
				Name:  "model",
				Usage: "Whisper model to use for transcription",
				Value: "whisper-1",
			},
			&cli.FloatFlag{
				Name:  "audio-speed-factor",
				Usage: "Speed up the audio by a factor",
				Value: 2.5,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			videoURL := cmd.Args().First()
			if videoURL == "" {
				return cli.Exit("Please provide a YouTube video URL to transcribe", 1)
			}

			apiKey := cmd.String("openai-api-key")
			if apiKey == "" {
				return cli.Exit("OpenAI API key is required", 1)
			}

			model := cmd.String("model")
			if model == "" {
				return cli.Exit("Transcription model is required", 1)
			}

			audioSpeedFactor := cmd.Float("audio-speed-factor")

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

			ffmpegProcessor, err := ffmpeg.NewFFMPEG()
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to initialize ffmpeg: %v", err), 1)
			}

			outputFile := filepath.Join(tempDir, "processed.mp3")
			fmt.Printf("Processing audio with ffmpeg (speed: %.2fx)...\n", audioSpeedFactor)
			err = ffmpegProcessor.SpeedUpAudio(downloadedMetadata.AudioFilePath, outputFile, audioSpeedFactor)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to process audio with ffmpeg: %v", err), 1)
			}

			audioTranscriber, err := ai.NewAudioTranscriber(apiKey, model)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to initialize audio transcriber: %v", err), 1)
			}

			fmt.Println("Transcribing audio with OpenAI API...")
			transcriptionText, err := audioTranscriber.TranscribeFile(ctx, outputFile)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to transcribe audio: %v", err), 1)
			}

			fmt.Println(transcriptionText)

			return nil
		},
	}
)
