package logger

import (
	"log"
	"os"
)

// InitLogger sets up logging with a given level
func InitLogger(level string) {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Logger initialized at level:", level)
}
