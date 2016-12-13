package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/cloudposse/github-authorized-keys/cmd"
	"os"
)

func main() {
	LoggerInit()
	cmd.Execute()
}

// LoggerInit - Initialize logger configuration used for cli
func LoggerInit() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}