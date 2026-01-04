package llm

import (
	"context"
	"errors"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const summarizerSystemPrompt = `
You are an expert video content analyzer. When the user provides a video title and transcription, create a comprehensive summary that extracts maximum context, insights and takeaways.
Do not use Markdown formatting, lists or bullet points. Write in a clear, engaging style suitable for a general audience.
Prioritize accuracy over speculation, but make reasonable inferences when context strongly suggests them.`

// Summarizer defines the interface for text summarization services
type Summarizer interface {
	SummarizeText(ctx context.Context, title, text string) (string, error)
}

// NoOpSummarizer is a disabled summarizer that returns empty summaries
type NoOpSummarizer struct{}

func (n *NoOpSummarizer) SummarizeText(ctx context.Context, title, text string) (string, error) {
	return "", nil
}

// OpenAICompatibleSummarizer uses the OpenAI API (compatible with both OpenAI and Ollama)
// AIDEV-NOTE: Ollama supports OpenAI-compatible API at /v1/chat/completions
type OpenAICompatibleSummarizer struct {
	client openai.Client
	model  string
}

// NewSummarizer creates a summarizer based on the provided configuration.
// Returns NoOpSummarizer if endpoint is empty (disabled by default).
// Uses OpenAI-compatible API for both OpenAI and Ollama providers.
func NewSummarizer(endpoint, token, model string) (Summarizer, error) {
	if endpoint == "" {
		return &NoOpSummarizer{}, nil
	}

	if model == "" {
		return nil, errors.New("model is required")
	}

	opts := []option.RequestOption{
		option.WithBaseURL(endpoint),
	}

	// Token is optional for Ollama but required for OpenAI
	if token != "" {
		opts = append(opts, option.WithAPIKey(token))
	}

	client := openai.NewClient(opts...)

	return &OpenAICompatibleSummarizer{
		client: client,
		model:  model,
	}, nil
}

func (s *OpenAICompatibleSummarizer) SummarizeText(ctx context.Context, title, text string) (string, error) {
	userPrompt := `
		Video Title: ` + title + `
		Transcription: ` + text + `
	`

	chatCompletion, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(summarizerSystemPrompt),
			openai.UserMessage(userPrompt),
		},
		Model:       openai.ChatModel(s.model),
		Temperature: openai.Float(1.0),
	})
	if err != nil {
		return "", err
	}

	if len(chatCompletion.Choices) == 0 {
		return "", errors.New("no response from LLM")
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
