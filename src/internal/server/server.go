package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/service"
	"airshift/openmos/pkg/logger"
)

// TCPServer represents the TCP socket server
type TCPServer struct {
	listener   net.Listener
	clients    map[string]*ClientConnection
	clientsMu  sync.RWMutex
	service    *service.MOSService
	config     *config.Config
	wg         sync.WaitGroup
	shutdownCh chan struct{}
}

// NewTCPServer creates a new TCP server instance
func NewTCPServer(cfg *config.Config, mosService *service.MOSService) (*TCPServer, error) {
	address := cfg.GetServerAddress()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener on %s: %w", address, err)
	}

	server := &TCPServer{
		listener:   listener,
		clients:    make(map[string]*ClientConnection),
		service:    mosService,
		config:     cfg,
		shutdownCh: make(chan struct{}),
	}

	return server, nil
}

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
					log.Printf("Error accepting connection: %v", err)
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
	log.Println("Shutting down server...")

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
		log.Println("Server shutdown complete")
		return nil
	}
}

// registerClient registers a client connection
func (s *TCPServer) registerClient(client *ClientConnection) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	s.clients[client.ID()] = client
	log.Printf("Client registered: %s", client.ID())
}

// unregisterClient removes a client connection
func (s *TCPServer) unregisterClient(clientID string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	delete(s.clients, clientID)
	log.Printf("Client unregistered: %s", clientID)
}

// GetClient returns a client connection by ID
func (s *TCPServer) GetClient(clientID string) (*ClientConnection, bool) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	client, ok := s.clients[clientID]
	return client, ok
}

// BroadcastMessage sends a message to all connected clients
func (s *TCPServer) BroadcastMessage(data []byte) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	for _, client := range s.clients {
		// Send in a non-blocking way to avoid one slow client affecting others
		go func(c *ClientConnection) {
			if err := c.Write(data); err != nil {
				log.Printf("Error broadcasting to client %s: %v", c.ID(), err)
			}
		}(client)
	}
}
