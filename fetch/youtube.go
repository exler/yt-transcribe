package fetch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VideoDownloader manages the downloading of YouTube videos
type VideoDownloader struct {
	OutputDir string
}

// NewVideoDownloader creates a new video downloader instance
func NewVideoDownloader(outputDir string) (*VideoDownloader, error) {
	// Create output directory if it doesn't exist
	if outputDir == "" {
		outputDir = "downloads"
	}

	absPath, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &VideoDownloader{
		OutputDir: absPath,
	}, nil
}

// CheckYTDLP verifies that yt-dlp is installed
func (d *VideoDownloader) CheckYTDLP() error {
	cmd := exec.Command("yt-dlp", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp not found: %w", err)
	}
	return nil
}

// DownloadVideo downloads a YouTube video using yt-dlp
func (d *VideoDownloader) DownloadVideo(videoURL string, options ...string) (string, error) {
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
		"--quiet",
	}

	// Add any additional options
	args = append(args, options...)

	// Add the video URL
	args = append(args, videoURL)

	// Execute yt-dlp command
	cmd := exec.Command("yt-dlp", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	// Return the directory where the video was saved
	return d.OutputDir, nil
}
