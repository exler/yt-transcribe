package cmd

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/queue"
	"github.com/exler/yt-transcribe/internal/summarize"
	"github.com/exler/yt-transcribe/internal/transcribe"
	"github.com/urfave/cli/v3"
)

// pageData holds the data for the template.
type pageData struct {
	Transcript             string
	Summary                string
	ErrorDetail            string // For general errors
	QueueAddSuccessMessage string
	QueueAddErrorMessage   string
}

func renderTemplate(w http.ResponseWriter, data pageData) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
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
		data.QueueAddErrorMessage = "YouTube URL is required."
		renderTemplate(w, data)
		return
	}

	// The OPENAI_API_KEY check might be relevant for workers, but not for adding to queue.
	// We can remove it from here for now, or keep it if other operations on the page still need it.
	// For this step, let's assume adding to queue doesn't need an API key directly.

	videoInfo, err := queue.Add(youtubeURL)
	if err != nil {
		log.Printf("Error adding video to queue: %v (URL: %s)", err, youtubeURL)
		data.QueueAddErrorMessage = err.Error()
	} else {
		log.Printf("Video added to queue: ID %s, Title: %s", videoInfo.VideoID, videoInfo.Title)
		data.QueueAddSuccessMessage = "Video '" + videoInfo.Title + "' added to queue successfully!"
	}

	renderTemplate(w, data)
}

func queueDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentQueue := queue.GetAll()
	w.Header().Set("Content-Type", "application/json")
	jsonData, err := json.Marshal(currentQueue)
	if err != nil {
		log.Printf("Error marshalling queue data: %v", err)
		http.Error(w, "Error preparing queue data", http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
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

	isAjax := r.Header.Get("X-Requested-With") == "XMLHttpRequest"
	textToSummarize := r.FormValue("text")

	// Define a helper struct for JSON responses
	type SummaryResponse struct {
		Summary    string `json:"summary"`
		Transcript string `json:"transcript"`
		Error      string `json:"error"`
	}

	if textToSummarize == "" {
		errorMsg := "Text to summarize is required."
		if isAjax {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(SummaryResponse{Error: errorMsg}); err != nil {
				log.Printf("Error encoding JSON response: %v", err)
				// Fallback to plain text error if JSON encoding fails for some reason
				http.Error(w, `{"error":"Text to summarize is required."}`, http.StatusBadRequest)
			}
		} else {
			data := pageData{ErrorDetail: errorMsg, Transcript: textToSummarize}
			renderTemplate(w, data)
		}
		return
	}

	// Initialize summarizer
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		errorMsg := "OpenAI API key not set. Please configure OPENAI_API_KEY environment variable."
		if isAjax {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(SummaryResponse{Error: errorMsg, Transcript: textToSummarize})
		} else {
			data := pageData{ErrorDetail: errorMsg, Transcript: textToSummarize}
			renderTemplate(w, data)
		}
		return
	}

	summarizer, err := summarize.NewLLMSummarizer(apiKey, "gpt-4.1-nano") // Consider model from config/param
	if err != nil {
		errorMsg := "Error initializing summarizer: " + err.Error()
		if isAjax {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(SummaryResponse{Error: errorMsg, Transcript: textToSummarize})
		} else {
			data := pageData{ErrorDetail: errorMsg, Transcript: textToSummarize}
			renderTemplate(w, data)
		}
		return
	}

	// Perform summarization
	summary, err := summarizer.SummarizeText(r.Context(), textToSummarize)
	if err != nil {
		errorMsg := "Error summarizing text: " + err.Error()
		if isAjax {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(SummaryResponse{Error: errorMsg, Transcript: textToSummarize})
		} else {
			data := pageData{ErrorDetail: errorMsg, Transcript: textToSummarize}
			renderTemplate(w, data)
		}
		return
	}

	// Success
	if isAjax {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SummaryResponse{
			Summary:    summary,
			Transcript: textToSummarize,
			Error:      "",
		})
	} else {
		data := pageData{Summary: summary, Transcript: textToSummarize}
		renderTemplate(w, data)
	}
}

