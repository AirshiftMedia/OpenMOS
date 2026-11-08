package server

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"airshift/openmos/internal/config"
	"airshift/openmos/internal/events"
	"airshift/openmos/internal/xml"
	"airshift/openmos/pkg/logger"

	"github.com/getsentry/sentry-go"
)

// ClientConnection represents a connected client
type ClientConnection struct {
	conn       net.Conn
	id         string
	server     *TCPServer // Forward declaration - TCPServer is defined in server.go
	heartbeat  *xml.HeartbeatMonitor
	parser     *xml.MessageParser
	closeChan  chan struct{}
	closeOnce  sync.Once
	writeMutex sync.Mutex
	config     *config.Config
}

// NewClientConnection creates a new client connection
func NewClientConnection(conn net.Conn, server *TCPServer, cfg *config.Config) *ClientConnection {
	clientID := conn.RemoteAddr().String()

	client := &ClientConnection{
		conn:      conn,
		id:        clientID,
		server:    server,
		parser:    xml.NewMessageParser(),
		closeChan: make(chan struct{}),
		config:    cfg,
	}

	// Create heartbeat monitor
	client.heartbeat = xml.NewHeartbeatMonitor(
		cfg.MOS.ID,
		clientID,
		cfg.MOS.ClientTimeout,
		cfg.MOS.HeartbeatInterval/2,
		client.Close,
	)

	return client
}

// Start starts processing for this client connection
func (c *ClientConnection) Start(ctx context.Context) {
	defer c.Close()

	// Start heartbeat monitoring
	monitorCtx, cancelMonitor := context.WithCancel(ctx)
	defer cancelMonitor()
	go c.heartbeat.Start(monitorCtx)

	// Subscribe to relevant events if event bus is available
	if c.server.eventBus != nil {
		roEvents := c.server.eventBus.Subscribe(events.RunningOrderUpdated, 10)
		// Start returns on every close path (EOF, read error, Close via
		// heartbeat timeout, context cancel), so this releases the subscription.
		defer c.server.eventBus.Unsubscribe(events.RunningOrderUpdated, roEvents)

		go c.forwardEvents(ctx, roEvents)
	}

	// Create a Sentry span for this client connection
	span := sentry.StartSpan(ctx, "client_connection")
	span.SetTag("client_id", c.id)
	span.SetTag("remote_addr", c.conn.RemoteAddr().String())
	defer span.Finish()

	// Read loop
	buffer := make([]byte, 4096)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closeChan:
			return
		default:
			// Set read deadline
			err := c.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			if err != nil {
				c.trackError(err, "set_read_deadline", nil)
				return
			}

			n, err := c.conn.Read(buffer)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					// Just a timeout, continue
					continue
				}
				if err == io.EOF {
					// Client closed connection
					logger.Infof("Client %s closed connection", c.id)
					return
				}
				c.trackError(err, "read", nil)
				return
			}

			// Process the data
			if n > 0 {
				c.processData(ctx, buffer[:n])
			}
		}
	}
}
