package queue

import (
	"fmt"
	"sync"
)

type VideoStatus string

const (
	// VideoStatusPending indicates the video is waiting to be processed
	VideoStatusPending             VideoStatus = "pending"
	VideoStatusFetchingMetadata    VideoStatus = "fetching_metadata"
	VideoStatusMetadataFailed      VideoStatus = "metadata_failed"
	VideoStatusProcessing          VideoStatus = "processing"
	VideoStatusDownloading         VideoStatus = "downloading"
	VideoStatusDownloadFailed      VideoStatus = "download_failed"
	VideoStatusTranscribing        VideoStatus = "transcribing"
	VideoStatusTranscriptionFailed VideoStatus = "transcription_failed"
	VideoStatusSummarizing         VideoStatus = "summarizing"
	VideoStatusSummaryFailed       VideoStatus = "summary_failed"
	VideoStatusCompleted           VideoStatus = "completed"
	VideoStatusFailed              VideoStatus = "failed"
)

// VideoInfo holds all information about a video in the transcription queue.
type VideoInfo struct {
	VideoURL      string // The original user-supplied YouTube URL
	VideoID       string
	Title         string
	Duration      string
	UploadDate    string
	Status        VideoStatus
	AudioFilePath string
	Transcript    string
	Summary       string
	Error         string
}

// NewVideoInfo is a simplified struct for adding new videos to the queue.
type NewVideoInfo struct {
	VideoURL   string
	VideoID    string
	Title      string
	Duration   string
	UploadDate string
}

var (
	transcriptionQueue []*VideoInfo
	queueMutex         sync.Mutex
)

func init() {
	transcriptionQueue = make([]*VideoInfo, 0)
}

// Add attempts to fetch video metadata and adds it to the queue.
// It does NOT download the audio file itself.
func Add(initialInfo NewVideoInfo) (*VideoInfo, error) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	// Check for existing VideoID
	for _, item := range transcriptionQueue {
		if item.VideoID == initialInfo.VideoID {
			return item, fmt.Errorf("video %s already in queue", initialInfo.VideoID)
		}
	}

	finalInfo := &VideoInfo{
		VideoURL:      initialInfo.VideoURL,
		VideoID:       initialInfo.VideoID,
		Title:         initialInfo.Title,
		Duration:      initialInfo.Duration,
		UploadDate:    initialInfo.UploadDate,
		Status:        VideoStatusPending, // Initial status
		AudioFilePath: "",
		Transcript:    "",
		Summary:       "",
		Error:         "",
	}

	transcriptionQueue = append(transcriptionQueue, finalInfo)
	return finalInfo, nil
}

// GetNext finds the next "pending" video, sets its status to "processing", and returns it.
func GetNext() *VideoInfo {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	for _, item := range transcriptionQueue {
		if item.Status == VideoStatusPending {
			item.Status = VideoStatusProcessing
			return item
		}
	}

	// No pending items
	return nil
}

// UpdateItem updates the status and optionally the error message and transcript of a video.
func UpdateItem(videoID string, status VideoStatus, errorMessage string, transcript string, summary string) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	for _, item := range transcriptionQueue {
		if item.VideoID == videoID {
			item.Status = status
			item.Error = errorMessage
			item.Transcript = transcript
			item.Summary = summary
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

// GetAll returns a copy of the current queue in LIFO order.
func GetAll() []*VideoInfo {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	length := len(transcriptionQueue)
	queueCopy := make([]*VideoInfo, length)

	for i, item := range transcriptionQueue {
		itemCopy := *item
		queueCopy[length-1-i] = &itemCopy
	}
	return queueCopy
}

// Helper function to clear the queue
func ClearQueue() {
	queueMutex.Lock()
	defer queueMutex.Unlock()
	transcriptionQueue = make([]*VideoInfo, 0)
}
