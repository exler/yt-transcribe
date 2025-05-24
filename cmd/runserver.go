package cmd

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/summarize"
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

	{{if .ErrorDetail}}
		<p style="color: red;">Error: {{.ErrorDetail}}</p>
	{{end}}

	{{if .Summary}}
		<h2>Summary:</h2>
		<p>{{.Summary}}</p>
	{{end}}
	
	{{if .Transcript}}
		<h2>Transcript:</h2>
		<p>{{.Transcript}}</p>

		<form method="POST" action="/summarize">
			<input type="hidden" name="text" value="{{.Transcript}}">
			<input type="submit" value="Summarize Transcript">
		</form>
	{{end}}
</body>
</html>`

// pageData holds the data for the template.
type pageData struct {
	Transcript  string
	Summary     string
	ErrorDetail string
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

func getHandler(w http.ResponseWriter, _ *http.Request) {
	data := pageData{}
	// Render the initial page with an empty transcript
	renderTemplate(w, data)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	data := pageData{}
	youtubeURL := r.FormValue("youtube_url")
	if youtubeURL == "" {
		data.ErrorDetail = "YouTube URL is required."
		renderTemplate(w, data)
		return
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		data.ErrorDetail = "OpenAI API key not set. Please configure OPENAI_API_KEY environment variable."
		renderTemplate(w, data)
		return
	}

	tempDir, err := os.MkdirTemp("", "yt-transcribe-*")
	if err != nil {
		data.ErrorDetail = "Error creating temporary directory: " + err.Error()
		renderTemplate(w, data)
		return
	}
	defer os.RemoveAll(tempDir)

	downloader, err := fetch.NewYouTubeDownloader(tempDir)
	if err != nil {
		data.ErrorDetail = "Error initializing downloader: " + err.Error()
		renderTemplate(w, data)
		return
	} else if err := downloader.CheckYTDLP(); err != nil {
		data.ErrorDetail = "yt-dlp check failed: " + err.Error()
		renderTemplate(w, data)
		return
	}

	audioOutputPath, err := downloader.DownloadAudio(youtubeURL)
	if err != nil {
		data.ErrorDetail = "Error downloading audio: " + err.Error()
		renderTemplate(w, data)
		return
	}

	model := "whisper-1"
	whisperTranscriber, err := transcribe.NewWhisperTranscriber(apiKey, model)
	if err != nil {
		data.ErrorDetail = "Error initializing transcriber: " + err.Error()
		renderTemplate(w, data)
		return
	}

	transcriptionText, err := whisperTranscriber.TranscribeFile(r.Context(), audioOutputPath)
	if err != nil {
		data.ErrorDetail = "Error transcribing audio: " + err.Error()
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

func summaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := pageData{}
	textToSummarize := r.FormValue("text")
	if textToSummarize == "" {
		data.ErrorDetail = "Text to summarize is required."
		renderTemplate(w, data)
		return
	}

	data.Transcript = textToSummarize

	summarizer, err := summarize.NewLLMSummarizer(os.Getenv("OPENAI_API_KEY"), "gpt-4.1-nano")
	if err != nil {
		data.ErrorDetail = "Error initializing summarizer: " + err.Error()
		renderTemplate(w, data)
		return
	}

	summary, err := summarizer.SummarizeText(r.Context(), textToSummarize)
	if err != nil {
		data.ErrorDetail = "Error summarizing text: " + err.Error()
		renderTemplate(w, data)
		return
	}

	data.Summary = summary
	renderTemplate(w, data)
}

var runserverCmd = &cli.Command{
	Name:  "runserver",
	Usage: "Start HTTP server for YouTube transcription",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		http.HandleFunc("/", mainHandler)
		http.HandleFunc("/summarize", summaryHandler)

		port := "8000"
		log.Printf("Running server on http://localhost:%s", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))

		return nil
	},
}
