package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/exler/yt-transcribe/internal/ai"
	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/queue"
)

type Server struct {
	summarizer *ai.LLMSummarizer
}

func NewServer(openaiAPIKey, model string) (*Server, error) {
	summarizer, err := ai.NewLLMSummarizer(openaiAPIKey, model)
	if err != nil {
		return nil, err
	}
	return &Server{summarizer: summarizer}, nil
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	data := pageData{}

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodGet {
		renderTemplate(w, "index", data)
		return
	}

	youtubeURL := r.FormValue("youtube_url")

	if youtubeURL == "" {
		data.QueueAddErrorMessage = "YouTube URL is required."
		renderTemplate(w, "index", data)
		return
	}

	downloader, err := fetch.NewYouTubeDownloader("") // OutputDir not used by GetVideoMetadata
	if err != nil {
		log.Printf("Error initializing YouTube downloader: %v", err)
		data.QueueAddErrorMessage = "Failed to initialize YouTube downloader."
		renderTemplate(w, "index", data)
		return
	}

	// Call the new GetVideoMetadata function which only fetches metadata
	videoMeta, err := downloader.GetVideoMetadata(youtubeURL)
	if err != nil {
		log.Printf("Error fetching video metadata: %v", err)
		data.QueueAddErrorMessage = "Failed to fetch video metadata."
		renderTemplate(w, "index", data)
		return
	}

	videoInfo, err := queue.Add(queue.NewVideoInfo{
		VideoURL:   youtubeURL,
		VideoID:    videoMeta.VideoID,
		Title:      videoMeta.Title,
		Duration:   videoMeta.Duration,
		UploadDate: videoMeta.UploadDate,
	})
	if err != nil {
		log.Printf("Error adding video to queue: %v (URL: %s)", err, youtubeURL)
		data.QueueAddErrorMessage = err.Error()
	} else {
		log.Printf("Video added to queue: ID %s, Title: %s", videoInfo.VideoID, videoInfo.Title)
		data.QueueAddSuccessMessage = "Video '" + videoInfo.Title + "' added to queue successfully!"
	}

	renderTemplate(w, "index", data)
}

func (s *Server) QueueDataHandler(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) EntryHandler(w http.ResponseWriter, r *http.Request) {
	videoID := r.PathValue("videoID")

	var found *queue.VideoInfo
	for _, v := range queue.GetAll() {
		if v.VideoID == videoID {
			found = v
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}

	renderTemplate(w, "entry", pageData{
		Title:                  found.Title,
		VideoID:                found.VideoID,
		Transcript:             found.Transcript,
		Summary:                found.Summary,
		ErrorDetail:            found.Error,
		Status:                 found.Status,
		QueueAddSuccessMessage: "",
		QueueAddErrorMessage:   "",
	})
}

func (s *Server) EntrySummarizeHandler(w http.ResponseWriter, r *http.Request) {
	videoID := r.PathValue("videoID")

	var found *queue.VideoInfo
	for _, v := range queue.GetAll() {
		if v.VideoID == videoID {
			found = v
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}

	summary, err := s.summarizer.SummarizeText(r.Context(), found.Transcript)
	if err != nil {
		http.Error(w, "Error summarizing transcript", http.StatusInternalServerError)
		return
	}
	found.Summary = summary

	queue.UpdateItem(found.VideoID, found.Status, "", found.Transcript, found.Summary)

	renderTemplate(w, "entry", pageData{
		Title:                  found.Title,
		VideoID:                found.VideoID,
		Transcript:             found.Transcript,
		Summary:                found.Summary,
		ErrorDetail:            found.Error,
		Status:                 found.Status,
		QueueAddSuccessMessage: "",
		QueueAddErrorMessage:   "",
	})
}
