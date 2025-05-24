package cmd

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/transcribe"
	"github.com/urfave/cli/v3"
)

// indexHTML is the HTML template for the server page.
var indexHTML = `<html>
<head><title>YouTube Transcription</title></head>
<body>
	<h1>YouTube Transcription</h1>
	<form method="POST" action="/">
		<input type="text" name="youtube_url" placeholder="Enter YouTube URL" size="50">
		<input type="submit" value="Transcribe">
	</form>
	{{if .Transcript}}
		<h2>Transcript:</h2>
		<p>{{.Transcript}}</p>
	{{end}}
</body>
</html>`

// pageData holds the data for the template.
type pageData struct {
	Transcript string
}

func renderTemplate(w http.ResponseWriter, data pageData) {
	tmpl, err := template.New("index").Parse(indexHTML)
	if err != nil {
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	data := pageData{}
	// Render the initial page with an empty transcript
	renderTemplate(w, data)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	data := pageData{}
	youtubeURL := r.FormValue("youtube_url")
	if youtubeURL == "" {
		data.Transcript = "YouTube URL is required."
		renderTemplate(w, data)
		return
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		data.Transcript = "OpenAI API key not set. Please configure OPENAI_API_KEY environment variable."
		renderTemplate(w, data)
		return
	}

	tempDir, err := os.MkdirTemp("", "yt-transcribe-*")
	if err != nil {
		data.Transcript = "Error creating temporary directory: " + err.Error()
		renderTemplate(w, data)
		return
	}
	defer os.RemoveAll(tempDir)

	downloader, err := fetch.NewYouTubeDownloader(tempDir)
	if err != nil {
		data.Transcript = "Error initializing downloader: " + err.Error()
		renderTemplate(w, data)
		return
	} else if err := downloader.CheckYTDLP(); err != nil {
		data.Transcript = "yt-dlp check failed: " + err.Error()
		renderTemplate(w, data)
		return
	}

	audioOutputPath, err := downloader.DownloadAudio(youtubeURL)
	if err != nil {
		data.Transcript = "Error downloading audio: " + err.Error()
		renderTemplate(w, data)
		return
	}

	model := "whisper-1"
	whisperTranscriber, err := transcribe.NewWhisperTranscriber(apiKey, model)
	if err != nil {
		data.Transcript = "Error initializing transcriber: " + err.Error()
		renderTemplate(w, data)
		return
	}

	transcriptionText, err := whisperTranscriber.TranscribeFile(r.Context(), audioOutputPath)
	if err != nil {
		data.Transcript = "Error transcribing audio: " + err.Error()
		renderTemplate(w, data)
		return
	}

	data.Transcript = transcriptionText
	renderTemplate(w, data)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getHandler(w, r)
	} else if r.Method == http.MethodPost {
		postHandler(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

var runserverCmd = &cli.Command{
	Name:  "runserver",
	Usage: "Start HTTP server for YouTube transcription",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		http.HandleFunc("/", mainHandler)

		port := "8000"
		log.Printf("Running server on http://localhost:%s", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))

		return nil
	},
}
