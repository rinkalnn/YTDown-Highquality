package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	logFile *os.File
	logger  *log.Logger
)

// InitLogger initializes the logger to write to app.log in the current directory
func InitLogger() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeDir := filepath.Dir(exePath)

	// Try to use current working directory first, fallback to executable directory
	cwd, err := os.Getwd()
	if err == nil {
		exeDir = cwd
	}

	// Prefer a per-user config/log directory so packaged app writes reliably
	logDir := GetConfigDir()
	if logDir == "" {
		// fallback to user's home directory
		if home, herr := os.UserHomeDir(); herr == nil {
			logDir = filepath.Join(home, ".ytdown")
		} else {
			// as a last resort, use exeDir
			logDir = exeDir
		}
	}

	_ = os.MkdirAll(logDir, 0755)
	logFilePath := filepath.Join(logDir, "app.log")

	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logFile = f
	// Write to both file and stdout so we capture logs for packaged and dev runs
	mw := io.MultiWriter(f, os.Stdout)
	logger = log.New(mw, "", 0) // We'll add our own timestamp format

	logger.Println(formatLog("INFO", "--- Logger Initialized ---"))
	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		logger.Println(formatLog("INFO", "--- Logger Closed ---"))
		logFile.Close()
	}
}

func formatLog(level, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)
}

// LogInfo logs an informational message
func LogInfo(format string, v ...interface{}) {
	if logger != nil {
		msg := fmt.Sprintf(format, v...)
		logger.Println(formatLog("INFO", msg))
	}
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	if logger != nil {
		msg := fmt.Sprintf(format, v...)
		logger.Println(formatLog("ERROR", msg))
	}
}

// LogDebug logs a debug message
func LogDebug(format string, v ...interface{}) {
	if logger != nil {
		msg := fmt.Sprintf(format, v...)
		logger.Println(formatLog("DEBUG", msg))
	}
}
