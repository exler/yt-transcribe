package transcribe

import (
	"context"
	"errors"
	"fmt"
	"os"

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

	if model == "" {
		return nil, errors.New("model is required")
	}

	// Create OpenAI client with API key
	client := openai.NewClient(option.WithAPIKey(apiKey))

	return &WhisperTranscriber{
		client: &client,
		model:  model,
	}, nil
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
