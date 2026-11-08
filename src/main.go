package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"airshift/openmos/internal/config"
	"airshift/openmos/pkg/logger"

	"github.com/getsentry/sentry-go"
)

func main() {
	// Define command-line flags
	generateConfig := flag.String("generate-config", "", "Generate a default configuration file at the specified path and exit")
	configPath := flag.String("config", "", "Path to the configuration file (default: search for config.yaml)")

	// Parse flags
	flag.Parse()

	// Handle config generation if requested
	if *generateConfig != "" {
		generateDefaultConfig(*generateConfig)
		return
	}

	// Set config file path if provided
	if *configPath != "" {
		os.Setenv("CONFIG_FILE", *configPath)
	}

	// Initialize standard logger
	standardLogger := logger.DefaultLogger()
	standardLogger.Info("Starting OpenMOS server...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		standardLogger.Fatalf("Failed to load configuration: %v", err)
	}

	log := setupLogging(cfg, standardLogger)

	// Set up context for the application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database := connectDatabase(cfg, log)
	defer func() {
		if err := database.Close(context.Background()); err != nil {
			log.Errorf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	tcpServer := newTCPServer(cfg, database, log)

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	startServer(ctx, cancel, cfg, tcpServer, log)

	log.Infof("OpenMOS server is running on %s", cfg.GetServerAddress())

	// Wait for shutdown signal
	sig := <-sigCh
	log.Infof("Received signal: %v", sig)

	// Cancel the server context to start the graceful shutdown
	cancel()

	// Flush Sentry events before exiting
	defer sentry.Flush(2 * time.Second)

	log.Info("Shutdown complete. Goodbye!")
}
