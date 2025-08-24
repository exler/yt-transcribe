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
			Usage:   "OpenAI API key for summarization",
			Sources: cli.EnvVars("OPENAI_API_KEY"),
		},
		&cli.StringFlag{
			Name:  "summarization-model",
			Usage: "LLM model to use for summarization",
			Value: "gpt-4.1-nano",
		},
		&cli.StringFlag{
			Name:  "whisper-model-path",
			Usage: "Path to ggml whisper.cpp model file (e.g. /models/ggml-small.bin)",
			Value: "/app/models/ggml-small.bin",
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
		&cli.IntFlag{
			Name:  "port",
			Usage: "Port to run the HTTP server on",
			Value: 8000,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		openaiAPIKey := cmd.String("openai-api-key")

		summarizationModel := cmd.String("summarization-model")

		modelPath := cmd.String("whisper-model-path")
		language := cmd.String("language")
		queueSize := cmd.Int("queue")

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

		worker, err := internalHttp.NewTranscriptionWorker(modelPath, language, queueSize)
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
