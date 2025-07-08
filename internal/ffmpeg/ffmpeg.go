package ffmpeg

import (
	"fmt"
	"os/exec"
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

// SpeedUpAudio speeds up the audio file and converts it to a low-bitrate MP3
func (f *FFMPEG) SpeedUpAudio(inputFile, outputFile string, speed float64) error {
	if err := f.CheckFFMPEG(); err != nil {
		return err
	}

	// Use ffmpeg to speed up the audio and convert to MP3
	// -i: input file
	// -filter:a: apply audio filter. "atempo=x" speeds up the audio by a factor of x
	// -ac 1: set audio channels to mono (1 channel)
	// -ab 64k: set audio bitrate to 64 kbps
	// -y: overwrite output file if it exists
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-filter:a", fmt.Sprintf("atempo=%.2f", speed), "-ac", "1", "-ab", "64k", "-y", outputFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run ffmpeg command: %w", err)
	}

	return nil
}
