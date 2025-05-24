package cmd

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/exler/yt-transcribe/static"
	"github.com/exler/yt-transcribe/templates"
)

// setupTestServer configures and returns a new httptest.Server.
// It replicates the essential routing from runserverCmd for serving templates and static files.
func setupTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Handler for the main page (simplified, only needs to render the template)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// For this test, we only need to ensure the template is served.
			// No need to replicate the full postHandler or pageData logic.
			tmpl, err := templates.Files.ReadFile("index.html")
			if err != nil {
				http.Error(w, "Error reading template: "+err.Error(), http.StatusInternalServerError)
				return
			}
			// A very basic render, assuming index.html is a full page.
			// For more complex templates, template.ParseFS and ExecuteTemplate would be needed.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(tmpl)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Handler for static files
	staticFS := http.FS(static.Files)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(staticFS)))

	return httptest.NewServer(mux)
}

func TestServerServesEmbeddedFiles(t *testing.T) {
	server := setupTestServer()
	defer server.Close()

	// Test 1: Root path ("/")
	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to GET root path: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d for root path, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body for root path: %v", err)
	}

	// Check for some content from index.html
	// This needs to be a stable string within your templates/index.html
	expectedContent := "<title>yt-transcript</title>"
	if !strings.Contains(string(body), expectedContent) {
		t.Errorf("Expected root path response to contain '%s', but it didn't. Body: %s", expectedContent, string(body))
	}

	// Test 2: Static file path ("/static/style.css")
	respCss, err := http.Get(server.URL + "/static/style.css")
	if err != nil {
		t.Fatalf("Failed to GET /static/style.css: %v", err)
	}
	defer respCss.Body.Close()

	if respCss.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d for /static/style.css, got %d", http.StatusOK, respCss.StatusCode)
	}

	expectedContentType := "text/css; charset=utf-8"
	contentType := respCss.Header.Get("Content-Type")
	if contentType != expectedContentType {
		t.Errorf("Expected Content-Type '%s' for /static/style.css, got '%s'", expectedContentType, contentType)
	}
}