var runserverCmd = &cli.Command{
	Name:  "runserver",
	Usage: "Start HTTP server for YouTube transcription and queue management",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		http.HandleFunc("/", mainHandler)
		http.HandleFunc("/summarize", summaryHandler)
		http.HandleFunc("/queue", queueDataHandler) // Register new queue handler

		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

		go startTranscriptionWorker() // Launch the background worker

		port := "8000"
		log.Printf("Running server on http://localhost:%s", port)
		log.Fatal(http.ListenAndServe(":"+port, nil))

		return nil
	},
}

func startTranscriptionWorker() {
	log.Println("Transcription worker started...")
	for {
		videoInfo := queue.GetNext()
		if videoInfo == nil {
			time.Sleep(5 * time.Second)
			continue
		}

		log.Printf("Processing video ID: %s, Title: %s", videoInfo.VideoID, videoInfo.Title)

		// Create a temporary directory for this video's processing
		tempDir, err := os.MkdirTemp("", "yt-transcribe-worker-*")
		if err != nil {
			log.Printf("Error creating temp directory for %s: %v", videoInfo.VideoID, err)
			queue.UpdateStatus(videoInfo.VideoID, "failed", "Failed to create temp directory: "+err.Error())
			continue
		}
		defer os.RemoveAll(tempDir) // Ensure cleanup

		// 1. Download Audio
		queue.UpdateStatus(videoInfo.VideoID, "downloading", "")
		downloader, err := fetch.NewYouTubeDownloader(tempDir)
		if err != nil {
			log.Printf("Error initializing downloader for %s: %v", videoInfo.VideoID, err)
			queue.UpdateStatus(videoInfo.VideoID, "failed", "Failed to initialize downloader: "+err.Error())
			continue
		}

		// Construct video URL. Assuming VideoID is just the ID.
		videoURL := "https://www.youtube.com/watch?v=" + videoInfo.VideoID
		downloadedMetadata, err := downloader.DownloadAudio(videoURL) // This returns (fetch.VideoMetadata, error)
		if err != nil {
			log.Printf("Error downloading audio for %s: %v", videoInfo.VideoID, err)
			queue.UpdateStatus(videoInfo.VideoID, "failed", "Failed to download audio: "+err.Error())
			continue
		}
		queue.SetAudioPath(videoInfo.VideoID, downloadedMetadata.AudioFilePath)
		log.Printf("Audio downloaded for %s to %s", videoInfo.VideoID, downloadedMetadata.AudioFilePath)

		// 2. Transcribe Audio
		queue.UpdateStatus(videoInfo.VideoID, "transcribing", "")
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Printf("OpenAI API key not set for %s.", videoInfo.VideoID)
			queue.UpdateStatus(videoInfo.VideoID, "failed", "OpenAI API key not configured on server.")
			continue
		}

		whisperTranscriber, err := transcribe.NewWhisperTranscriber(apiKey, "whisper-1")
		if err != nil {
			log.Printf("Error initializing transcriber for %s: %v", videoInfo.VideoID, err)
			queue.UpdateStatus(videoInfo.VideoID, "failed", "Failed to initialize transcriber: "+err.Error())
			continue
		}

		// Use a background context for transcription
		transcriptionText, err := whisperTranscriber.TranscribeFile(context.Background(), downloadedMetadata.AudioFilePath)
		if err != nil {
			log.Printf("Error transcribing audio for %s: %v", videoInfo.VideoID, err)
			queue.UpdateStatus(videoInfo.VideoID, "failed", "Failed to transcribe audio: "+err.Error())
			continue
		}

		queue.UpdateStatus(videoInfo.VideoID, "completed", "", transcriptionText)
		log.Printf("Successfully transcribed video ID: %s, Title: %s", videoInfo.VideoID, videoInfo.Title)
	}
}
