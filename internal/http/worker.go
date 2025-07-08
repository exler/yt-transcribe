package http

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/exler/yt-transcribe/internal/ai"
	"github.com/exler/yt-transcribe/internal/fetch"
	"github.com/exler/yt-transcribe/internal/ffmpeg"
	"github.com/exler/yt-transcribe/internal/queue"
)

type TranscriptionWorker struct {
	transcriber      *ai.AudioTranscriber
	ffmpeg           *ffmpeg.FFMPEG
	audioSpeedFactor float64
}

func NewTranscriptionWorker(openaiAPIKey, model string, audioSpeedFactor float64) (*TranscriptionWorker, error) {
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

	ffmpeg, err := ffmpeg.NewFFMPEG()
	if err != nil {
		log.Fatalf("Failed to initialize ffmpeg: %v", err)
	}

	return &TranscriptionWorker{
		transcriber:      audioTranscriber,
		ffmpeg:           ffmpeg,
		audioSpeedFactor: audioSpeedFactor,
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

		// Post-process audio
		outputFile := filepath.Join(tempDir, "processed.mp3")
		log.Printf("Processing audio with ffmpeg (speed: %.2fx)...", w.audioSpeedFactor)
		err = w.ffmpeg.SpeedUpAudio(downloadedMetadata.AudioFilePath, outputFile, w.audioSpeedFactor)
		if err != nil {
			log.Printf("Error processing audio for %s: %v", videoInfo.VideoID, err)
			queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusFailed, "Failed to process audio: "+err.Error(), "", "")
			continue
		}

		// Transcribe audio
		queue.UpdateItem(videoInfo.VideoID, queue.VideoStatusTranscribing, "", "", "")

		transcriptionText, err := w.transcriber.TranscribeFile(ctx, outputFile)
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
