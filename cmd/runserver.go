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
			Name:    "llm-endpoint",
			Usage:   "Endpoint URL for LLM API (e.g., http://localhost:11434/v1 for Ollama, https://api.openai.com/v1 for OpenAI). Leave empty to disable summarization.",
			Value:   "",
			Sources: cli.EnvVars("LLM_ENDPOINT"),
		},
		&cli.StringFlag{
			Name:    "llm-token",
			Usage:   "API token for LLM (required for OpenAI, optional for Ollama)",
			Value:   "",
			Sources: cli.EnvVars("LLM_TOKEN"),
		},
		&cli.StringFlag{
			Name:    "llm-model",
			Usage:   "Model name to use for LLM",
			Value:   "phi3:mini",
			Sources: cli.EnvVars("LLM_MODEL"),
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
		llmEndpoint := cmd.String("llm-endpoint")
		llmToken := cmd.String("llm-token")
		llmModel := cmd.String("llm-model")
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

		worker, err := internalHttp.NewTranscriptionWorker(llmEndpoint, llmToken, llmModel, whisperModelPath, whisperLanguage, whisperQueueSize)
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
