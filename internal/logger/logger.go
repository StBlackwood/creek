package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

// Logger is a singleton instance of Logrus
var (
	log  *logrus.Logger
	once sync.Once
)

// InitLogger initializes the logger with a given log level
func InitLogger(level string) {
	once.Do(func() {
		log = logrus.New()
		log.SetOutput(os.Stdout)
		log.SetFormatter(&logrus.JSONFormatter{}) // Structured JSON output

		// Set log level based on input
		switch level {
		case "debug":
			log.SetLevel(logrus.DebugLevel)
		case "info":
			log.SetLevel(logrus.InfoLevel)
		case "warn":
			log.SetLevel(logrus.WarnLevel)
		case "error":
			log.SetLevel(logrus.ErrorLevel)
		default:
			log.SetLevel(logrus.InfoLevel) // Default to INFO
		}
	})
}

// GetLogger returns the global logger instance
func GetLogger() *logrus.Logger {
	if log == nil {
		InitLogger("info") // Default initialization
	}
	return log
}
