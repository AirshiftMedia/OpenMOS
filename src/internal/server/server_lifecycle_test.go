package server

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"airshift/openmos/internal/config"
)

// shutdownTimeout is far above the expected shutdown time, so a test that
// only finishes by hitting it fails the "well under the timeout" check.
const shutdownTimeout = 10 * time.Second

func newListeningServer(t *testing.T) *TCPServer {
	t.Helper()
	cfg := &config.Config{}
	cfg.Server.Host, cfg.Server.Port = "127.0.0.1", 0
	cfg.Server.ShutdownTimeout = shutdownTimeout
	cfg.Server.WriteTimeout = time.Second
	cfg.MOS.ID = "test.mos"
	cfg.MOS.HeartbeatInterval, cfg.MOS.ClientTimeout = time.Minute, time.Minute

	srv, err := NewTCPServer(cfg, nil, nil)
	if err != nil {
		t.Fatalf("NewTCPServer: %v", err)
	}
	return srv
}

func (s *TCPServer) clientCount() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

func waitForClients(t *testing.T, srv *TCPServer, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for srv.clientCount() != want {
		if time.Now().After(deadline) {
			t.Fatalf("client count = %d, want %d", srv.clientCount(), want)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Regression test for MOS-30: cancelling the context must shut the server
// down promptly, with or without connected clients, and close every client.
func TestStartShutsDownPromptlyOnCancel(t *testing.T) {
	for _, n := range []int{0, 1, 3} {
		srv := newListeningServer(t)
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan error, 1)
		go func() { result <- srv.Start(ctx) }()

		conns := make([]net.Conn, n)
		for i := range conns {
			c, err := net.Dial("tcp", srv.listener.Addr().String())
			if err != nil {
				t.Fatalf("dial: %v", err)
			}
			defer c.Close()
			conns[i] = c
		}
		waitForClients(t, srv, n)

		start := time.Now()
		cancel()
		select {
		case err := <-result:
			if err != nil {
				t.Errorf("%d clients: Start returned %v, want nil", n, err)
			}
		case <-time.After(shutdownTimeout + 2*time.Second):
			t.Fatalf("%d clients: Start did not return (deadlock)", n)
		}
		if elapsed := time.Since(start); elapsed > 3*time.Second {
			t.Errorf("%d clients: shutdown took %v, want well under %v", n, elapsed, shutdownTimeout)
		}

		if got := srv.clientCount(); got != 0 {
			t.Errorf("%d clients: %d still registered after shutdown", n, got)
		}
		// The server must have closed each connection
		for i, c := range conns {
			c.SetReadDeadline(time.Now().Add(time.Second))
			if _, err := c.Read(make([]byte, 1)); err != io.EOF {
				t.Errorf("%d clients: conn %d read = %v, want EOF", n, i, err)
			}
		}
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	srv := newListeningServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- srv.Start(ctx) }()

	cancel()
	select {
	case <-result:
	case <-time.After(shutdownTimeout + 2*time.Second):
		t.Fatal("Start did not return")
	}

	// A second Shutdown (e.g. from main) must not panic on the closed channel
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Errorf("second Shutdown = %v, want nil", err)
	}
}
