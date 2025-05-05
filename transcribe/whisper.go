package transcribe

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// WhisperTranscriber handles audio transcription using OpenAI's Whisper API
type WhisperTranscriber struct {
	client *openai.Client
	model  string
}

// NewWhisperTranscriber creates a new transcriber using OpenAI's Whisper API
func NewWhisperTranscriber(apiKey, model string) (*WhisperTranscriber, error) {
	if apiKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	// Set default model if empty
	if model == "" {
		model = "whisper-1"
	}

	// Create OpenAI client with API key
	client := openai.NewClient(option.WithAPIKey(apiKey))

	return &WhisperTranscriber{
		client: &client,
		model:  model,
	}, nil
}

// FindAudioFile searches for audio files in the given directory
func FindAudioFile(directory string) (string, error) {
	files, err := os.ReadDir(directory)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	// List of common audio extensions
	audioExts := []string{".mp3", ".m4a", ".wav", ".flac", ".ogg", ".webm"}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		ext := strings.ToLower(filepath.Ext(fileName))

		for _, audioExt := range audioExts {
			if ext == audioExt {
				return filepath.Join(directory, fileName), nil
			}
		}
	}

	return "", errors.New("no audio file found in the directory")
}

// TranscribeFile transcribes the given audio file using OpenAI's Whisper API
func (t *WhisperTranscriber) TranscribeFile(ctx context.Context, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create transcription request parameters
	params := openai.AudioTranscriptionNewParams{
		Model: t.model,
		File:  file,
	}

	// Call OpenAI API to transcribe the audio
	resp, err := t.client.Audio.Transcriptions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	return resp.Text, nil
}
