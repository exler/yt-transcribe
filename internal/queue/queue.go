package queue

import (
	"fmt"
	"sync"

	"github.com/exler/yt-transcribe/internal/fetch" // Assuming this is the correct import path
)

// VideoInfo holds all information about a video in the transcription queue.
type VideoInfo struct {
	VideoID       string
	Title         string
	Duration      string
	UploadDate    string
	Status        string // "pending", "fetching_metadata", "metadata_failed", "processing", "downloading", "download_failed", "transcribing", "transcription_failed", "completed", "failed"
	AudioFilePath string
	Transcript    string
	Error         string
}

var (
	transcriptionQueue []*VideoInfo
	queueMutex         sync.Mutex
)

// Initialize the queue (optional, as zero value for slice is nil)
func init() {
	transcriptionQueue = make([]*VideoInfo, 0)
}

// Add attempts to fetch video metadata and adds it to the queue.
// It does NOT download the audio file itself.
func Add(videoURL string) (*VideoInfo, error) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	// For metadata fetching, OutputDir might not be strictly needed by yt-dlp if not downloading.
	// However, the existing NewYouTubeDownloader requires it.
	// We'll pass a temporary/placeholder value. The actual downloader worker
	// will use a proper OutputDir.
	// For metadata fetching, OutputDir of YouTubeDownloader is not used by GetVideoMetadata.
	// Pass an empty string for outputDir, which NewYouTubeDownloader will resolve to the current dir.
	// This is acceptable as GetVideoMetadata does not write files.
	downloader, err := fetch.NewYouTubeDownloader("") // OutputDir not used by GetVideoMetadata
	if err != nil {
		return nil, fmt.Errorf("failed to initialize downloader: %w", err)
	}

	// Call the new GetVideoMetadata function which only fetches metadata
	videoMeta, err := downloader.GetVideoMetadata(videoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	// Check for existing VideoID
	for _, item := range transcriptionQueue {
		if item.VideoID == videoMeta.VideoID {
			return item, fmt.Errorf("video %s already in queue", videoMeta.VideoID)
		}
	}

	info := &VideoInfo{
		VideoID:    videoMeta.VideoID,
		Title:      videoMeta.Title,
		Duration:   videoMeta.Duration,
		UploadDate: videoMeta.UploadDate,
		Status:     "pending", // Initial status
		// AudioFilePath will be set by the worker after actual download
		AudioFilePath: "", // Explicitly empty
		Transcript:    "",
		Error:         "",
	}

	transcriptionQueue = append(transcriptionQueue, info)
	return info, nil
}

// GetNext finds the next "pending" video, sets its status to "processing", and returns it.
func GetNext() *VideoInfo {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	for _, item := range transcriptionQueue {
		if item.Status == "pending" {
			item.Status = "processing" // Mark as processing
			return item
		}
	}
	return nil // No pending items
}

// UpdateStatus updates the status and optionally the error message and transcript of a video.
func UpdateStatus(videoID string, status string, errorMessage string, transcript ...string) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	for _, item := range transcriptionQueue {
		if item.VideoID == videoID {
			item.Status = status
			if errorMessage != "" {
				item.Error = errorMessage
			}
			if len(transcript) > 0 {
				item.Transcript = transcript[0]
			}
			return
		}
	}
}

// SetAudioPath sets the audio file path for a given video.
func SetAudioPath(videoID string, audioPath string) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	for _, item := range transcriptionQueue {
		if item.VideoID == videoID {
			item.AudioFilePath = audioPath
			return
		}
	}
}

// GetAll returns a copy of the current queue.
func GetAll() []*VideoInfo {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	// Return a copy to prevent external modification
	queueCopy := make([]*VideoInfo, len(transcriptionQueue))
	for i, item := range transcriptionQueue {
		// Shallow copy of the VideoInfo struct is usually fine if fields are simple types.
		// If VideoInfo contained pointers/slices that could be modified, a deep copy might be needed.
		// For now, a shallow copy of each item is made into the new slice.
		itemCopy := *item
		queueCopy[i] = &itemCopy
	}
	return queueCopy
}

// Helper function to clear the queue, useful for testing
func ClearQueueForTesting() {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	transcriptionQueue = make([]*VideoInfo, 0)
}
