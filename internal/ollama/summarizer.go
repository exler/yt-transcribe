package ollama

import (
	"context"
)

const summarizerSystemPrompt = `
You are an expert video content analyzer. When the user provides a video title and transcription, create a comprehensive summary that extracts maximum context and insights.

Analyze the content deeply and provide:
   - Core message and key points
   - Overall narrative arc or argument structure
   - Primary takeaways
   
Format your response with whitespaces but avoid using any special characters. Prioritize accuracy over speculation, but make reasonable inferences when context strongly suggests them.`

func (s *Ollama) SummarizeText(ctx context.Context, title, text string) (string, error) {
	userPrompt := `
		Video Title: ` + title + `
		Transcription: ` + text + `
	`

	return s.SendPrompt(ctx, summarizerSystemPrompt, userPrompt, 0.3)
}
