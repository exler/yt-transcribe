package fetch

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

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
	metadataPrintTemplate := "%(id)s;%(title)s;%(duration_string)s;%(upload_date)s"

	// Base arguments for yt-dlp
	args := []string{
		"--format", "bestaudio/best",
		"--output", outputTemplate,
		"--no-playlist",
		"--no-simulate",
		"--quiet",
		"--print", metadataPrintTemplate, // Print metadata to stdout
		// Ensure filename is printed separately, for example, to a temporary file or stderr if yt-dlp supports it.
		// For simplicity, we'll reconstruct the filename based on VideoID and extension later,
		// or rely on a known extension like .opus or .m4a if format is fixed.
		// A more robust way would be to get the final filename from yt-dlp if possible without mixing with metadata.
	}

	// Add any additional options
	args = append(args, options...)

	// Add the video URL
	args = append(args, videoURL)

	// Execute yt-dlp command
	cmd := exec.Command("yt-dlp", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer // Capture stderr for more detailed error messages
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return metadata, fmt.Errorf("failed to download video: %w\nStderr: %s", err, stderr.String())
	}

	// yt-dlp prints the metadata string to stdout then the filename.
	// We expect the metadata string first, then potentially other output if not careful.
	// The --print option output appears after download/processing.
	outputStr := strings.TrimSpace(stdout.String())
	if outputStr == "" {
		return metadata, fmt.Errorf("yt-dlp did not return any output")
	}

	// Assuming the last line of output is the metadata string.
	// This might need adjustment if yt-dlp outputs other info after the metadata.
	lines := strings.Split(outputStr, "\n")
	metadataLine := lines[len(lines)-1] // Assume metadata is the last line printed by --print

	parts := strings.SplitN(metadataLine, ";", 4)
	if len(parts) < 4 {
		return metadata, fmt.Errorf("failed to parse metadata output: expected 4 parts, got %d from '%s'", len(parts), metadataLine)
	}

	metadata.VideoID = parts[0]
	metadata.Title = parts[1]
	metadata.Duration = parts[2]
	metadata.UploadDate = parts[3]

	// Construct the audio file path. This assumes a fixed extension or requires knowledge of it.
	// yt-dlp by default might use .opus, .m4a, etc.
	// For now, let's assume we know the extension or can determine it.
	// A common audio extension used by yt-dlp for 'bestaudio' is often .m4a or .opus.
	// This is a simplification. A more robust solution involves getting the exact filename from yt-dlp.
	// One way is to use `--print filename` as before, but parse it separately.
	// Or, if we fix the audio format, e.g., to m4a:
	// args = append(args, "--extract-audio", "--audio-format", "m4a")
	// Then the extension would be known.
	// For now, let's try to get the filename using a second command, or by finding it.
	// This is inefficient. Let's try to get it from the output template if possible.

	// Re-evaluating filename retrieval:
	// The original command had `--print filename`. If we add that back, stdout will have:
	// 1. Metadata string
	// 2. Filename
	// We need to parse this carefully.

	// Let's refine the command and parsing logic.
	// We will use one call. --print for metadata, and --print filename for the filepath.
	// The order of these print statements in stdout needs to be handled.
	// Typically, --print "X" prints X then a newline. If multiple --print are used, they print in order.

	// New approach for args:
	args = []string{
		"--format", "bestaudio/best",
		"--output", outputTemplate, // yt-dlp will replace %(id)s and %(ext)s
		"--no-playlist",
		"--no-simulate",
		"--quiet",
		"--print", metadataPrintTemplate, // Metadata line
		"--print", "filename", // Filename line
	}
	args = append(args, options...)
	args = append(args, videoURL)

	cmd = exec.Command("yt-dlp", args...)
	stdout.Reset() // Reset buffer
	stderr.Reset() // Reset buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return metadata, fmt.Errorf("failed to download video (รอบสอง): %w\nStderr: %s", err, stderr.String())
	}

	outputLines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(outputLines) < 2 {
		return metadata, fmt.Errorf("yt-dlp output did not contain enough lines for metadata and filename. Output: %s", stdout.String())
	}

	// Assuming the first line is metadata and the second is filename.
	// This depends on the order yt-dlp processes --print flags.
	// Based on documentation and common behavior, --print flags are processed after the download.
	// If multiple --print are used, they are printed in the order they appear in the command.
	metadataLine = outputLines[0]
	audioFilePath := outputLines[1]

	parts = strings.SplitN(metadataLine, ";", 4)
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
	if !strings.HasPrefix(filepath.Base(metadata.AudioFilePath), metadata.VideoID) {
		// Potentially, yt-dlp might use a different ID in filename vs metadata (e.g. if ID comes from chapter, not video)
		// However, for simple video downloads, ID in filename should match the video ID.
		// We should ensure outputTemplate uses %(id)s to be sure.
		// The current outputTemplate `%(id)s.%(ext)s` should ensure this.
		// This check is a sanity check.
		return metadata, fmt.Errorf("filename's video ID '%s' does not seem to match metadata VideoID '%s'. Path: '%s'", filepath.Base(metadata.AudioFilePath), metadata.VideoID, metadata.AudioFilePath)
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

	metadataPrintTemplate := "%(id)s;%(title)s;%(duration_string)s;%(upload_date)s"

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
	metadata.AudioFilePath = "" // Explicitly empty as no file is downloaded

	if metadata.VideoID == "" {
		return metadata, fmt.Errorf("extracted VideoID is empty from metadata. Raw output: '%s'", outputStr)
	}

	return metadata, nil
}
