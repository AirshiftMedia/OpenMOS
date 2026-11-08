package server

import (
	"context"
	"errors"
	"net"
	"time"

	"airshift/openmos/pkg/logger"
)

// acceptLoop accepts client connections until ctx is cancelled or the
// server shuts down.
func (s *TCPServer) acceptLoop(ctx context.Context) {
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
				if errors.Is(err, net.ErrClosed) {
					// Listener closed by Shutdown
					return
				}
				logger.Errorf("Error accepting connection: %v", err)
				continue
			}

			// Create new client connection
			client := NewClientConnection(conn, s, s.config)

			// Register client
			s.registerClient(client)

			// Handle client in a goroutine. The accept loop's own count keeps
			// the wait group above zero here, so Add cannot race with Wait.
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				client.Start(ctx)
			}()
		}
	}
}
