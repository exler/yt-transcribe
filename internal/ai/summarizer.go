package ai

import (
	"context"
	"errors"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type LLMSummarizer struct {
	client *openai.Client
	model  string
}

// NewLLMSummarizer creates a new summarizer using OpenAI's LLM API
func NewLLMSummarizer(apiKey, model string) (*LLMSummarizer, error) {
	if apiKey == "" {
		return nil, errors.New("OpenAI API key is required")
	}

	if model == "" {
		return nil, errors.New("model is required")
	}

	// Create OpenAI client with API key
	client := openai.NewClient(option.WithAPIKey(apiKey))

	return &LLMSummarizer{
		client: &client,
		model:  model,
	}, nil
}

func (s *LLMSummarizer) SummarizeText(ctx context.Context, text string) (string, error) {
	if text == "" {
		return "", errors.New("text to summarize cannot be empty")
	}

	// Create summarization request parameters
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("Transcribe the user given video transcription into a concise summary."),
		openai.UserMessage(text),
	}
	params := openai.ChatCompletionNewParams{
		Model:    s.model,
		Messages: messages,
	}

	// Call OpenAI API to summarize the text
	resp, err := s.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
