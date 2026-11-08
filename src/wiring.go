package main

import (
	"context"
	"time"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/events"
	"airshift/openmos/internal/repository"
	"airshift/openmos/internal/repository/memory"
	"airshift/openmos/internal/server"
	"airshift/openmos/internal/service"
	"airshift/openmos/pkg/logger"

	"github.com/getsentry/sentry-go"
)

// newRepository returns the storage the server runs on. Until the PostgreSQL
// store lands (MOS-14, wired in MOS-16) this is the in-memory store, so all
// data is lost on restart.
func newRepository(log *logger.SentryLogger) repository.Repository {
	log.Warning("Using in-memory storage: data is not persisted")
	return memory.New()
}

// newTCPServer wires repositories, the event bus and the MOS service into a
// TCP server, and exits the process if the server cannot be created.
func newTCPServer(cfg *config.Config, repo repository.Repository, log *logger.SentryLogger) *server.TCPServer {
	// Create event bus for pub-sub messaging
	eventBus := events.NewEventBus()

	// Create service
	mosService := service.NewMOSService(repo.RunningOrders(), repo.Stories(), repo.Items(), repo.Objects(), eventBus)

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
