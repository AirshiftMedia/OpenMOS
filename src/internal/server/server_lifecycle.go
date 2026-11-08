package server

import (
	"context"
	"fmt"

	"airshift/openmos/pkg/logger"
)

// Start accepts connections until ctx is cancelled, then shuts the server
// down and returns the result of Shutdown.
func (s *TCPServer) Start(ctx context.Context) error {
	address := s.listener.Addr().String()
	logger.Infof("Server listening on %s", address)

	// Accept connections in a loop. Start itself is not counted in the wait
	// group, because Shutdown below waits on it.
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.acceptLoop(ctx)
	}()

	<-ctx.Done()
	return s.Shutdown(context.Background())
}

// Shutdown gracefully shuts down the server: it stops accepting, closes all
// client connections and waits for their goroutines. It is safe to call more
// than once; later calls return the first call's result.
func (s *TCPServer) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		s.shutdownErr = s.shutdown(ctx)
	})
	return s.shutdownErr
}

func (s *TCPServer) shutdown(ctx context.Context) error {
	logger.Info("Shutting down server...")

	// Signal all goroutines to stop
	close(s.shutdownCh)

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all client connections. Copy them first: Close unregisters the
	// client, which takes clientsMu, so it must not be called under the lock.
	s.clientsMu.RLock()
	clients := make([]*ClientConnection, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	s.clientsMu.RUnlock()

	for _, client := range clients {
		client.Close()
	}

	// Wait for all goroutines to finish with a timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.Server.ShutdownTimeout)
	defer cancel()

	// Create a channel to signal when all goroutines have finished
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	// Wait for either the context to be canceled or all goroutines to finish
	select {
	case <-shutdownCtx.Done():
		return fmt.Errorf("server shutdown timed out")
	case <-done:
		logger.Info("Server shutdown complete")
		return nil
	}
}
