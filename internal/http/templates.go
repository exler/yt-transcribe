package http

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/exler/yt-transcribe/internal/queue"
)

// pageData holds the data for the template.
type pageData struct {
	Title                  string
	VideoID                string
	Status                 queue.VideoStatus
	Transcript             string
	Summary                string
	ErrorDetail            string // For general errors
	QueueAddSuccessMessage string
	QueueAddErrorMessage   string
}

func renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	tmpl, err := template.ParseFS(TemplateFiles, fmt.Sprintf("templates/%s.html", templateName))
	if err != nil {
		log.Printf("Error parsing template from FS: %v", err)
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, fmt.Sprintf("%s.html", templateName), data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Template execution error", http.StatusInternalServerError)
	}
}
