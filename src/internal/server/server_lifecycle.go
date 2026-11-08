package server

import (
	"airshift/openmos/pkg/logger"
	"context"
	"fmt"
	"net"
	"time"
)

// Start begins accepting connections
func (s *TCPServer) Start(ctx context.Context) error {
	defer s.wg.Done()
	s.wg.Add(1)

	address := s.listener.Addr().String()
	logger.Infof("Server listening on %s", address)

	// Accept connections in a loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.shutdownCh:
				return
			default:
				// Set accept timeout so we can check for shutdown
				tcpListener, ok := s.listener.(*net.TCPListener)
				if ok {
					tcpListener.SetDeadline(time.Now().Add(time.Second))
				}

				conn, err := s.listener.Accept()
				if err != nil {
					if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
						// This is just a timeout from our deadline, continue
						continue
					}
					logger.Errorf("Error accepting connection: %v", err)
					continue
				}

				// Create new client connection
				client := NewClientConnection(conn, s, s.config)

				// Register client
				s.registerClient(client)

				// Handle client in a goroutine
				s.wg.Add(1)
				go func() {
					defer s.wg.Done()
					client.Start(ctx)
				}()
			}
		}
	}()

	<-ctx.Done()
	return s.Shutdown(context.Background())
}

// Shutdown gracefully shuts down the server
func (s *TCPServer) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down server...")

	// Signal all goroutines to stop
	close(s.shutdownCh)

	// Close listener
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all client connections
	s.clientsMu.Lock()
	for _, client := range s.clients {
		client.Close()
	}
	s.clientsMu.Unlock()

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
