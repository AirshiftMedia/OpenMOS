package server

import (
	"fmt"
	"net"
	"sync"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/events"
	"airshift/openmos/internal/service"
)

// TCPServer represents the TCP socket server
type TCPServer struct {
	listener   net.Listener
	clients    map[string]*ClientConnection
	clientsMu  sync.RWMutex
	service    *service.MOSService
	config     *config.Config
	eventBus   *events.EventBus
	wg         sync.WaitGroup // accept loop and client goroutines
	shutdownCh chan struct{}

	shutdownOnce sync.Once
	shutdownErr  error
}

// NewTCPServer creates a new TCP server instance
func NewTCPServer(cfg *config.Config, mosService *service.MOSService, eventBus *events.EventBus) (*TCPServer, error) {
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
		eventBus:   eventBus,
		shutdownCh: make(chan struct{}),
	}

	return server, nil
}
