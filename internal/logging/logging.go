package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const logFile = "app.log"
const oldLogsDir = "old_logs"

func SetupLogging() {
	// Ensure old_logs directory exists
	if _, err := os.Stat(oldLogsDir); os.IsNotExist(err) {
		os.Mkdir(oldLogsDir, 0755)
	}

	// Rotate previous log file if it exists
	if _, err := os.Stat(logFile); err == nil {
		backupName := filepath.Join(oldLogsDir, fmt.Sprintf("app_%s.log", time.Now().Format("20060102_150405")))
		os.Rename(logFile, backupName)
	}

	// Create new log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}
