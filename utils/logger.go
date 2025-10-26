package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	AuthLogger  *log.Logger
)

func InitLogger() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Open log files
	infoFile, err := os.OpenFile("logs/info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open info log file:", err)
	}

	errorFile, err := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open error log file:", err)
	}

	authFile, err := os.OpenFile("logs/auth.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open auth log file:", err)
	}

	// Create loggers that write to both files and console
	InfoLogger = log.New(io.MultiWriter(infoFile, os.Stdout), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(io.MultiWriter(errorFile, os.Stderr), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	AuthLogger = log.New(io.MultiWriter(authFile, os.Stdout), "AUTH: ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger.Println("Logger initialized successfully")
}

func LogInfo(message string) {
	InfoLogger.Println(message)
}

func LogError(message string) {
	ErrorLogger.Println(message)
}

func LogAuth(event, username, ip string, success bool) {
	status := "FAILED"
	if success {
		status = "SUCCESS"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s - User: %s, IP: %s, Status: %s",
		timestamp, event, username, ip, status)

	AuthLogger.Println(logMessage)
}

func LogSignup(username, email, ip string, success bool) {
	LogAuth("SIGNUP", username, ip, success)
}

func LogLogin(username, ip string, success bool) {
	LogAuth("LOGIN", username, ip, success)
}

func LogLogout(username, ip string) {
	LogAuth("LOGOUT", username, ip, true)
}

// TestLogging - Function to test if logging is working
func TestLogging() {
	LogInfo("Testing info logging...")
	LogError("Testing error logging...")
	LogAuth("TEST", "testuser", "127.0.0.1", true)
}
