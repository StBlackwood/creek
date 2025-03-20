package main

import (
	"creek/internal/config"
	"creek/internal/logger"
	"creek/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration
	err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	cfg := config.Conf

	// Initialize logger
	logger.InitLogger(cfg.LogLevel)

	log := logger.GetLogger()

	// Create and start TCP server
	tcpServer := server.New(cfg.ServerAddress)
	go tcpServer.Start()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Info("Shutting down server...")
	tcpServer.Stop()
	log.Info("Server stopped.")
}
