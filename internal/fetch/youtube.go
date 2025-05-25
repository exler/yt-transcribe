package fetch

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const metadataPrintTemplate = "%(id)s;%(title)s;%(duration_string)s;%(upload_date)s"

// VideoMetadata holds metadata for a downloaded YouTube video
type VideoMetadata struct {
	Title         string
	Duration      string // e.g., "10:35"
	UploadDate    string // e.g., "20231026"
	AudioFilePath string
	VideoID       string
}

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
// and returns its metadata.
func (d *YouTubeDownloader) DownloadAudio(videoURL string, options ...string) (VideoMetadata, error) {
	metadata := VideoMetadata{}

	if err := d.CheckYTDLP(); err != nil {
		return metadata, err
	}

	// Create a unique filename based on the video ID
	outputTemplate := filepath.Join(d.OutputDir, "%(id)s.%(ext)s")

	args := []string{
		"--format", "bestaudio/best",
		"--output", outputTemplate,
		"--no-playlist",
		"--no-simulate",
		"--quiet",
		"--print", metadataPrintTemplate,
		"--print", "filename",
	}
	args = append(args, options...)
	args = append(args, videoURL)

	cmd := exec.Command("yt-dlp", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return metadata, fmt.Errorf("failed to download video: %w\nStderr: %s", err, stderr.String())
	}

	outputLines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(outputLines) < 2 {
		return metadata, fmt.Errorf("yt-dlp output did not contain enough lines for metadata and filename. Output: %s", stdout.String())
	}

	// The first line is metadata and the second is filename.
	// As multiple --print args are used, they are printed in the order they appear in the command.
	metadataLine := outputLines[0]
	audioFilePath := outputLines[1]

	parts := strings.SplitN(metadataLine, ";", 4)
	if len(parts) < 4 {
		return metadata, fmt.Errorf("failed to parse metadata output: expected 4 parts, got %d from '%s'", len(parts), metadataLine)
	}

	metadata.VideoID = parts[0]
	metadata.Title = parts[1]
	metadata.Duration = parts[2]
	metadata.UploadDate = parts[3]
	metadata.AudioFilePath = audioFilePath

	if metadata.VideoID == "" {
		return metadata, fmt.Errorf("extracted VideoID is empty. Raw metadata line: '%s'", metadataLine)
	}

	// Return the populated metadata
	return metadata, nil
}

// GetVideoMetadata fetches metadata for a YouTube video without downloading the video.
func (d *YouTubeDownloader) GetVideoMetadata(videoURL string) (VideoMetadata, error) {
	metadata := VideoMetadata{}

	if err := d.CheckYTDLP(); err != nil {
		return metadata, err
	}

	args := []string{
		"--skip-download", // Do not download the video
		"--print", metadataPrintTemplate,
		"--no-playlist",
		"--no-simulate", // Ensure it processes the URL
		"--quiet",
		videoURL,
	}

	cmd := exec.Command("yt-dlp", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return metadata, fmt.Errorf("failed to fetch video metadata: %w\nStderr: %s", err, stderr.String())
	}

	outputStr := strings.TrimSpace(stdout.String())
	if outputStr == "" {
		return metadata, fmt.Errorf("yt-dlp did not return any metadata output")
	}

	// Expecting a single line of output
	parts := strings.SplitN(outputStr, ";", 4)
	if len(parts) < 4 {
		return metadata, fmt.Errorf("failed to parse metadata output: expected 4 parts, got %d from '%s'", len(parts), outputStr)
	}

	metadata.VideoID = parts[0]
	metadata.Title = parts[1]
	metadata.Duration = parts[2]
	metadata.UploadDate = parts[3]
	metadata.AudioFilePath = ""

	if metadata.VideoID == "" {
		return metadata, fmt.Errorf("extracted VideoID is empty from metadata. Raw output: '%s'", outputStr)
	}

	return metadata, nil
}
