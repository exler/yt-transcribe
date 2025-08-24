package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// FFMPEG wraps the ffmpeg command-line tool.
// It provides methods to manipulate audio files.
type FFMPEG struct{}

// NewFFMPEG creates a new FFMPEG instance
func NewFFMPEG() (*FFMPEG, error) {
	return &FFMPEG{}, nil
}

// CheckFFMPEG verifies that ffmpeg is installed
func (f *FFMPEG) CheckFFMPEG() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found: %w", err)
	}
	return nil
}

// GetFFMPEGVersion retrieves the version of the ffmpeg package
func (f *FFMPEG) GetFFMPEGVersion() (string, error) {
	cmd := exec.Command("ffmpeg", "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ffmpeg not found: %w", err)
	}

	// Parse first line to extract version info
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 && strings.Contains(lines[0], "ffmpeg version") {
		// Extract version from line like "ffmpeg version 4.4.2-0ubuntu0.22.04.1"
		parts := strings.Fields(lines[0])
		if len(parts) >= 3 {
			return parts[2], nil
		}
	}

	return "unknown", nil
}

// FFmpeg whisper filter integration. Requires ffmpeg built with --enable-whisper (FFmpeg 8+)
// TranscribeWithWhisperFilter runs the FFmpeg 'whisper' audio filter and returns the transcription text.
//
// Reference: https://ffmpeg.org/ffmpeg-filters.html#whisper-1
func (f *FFMPEG) TranscribeWithWhisperFilter(inputFile, modelPath, language string, queue int) (string, error) {
	if err := f.CheckFFMPEG(); err != nil {
		return "", err
	}

	// Create a temporary destination file to collect the transcription output from the filter
	tmpFile, err := os.CreateTemp("", "ffmpeg-whisper-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file for transcription: %w", err)
	}
	destPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(destPath)

	// Build the whisper filter string
	// Example: whisper=model=/models/ggml-small.bin:language=auto:queue=15:destination=/tmp/out.txt:format=text
	filter := fmt.Sprintf("whisper=model=%s:language=%s:queue=%d:destination=%s:format=srt", modelPath, language, queue, destPath)

	// Run ffmpeg to process audio only (-vn) and write null output while the filter writes to destination
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-vn", "-af", filter, "-f", "null", "-", "-y")
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to run ffmpeg whisper filter: %w\nffmpeg output: %s", err, string(out))
	}

	// Read the transcription text
	data, err := os.ReadFile(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to read transcription output: %w", err)
	}
	return string(data), nil
}
