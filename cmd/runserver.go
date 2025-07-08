package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	internalHttp "github.com/exler/yt-transcribe/internal/http"
	"github.com/urfave/cli/v3"
)

var runserverCmd = &cli.Command{
	Name:  "runserver",
	Usage: "Start HTTP server for YouTube transcription and queue management",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "openai-api-key",
			Usage:   "OpenAI API key for transcription and summarization",
			Sources: cli.EnvVars("OPENAI_API_KEY"),
		},
		&cli.StringFlag{
			Name:  "transcription-model",
			Usage: "Whisper model to use for transcription",
			Value: "whisper-1",
		},
		&cli.StringFlag{
			Name:  "summarization-model",
			Usage: "LLM model to use for summarization",
			Value: "gpt-4.1-nano",
		},
		&cli.IntFlag{
			Name:  "port",
			Usage: "Port to run the HTTP server on",
			Value: 8000,
		},
		&cli.FloatFlag{
			Name:  "audio-speed-factor",
			Usage: "Speed up the audio by a factor",
			Value: 2.5,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		openaiAPIKey := cmd.String("openai-api-key")
		if openaiAPIKey == "" {
			return cli.Exit("OpenAI API key is required", 1)
		}

		transcriptionModel := cmd.String("transcription-model")
		if transcriptionModel == "" {
			return cli.Exit("Transcription model is required", 1)
		}

		summarizationModel := cmd.String("summarization-model")
		if summarizationModel == "" {
			return cli.Exit("Summarization model is required", 1)
		}

		audioSpeedFactor := cmd.Float("audio-speed-factor")

		server, err := internalHttp.NewServer(
			openaiAPIKey,
			summarizationModel,
		)
		if err != nil {
			return cli.Exit("Failed to initialize server: "+err.Error(), 1)
		}

		http.HandleFunc("/", server.IndexHandler)
		http.HandleFunc("/queue", server.QueueDataHandler)
		http.HandleFunc("/entry/{videoID}", server.EntryHandler)
		http.HandleFunc("/entry/{videoID}/summarize", server.EntrySummarizeHandler)

		staticFiles, err := fs.Sub(internalHttp.StaticFiles, "static")
		if err != nil {
			return cli.Exit("Failed to load static files: "+err.Error(), 1)
		}
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

		worker, err := internalHttp.NewTranscriptionWorker(openaiAPIKey, transcriptionModel, audioSpeedFactor)
		if err != nil {
			return cli.Exit("Failed to initialize transcription worker: "+err.Error(), 1)
		}

		go worker.RunTranscriptionWorker(ctx) // Launch the background worker

		port := cmd.Int("port")

		log.Printf("Running server on http://localhost:%d", port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

		return nil
	},
}
