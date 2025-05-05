package fetch

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// YouTubeDownloader manages the downloading of YouTube videos
type YouTubeDownloader struct {
	OutputDir string
}

// NewYouTubeDownloader creates a new video downloader instance
func NewYouTubeDownloader(outputDir string) (*YouTubeDownloader, error) {
	// Get absolute path for output directory
	absPath, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	return &YouTubeDownloader{
		OutputDir: absPath,
	}, nil
}

// CheckYTDLP verifies that yt-dlp is installed
func (d *YouTubeDownloader) CheckYTDLP() error {
	cmd := exec.Command("yt-dlp", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp not found: %w", err)
	}
	return nil
}

// DownloadAudio downloads a YouTube video (converted to audio format) using yt-dlp
func (d *YouTubeDownloader) DownloadAudio(videoURL string, options ...string) (string, error) {
	if err := d.CheckYTDLP(); err != nil {
		return "", err
	}

	// Create a unique filename based on the video ID
	outputTemplate := filepath.Join(d.OutputDir, "%(id)s.%(ext)s")

	// Base arguments for yt-dlp
	args := []string{
		"--format", "bestaudio/best",
		"--output", outputTemplate,
		"--no-playlist",
		"--no-simulate",
		"--quiet",
		"--print", "filename",
	}

	// Add any additional options
	args = append(args, options...)

	// Add the video URL
	args = append(args, videoURL)

	// Execute yt-dlp command
	cmd := exec.Command("yt-dlp", args...)
	// Capture stdout (filename) and send stderr to os.Stderr for errors
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	filename := strings.TrimSpace(stdout.String())
	if filename == "" {
		return "", fmt.Errorf("could not determine output filename")
	}

	// Return the full path to the downloaded file
	return filename, nil
}
