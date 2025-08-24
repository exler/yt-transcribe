package ollama

import (
	"context"
)

const summarizerSystemPrompt = `
You are an expert video content analyzer. When the user provides a video title and transcription, create a comprehensive summary that extracts maximum context, insights and takeaways.
Do not use Markdown formatting, lists or bullet points. Write in a clear, engaging style suitable for a general audience.
Prioritize accuracy over speculation, but make reasonable inferences when context strongly suggests them.`

func (s *Ollama) SummarizeText(ctx context.Context, title, text string) (string, error) {
	userPrompt := `
		Video Title: ` + title + `
		Transcription: ` + text + `
	`

	return s.SendPrompt(ctx, summarizerSystemPrompt, userPrompt, 0.3)
}
