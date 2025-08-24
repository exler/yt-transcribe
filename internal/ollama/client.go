package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Ollama struct {
	httpClient *http.Client
	baseURL    string // e.g. http://127.0.0.1:11434
	model      string // e.g. phi3:mini
}

// NewOllama creates a new Ollama client instance
func NewOllama(baseURL, model string) (*Ollama, error) {
	if model == "" {
		return nil, errors.New("model is required")
	}
	if baseURL == "" {
		return nil, errors.New("baseURL is required")
	}
	return &Ollama{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		baseURL:    baseURL,
		model:      model,
	}, nil
}

// OllamaChatRequest mirrors the /api/chat JSON shape
type OllamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []OllamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
	Options  map[string]any      `json:"options,omitempty"`
}

type OllamaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaChatResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func (s *Ollama) SendPrompt(ctx context.Context, systemPrompt, userPrompt string, temperature float32) (string, error) {
	if userPrompt == "" || systemPrompt == "" {
		return "", errors.New("both prompts are required")
	}

	payload := OllamaChatRequest{
		Model:  s.model,
		Stream: false,
		Messages: []OllamaChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Options: map[string]any{
			"temperature": temperature,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama chat error: status %d, body: %s", resp.StatusCode, string(b))
	}

	var out OllamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Message.Content, nil
}
