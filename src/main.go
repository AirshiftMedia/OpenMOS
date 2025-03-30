package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/db"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/server"
	"airshift/openmos/internal/service"
	"airshift/openmos/pkg/logger"

	"github.com/getsentry/sentry-go"
)

func main() {
	// Initialize standard logger first
	standardLogger := logger.DefaultLogger()
	standardLogger.Info("Starting OpenMOS server...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		standardLogger.Fatalf("Failed to load configuration: %v", err)
	}

	// Configure log level
	logLevel, exists := logger.LevelValues[strings.ToLower(cfg.Logging.Level)]
	if !exists {
		standardLogger.Warningf("Unknown log level: %s. Using 'info' level.", cfg.Logging.Level)
		logLevel = logger.LevelInfo
	}
	standardLogger.SetLevel(logLevel)

	// Configure Sentry if DSN is provided
	var log *logger.SentryLogger
	if cfg.Sentry.DSN != "" {
		sentryConfig := logger.SentryConfig{
			DSN:              cfg.Sentry.DSN,
			Environment:      cfg.Sentry.Environment,
			Release:          cfg.App.Version,
			Debug:            cfg.Sentry.Debug,
			AttachStacktrace: cfg.Sentry.AttachStacktrace,
			SampleRate:       cfg.Sentry.SampleRate,
			TracesSampleRate: cfg.Sentry.TracesSampleRate,
			ServerName:       cfg.App.Name,
		}

		sentryLogger, err := logger.ConfigureSentry(standardLogger, sentryConfig)
		if err != nil {
			standardLogger.Errorf("Failed to configure Sentry: %v, continuing without Sentry integration", err)
			log = logger.NewSentryLogger(standardLogger, cfg.App.Environment, cfg.App.Version)
		} else {
			log = sentryLogger
			log.Info("Sentry integration configured successfully")
		}
	} else {
		log = logger.NewSentryLogger(standardLogger, cfg.App.Environment, cfg.App.Version)
		log.Info("Sentry DSN not provided, continuing without Sentry integration")
	}

	// Set as global logger
	logger.SetGlobalLogger(standardLogger)

	// Set up context for the application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to MongoDB
	log.Info("Connecting to MongoDB...")
	database, err := db.NewMongoDB(cfg)
	if err != nil {
		// Capture the error in Sentry and then log and exit
		log.CaptureException(err, map[string]string{
			"component": "database",
			"action":    "connect",
		}, nil)
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err := database.Close(context.Background()); err != nil {
			log.Errorf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Create repositories
	runningOrderRepo := repository.NewMongoRunningOrderRepository(database)
	storyRepo := repository.NewMongoStoryRepository(database)
	itemRepo := repository.NewMongoItemRepository(database)
	objectRepo := repository.NewMongoObjectRepository(database)

	// Create service
	mosService := service.NewMOSService(runningOrderRepo, storyRepo, itemRepo, objectRepo)

	// Create and start TCP server
	log.Info("Starting TCP server...")
	tcpServer, err := server.NewTCPServer(cfg, mosService)
	if err != nil {
		log.CaptureException(err, map[string]string{
			"component": "server",
			"action":    "start",
		}, nil)
		log.Fatalf("Failed to create TCP server: %v", err)
	}

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	serverSpan := log.StartTransaction("server_lifecycle", "server")
	serverSpan.SetTag("server_address", cfg.GetServerAddress())

	go func() {
		defer serverSpan.Finish()

		if err := tcpServer.Start(ctx); err != nil {
			serverSpan.Status = "internal_error"
			log.CaptureException(err, map[string]string{
				"component": "server",
				"action":    "run",
			}, nil)
			log.Errorf("Server error: %v", err)
			cancel()
		} else {
			serverSpan.Status = "ok"
		}
	}()

	log.Infof("OpenMOS server is running on %s", cfg.GetServerAddress())

	// Wait for shutdown signal
	sig := <-sigCh
	log.Infof("Received signal: %v", sig)

	// Create a context with timeout for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Cancel the server context to start the graceful shutdown
	cancel()

	// Flush Sentry events before exiting
	defer sentry.Flush(2 * time.Second)

	log.Info("Shutdown complete. Goodbye!")
}
