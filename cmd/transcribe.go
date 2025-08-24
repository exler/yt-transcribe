package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/ffmpeg"
	"github.com/exler/yt-transcribe/internal/ollama"
	"github.com/urfave/cli/v3"
)

var (
	transcribeCmd = &cli.Command{
		Name:  "transcribe",
		Usage: "Transcribe a YouTube video",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "summarize",
				Usage: "Whether to summarize the transcription",
				Value: false,
			},
			&cli.StringFlag{
				Name:    "ollama-base-url",
				Usage:   "Base URL for Ollama",
				Value:   "http://localhost:11434",
				Sources: cli.EnvVars("OLLAMA_BASE_URL"),
			},
			&cli.StringFlag{
				Name:    "ollama-model",
				Usage:   "Model name to use for Ollama",
				Value:   "phi3:mini",
				Sources: cli.EnvVars("OLLAMA_MODEL"),
			},
			&cli.StringFlag{
				Name:    "whisper-model-path",
				Usage:   "Path to ggml whisper.cpp model file",
				Value:   "models/ggml-small.bin",
				Sources: cli.EnvVars("WHISPER_MODEL_PATH"),
			},
			&cli.StringFlag{
				Name:    "whisper-language",
				Usage:   "Language code to use or 'auto' to autodetect",
				Value:   "auto",
				Sources: cli.EnvVars("WHISPER_LANGUAGE"),
			},
			&cli.IntFlag{
				Name:    "whisper-queue",
				Usage:   "FFmpeg whisper filter queue size in seconds",
				Value:   15,
				Sources: cli.EnvVars("WHISPER_QUEUE"),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			videoURL := cmd.Args().First()
			if videoURL == "" {
				return cli.Exit("Please provide a YouTube video URL to transcribe", 1)
			}

			summarize := cmd.Bool("summarize")
			ollamaBaseURL := cmd.String("ollama-base-url")
			ollamaModel := cmd.String("ollama-model")
			whisperModelPath := cmd.String("whisper-model-path")
			whisperLanguage := cmd.String("whisper-language")
			whisperQueueSize := cmd.Int("whisper-queue")

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

			fmt.Println("Transcribing audio with FFmpeg whisper filter...")
			transcriptionText, err := ff.TranscribeWithWhisperFilter(downloadedMetadata.AudioFilePath, whisperModelPath, whisperLanguage, whisperQueueSize)
			if err != nil {
				return cli.Exit(fmt.Sprintf("Failed to transcribe audio with whisper filter: %v", err), 1)
			}

			if summarize {
				ollama, err := ollama.NewOllama(ollamaBaseURL, ollamaModel)
				if err != nil {
					return cli.Exit(fmt.Sprintf("Failed to initialize Ollama: %v", err), 1)
				}

				summary, err := ollama.SummarizeText(ctx, downloadedMetadata.Title, transcriptionText)
				if err != nil {
					return cli.Exit(fmt.Sprintf("Failed to summarize transcription: %v", err), 1)
				}

				fmt.Println("Summary:")
				fmt.Println(summary)
			} else {
				fmt.Println(transcriptionText)
			}

			return nil
		},
	}
)
