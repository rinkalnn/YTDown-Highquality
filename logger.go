package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

// InitLogger is kept for compatibility but doesn't do much since we log to stdout now
func InitLogger() error {
	return nil
}

// CloseLogger is kept for compatibility
func CloseLogger() {
}

func formatLog(level, message string) string {
	timestamp := time.Now().Format("15:04:05")
	return fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)
}

// LogInfo logs an informational message
func LogInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logger.Println(formatLog("INFO", msg))
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logger.Println(formatLog("ERROR", msg))
}

// LogWarning logs a warning message
func LogWarning(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logger.Println(formatLog("WARNING", msg))
}

// LogDebug logs a debug message
func LogDebug(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	logger.Println(formatLog("DEBUG", msg))
}
