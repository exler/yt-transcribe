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
		&cli.IntFlag{
			Name:  "port",
			Usage: "Port to run the HTTP server on",
			Value: 8000,
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		ollamaBaseURL := cmd.String("ollama-base-url")
		ollamaModel := cmd.String("ollama-model")
		whisperModelPath := cmd.String("whisper-model-path")
		whisperLanguage := cmd.String("whisper-language")
		whisperQueueSize := cmd.Int("whisper-queue")

		server, err := internalHttp.NewServer()
		if err != nil {
			return cli.Exit("Failed to initialize server: "+err.Error(), 1)
		}

		http.HandleFunc("/", server.IndexHandler)
		http.HandleFunc("/queue", server.QueueDataHandler)
		http.HandleFunc("/entry/{videoID}", server.EntryHandler)

		staticFiles, err := fs.Sub(internalHttp.StaticFiles, "static")
		if err != nil {
			return cli.Exit("Failed to load static files: "+err.Error(), 1)
		}
		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFiles))))

		worker, err := internalHttp.NewTranscriptionWorker(ollamaBaseURL, ollamaModel, whisperModelPath, whisperLanguage, whisperQueueSize)
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
