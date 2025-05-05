package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/exler/yt-transcribe/fetch"
	"github.com/exler/yt-transcribe/transcribe"
	"github.com/urfave/cli/v3"
)

var (
	transcribeCmd = &cli.Command{
		Name:  "transcribe",
		Usage: "Transcribe a YouTube video",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output-dir",
				Aliases: []string{"o"},
				Usage:   "Directory where downloaded videos will be stored",
				Value:   "downloads",
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Video format to download (default: bestaudio/best)",
				Value:   "bestaudio/best",
			},
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

			outputDir := cmd.String("output-dir")
			format := cmd.String("format")

			// Check for API key in environment variable if not provided via flag
			apiKey := cmd.String("api-key")
			if apiKey == "" {
				apiKey = os.Getenv("OPENAI_API_KEY")
			}

			model := cmd.String("model")

			fmt.Printf("Downloading video: %s\n", videoURL)
			fmt.Printf("Output directory: %s\n", outputDir)

			downloader, err := fetch.NewVideoDownloader(outputDir)
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
			outputPath, err := downloader.DownloadVideo(videoURL, "--format", format)
			if err != nil {
				return fmt.Errorf("failed to download video: %w", err)
			}

			fmt.Printf("Video successfully downloaded to: %s\n", outputPath)

			// Find the audio file in the output directory
			audioFilePath, err := transcribe.FindAudioFile(outputPath)
			if err != nil {
				return fmt.Errorf("failed to find audio file: %w", err)
			}

			// Initialize the transcriber
			whisperTranscriber, err := transcribe.NewWhisperTranscriber(apiKey, model)
			if err != nil {
				return fmt.Errorf("failed to initialize transcriber: %w", err)
			}

			fmt.Println("Transcribing audio with OpenAI Whisper API...")
			transcriptionText, err := whisperTranscriber.TranscribeFile(ctx, audioFilePath)
			if err != nil {
				return fmt.Errorf("failed to transcribe audio: %w", err)
			}

			fmt.Println(transcriptionText)

			return nil
		},
	}
)
