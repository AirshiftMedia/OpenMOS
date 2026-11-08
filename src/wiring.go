package main

import (
	"context"
	"time"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/db"
	"airshift/openmos/internal/events"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/server"
	"airshift/openmos/internal/service"
	"airshift/openmos/pkg/logger"

	"github.com/getsentry/sentry-go"
)

// connectDatabase connects to MongoDB and exits the process on failure.
func connectDatabase(cfg *config.Config, log *logger.SentryLogger) *db.MongoDB {
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
	return database
}

// newTCPServer wires repositories, the event bus and the MOS service into a
// TCP server, and exits the process if the server cannot be created.
func newTCPServer(cfg *config.Config, database *db.MongoDB, log *logger.SentryLogger) *server.TCPServer {
	// Create repositories
	runningOrderRepo := repository.NewMongoRunningOrderRepository(database)
	storyRepo := repository.NewMongoStoryRepository(database)
	itemRepo := repository.NewMongoItemRepository(database)
	objectRepo := repository.NewMongoObjectRepository(database)

	// Create event bus for pub-sub messaging
	eventBus := events.NewEventBus()

	// Create service
	mosService := service.NewMOSService(runningOrderRepo, storyRepo, itemRepo, objectRepo, eventBus)

	// Create and start TCP server
	log.Info("Starting TCP server...")
	tcpServer, err := server.NewTCPServer(cfg, mosService, eventBus)
	if err != nil {
		log.CaptureException(err, map[string]string{
			"component": "server",
			"action":    "start",
		}, nil)
		log.Fatalf("Failed to create TCP server: %v", err)
	}
	return tcpServer
}

// startServer runs the TCP server in a goroutine inside a Sentry
// transaction; a server error cancels the application context. The returned
// channel is closed once the server has stopped, including its shutdown.
func startServer(ctx context.Context, cancel context.CancelFunc, cfg *config.Config,
	tcpServer *server.TCPServer, log *logger.SentryLogger) <-chan struct{} {
	// Start server monitoring
	serverSpan := log.StartTransaction("server_lifecycle", "server")
	serverSpan.SetTag("server_address", cfg.GetServerAddress())

	done := make(chan struct{})

	// Start server in a goroutine
	go func() {
		defer close(done)
		defer serverSpan.Finish()

		if err := tcpServer.Start(ctx); err != nil {
			// Set error status on span
			serverSpan.Status = sentry.SpanStatusInternalError

			// Log and capture the error
			log.CaptureException(err, map[string]string{
				"component": "server",
				"action":    "run",
			}, nil)
			log.Errorf("Server error: %v", err)
			cancel()
		}
	}()

	return done
}

// waitForServer waits for the server to finish shutting down. Shutdown is
// bounded by ShutdownTimeout; the extra grace only guards against a hang.
func waitForServer(done <-chan struct{}, cfg *config.Config, log *logger.SentryLogger) {
	select {
	case <-done:
	case <-time.After(cfg.Server.ShutdownTimeout + 5*time.Second):
		log.Errorf("Server did not stop within %v, exiting anyway", cfg.Server.ShutdownTimeout)
	}
}
