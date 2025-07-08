package http

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/exler/yt-transcribe/internal/ai"
	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/queue"
)

type TranscriptionWorker struct {
	transcriber *ai.AudioTranscriber
}

func NewTranscriptionWorker(openaiAPIKey, model string) (*TranscriptionWorker, error) {
	if openaiAPIKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}
	if model == "" {
		return nil, errors.New("transcription model is required")
	}

	audioTranscriber, err := ai.NewAudioTranscriber(openaiAPIKey, model)
	if err != nil {
		log.Fatalf("Failed to initialize audio transcriber: %v", err)
	}

	return &TranscriptionWorker{
		transcriber: audioTranscriber,
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
		defer os.RemoveAll(tempDir) // Ensure cleanup

		// 1. Download Audio
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

		// 2. Transcribe Audio
		queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusTranscribing, "", "", "")

		transcriptionText, err := w.transcriber.TranscribeFile(ctx, downloadedMetadata.AudioFilePath)
		if err != nil {
			log.Printf("Error transcribing audio for %s: %v", videoInfo.VideoID, err)
			queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusFailed, "Failed to transcribe audio: "+err.Error(), "", "")
			continue
		}

		queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusCompleted, "", transcriptionText, "")
		log.Printf("Successfully transcribed video ID: %s, Title: %s", videoInfo.VideoID, videoInfo.Title)
	}
}
