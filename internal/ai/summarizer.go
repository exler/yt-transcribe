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

const summarizerSystemPrompt = `
You are an expert video content analyzer. When the user provides a video title and transcription, create a comprehensive summary that extracts maximum context and insights.

**Your Task:**
Analyze the content deeply and provide:

1. **Main Summary** (2-3 paragraphs)
   - Core message and key points
   - Overall narrative arc or argument structure
   - Primary takeaways

2. **Key Information Extraction**
   - Main topics/themes discussed
   - Important facts, statistics, or claims made
   - Any references to people, places, events, or dates
   - Technical terms or specialized vocabulary used

3. **Inferred Details**
   - Speaker(s) identity/role (if apparent from context)
   - Setting/location (if mentioned or implied)
   - Time period or currency of information
   - Any visual elements likely present (demonstrations, slides, etc.)

4. **Practical Value**
   - Actionable insights or advice given
   - Problems addressed and solutions offered
   - Skills or knowledge viewers would gain

5. **Critical Notes**
   - Any potential biases or perspectives presented
   - Claims that might need fact-checking
   - Missing context or incomplete explanations

Format your response with clear headers and bullet points for easy scanning. Prioritize accuracy over speculation, but make reasonable inferences when context strongly suggests them.`

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

func (s *LLMSummarizer) SummarizeText(ctx context.Context, title, text string) (string, error) {
	if text == "" {
		return "", errors.New("text to summarize cannot be empty")
	}

	// Replace placeholders in the prompt with actual title and text
	userPrompt := `
		**Video Title:** ` + title + `
		**Transcription:** ` + text + `
	`

	// Create summarization request parameters
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(summarizerSystemPrompt),
		openai.UserMessage(userPrompt),
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
