package http

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/ffmpeg"
	"github.com/exler/yt-transcribe/internal/queue"
)

type TranscriptionWorker struct {
	ffmpeg *ffmpeg.FFMPEG
	// Path to the `ggml` converted Whisper models.
	// https://github.com/ggml-org/whisper.cpp/blob/master/models/README.md
	ffmpegWhisperModelPath string
	// Transcription language or `auto` for automatic detection.
	// Make sure your model supports the specified language.
	ffmpegTranscriptionLanguage string
	// Maximum size that will be queued into the filter before processing the audio.
	ffmpegQueueSize int
}

func NewTranscriptionWorker(ffmpegWhisperModelPath, ffmpegTranscriptionLanguage string, ffmpegQueueSize int) (*TranscriptionWorker, error) {
	if ffmpegWhisperModelPath == "" {
		return nil, errors.New("whisper model path is required")
	}
	if ffmpegTranscriptionLanguage == "" {
		ffmpegTranscriptionLanguage = "auto"
	}

	f, err := ffmpeg.NewFFMPEG()
	if err != nil {
		log.Fatalf("Failed to initialize ffmpeg: %v", err)
	}

	return &TranscriptionWorker{
		ffmpeg:                      f,
		ffmpegWhisperModelPath:      ffmpegWhisperModelPath,
		ffmpegTranscriptionLanguage: ffmpegTranscriptionLanguage,
		ffmpegQueueSize:             ffmpegQueueSize,
	}, nil
}

func (w *TranscriptionWorker) RunTranscriptionWorker(ctx context.Context) {
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
			queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusFailed, "Failed to create temp directory: "+err.Error(), "", "")
			continue
		}

		// Download audio
		queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusDownloading, "", "", "")
		downloader, err := fetch.NewYouTubeDownloader(tempDir)
		if err != nil {
			log.Printf("Error initializing downloader for %s: %v", videoInfo.VideoID, err)
			queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusFailed, "Failed to initialize downloader: "+err.Error(), "", "")
			continue
		}

		downloadedMetadata, err := downloader.DownloadAudio(videoInfo.VideoURL)
		if err != nil {
			log.Printf("Error downloading audio for %s: %v", videoInfo.VideoID, err)
			queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusFailed, "Failed to download audio: "+err.Error(), "", "")
			continue
		}
		queue.SetAudioPath(videoInfo.VideoID, downloadedMetadata.AudioFilePath)
		log.Printf("Audio downloaded for %s to %s", videoInfo.VideoID, downloadedMetadata.AudioFilePath)

		// Transcribe audio using FFmpeg whisper filter
		queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusTranscribing, "", "", "")

		transcriptionText, err := w.ffmpeg.TranscribeWithWhisperFilter(downloadedMetadata.AudioFilePath, w.ffmpegWhisperModelPath, w.ffmpegTranscriptionLanguage, w.ffmpegQueueSize)
		if err != nil {
			log.Printf("Error transcribing audio for %s: %v", videoInfo.VideoID, err)
			queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusFailed, "Failed to transcribe audio: "+err.Error(), "", "")
			continue
		}

		// Clean up temporary files
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Error cleaning up temp directory for %s: %v", videoInfo.VideoID, err)
		}

		queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusCompleted, "", transcriptionText, "")
		log.Printf("Successfully transcribed video ID: %s, Title: %s", videoInfo.VideoID, videoInfo.Title)
	}
}
