package ai

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// AudioTranscriber handles audio transcription using OpenAI's transcription API
// https://platform.openai.com/docs/guides/speech-to-text
type AudioTranscriber struct {
	client *openai.Client
	model  string
}

// NewAudioTranscriber creates a new transcriber using OpenAI's Whisper API
func NewAudioTranscriber(apiKey, model string) (*AudioTranscriber, error) {
	if apiKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	if model == "" {
		return nil, errors.New("model is required")
	}

	// Create OpenAI client with API key
	client := openai.NewClient(option.WithAPIKey(apiKey))

	return &AudioTranscriber{
		client: &client,
		model:  model,
	}, nil
}

// TranscribeFile transcribes the given audio file using OpenAI's Whisper API
func (t *AudioTranscriber) TranscribeFile(ctx context.Context, filePath string) (string, error) {
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
