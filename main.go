package main

import (
	"log"

	"github.com/exler/yt-transcribe/cmd"
)

func main() {
	if err := cmd.Run(); //#nosec
	err != nil {
		log.Fatalf("Error running command: %v", err)
	}
}
